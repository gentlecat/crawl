package parser

import (
	"bytes"
	"errors"
	"golang.org/x/net/html"
	"net/url"
	"strings"
)

// GetAllURLs retrieves all URLs from an HTML page.
func GetAllURLs(pageContent string) ([]url.URL, error) {
	var urls []url.URL

	tokenizer := html.NewTokenizer(bytes.NewReader([]byte(pageContent)))
	for {
		tt := tokenizer.Next()
		switch {
		case tt == html.ErrorToken:
			return urls, nil
		case tt == html.StartTagToken:
			t := tokenizer.Token()
			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}
			link, err := extractLink(t)
			if err != nil {
				continue
			}
			u, err := url.ParseRequestURI(link)
			if err != nil {
				continue
			}
			if !(strings.EqualFold(u.Scheme, "HTTPS") || strings.EqualFold(u.Scheme, "HTTP")) {
				continue
			}
			urls = append(urls, *u)
		}
	}
	return urls, nil
}

func extractLink(t html.Token) (string, error) {
	for _, a := range t.Attr {
		if a.Key == "href" {
			return a.Val, nil
		}
	}
	return "", errors.New("Link not found")
}
