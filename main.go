package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"slices"
	"sort"
	"strconv"
	"sync"
)

type config struct {
	pages              map[string]int
	baseURL            *url.URL
	mu                 *sync.Mutex
	concurrencyControl chan struct{}
	wg                 *sync.WaitGroup
	maxPages		   int
}

func main() {
	if len(os.Args) != 4 {
		fmt.Println("usage: ./crawler URL maxConcurrency maxPages")
		os.Exit(1)
	}

	baseURL, err := url.Parse(os.Args[1])
	if err != nil {
		fmt.Println("Invalid URL")
		os.Exit(1)
	}

	maxConcurrency, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Invalid max concurrency")
		os.Exit(1)
	}

	maxPages, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Println("Invalid max page number")
		os.Exit(1)
	}

	cfg := config {
		pages: make(map[string]int),
		baseURL: baseURL,
		mu: &sync.Mutex{},
		concurrencyControl: make(chan struct{}, maxConcurrency),
		wg: &sync.WaitGroup{},
		maxPages: maxPages,
	}

	cfg.wg.Add(1)
	go cfg.crawlPage(os.Args[1])
	cfg.wg.Wait()

	printReport(cfg.pages, os.Args[1])
}

func getHTML(rawURL string) (string, error) {
	resp, err := http.Get(rawURL)
	if err != nil {
		fmt.Println("Error connecting to server")
		os.Exit(1)
	}

	if sc := resp.StatusCode; 400 <= sc && sc < 500 {
		return "", fmt.Errorf("Error status code: %v", sc)
	}

	contentType := resp.Header["content-type"]
	if contentType != nil && !slices.Contains(contentType, "text/html") {
		return "", fmt.Errorf("Error invalid content type")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (cfg *config) crawlPage(rawCurrentURL string) {
	cfg.concurrencyControl <- struct{}{}
	defer func() {
		cfg.wg.Done()
		<-cfg.concurrencyControl
	}()

	cfg.mu.Lock()
	if len(cfg.pages) >= cfg.maxPages {
		cfg.mu.Unlock()
		return
	}
	cfg.mu.Unlock()

	currentURL, err := url.Parse(rawCurrentURL)
	if err != nil {
		fmt.Printf("Error parsing %s\n", rawCurrentURL)
		return
	}

	if cfg.baseURL.Host != currentURL.Host {
		return
	}

	normalized, _ := normalizeURL(rawCurrentURL)
	isFirst := cfg.addPageVisit(normalized)
	if !isFirst {
		return
	}

	html, err := getHTML(rawCurrentURL)
	if err != nil {
		return
	}

	urls, err := getURLsFromHTML(html, cfg.baseURL.ResolveReference(&url.URL{Path: ""}).String())
	if err != nil {
		return
	}

	for _, url := range urls {
		cfg.wg.Add(1)
		go cfg.crawlPage(url)
	}
}

func (cfg *config) addPageVisit(normalizedURL string) (isFirst bool) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	cfg.pages[normalizedURL] += 1
	if cfg.pages[normalizedURL] == 1 {
		isFirst = true
		return
	}
	isFirst = false
	return
}

func printReport(pages map[string]int, baseURL string) {
	var urls []string
	for url := range pages {
		urls = append(urls, url)
	}

	sort.Strings(urls)

	fmt.Printf(`
=============================
REPORT for %s
=============================
`, baseURL)

	for _, url := range urls {
		fmt.Printf("Found %v internal links to %s\n", pages[url], url)
	}
}
