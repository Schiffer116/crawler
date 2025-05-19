package main

import (
	"net/url"
	"strings"
)

func normalizeURL(raw string) (string, error) {
	url, err := url.Parse(raw);
	if err != nil {
		return "", err;
	}
	return url.Host + "/"+ strings.Trim(url.Path, "/") + url.RawQuery, nil;
}
