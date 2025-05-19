package main

import (
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func getURLsFromHTML(htmlBody, rawBaseURL string) ([]string, error) {
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return []string{}, err
	}

	var urls []string
	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					url, err := url.Parse(a.Val)
					if err != nil {
						break
					}

					if url.IsAbs() {
						urls = append(urls, a.Val)
					} else {
						urls = append(urls, rawBaseURL + a.Val)
					}
					break
				}
			}
		}
	}

	return urls, nil;
}
