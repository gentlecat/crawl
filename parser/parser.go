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

// GetURLsInElement retrieves all URLs from a page (`pageURL`) which are
// inside a tag that has a specific ID (`elementID`).
func GetURLsInElement(pageURL url.URL, elementID string) []url.URL {
	resp, err := http.Get(pageURL.String())
	check(err)
	defer resp.Body.Close()

	var urls []url.URL

	z := html.NewTokenizer(resp.Body)
	depth := 0
	found := false
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return expandRelativeURLs(urls, pageURL)
		case html.StartTagToken, html.EndTagToken:
			t := z.Token()
			if found {
				if tt == html.StartTagToken {
					depth++
				} else {
					depth--
					if depth <= 0 {
						return expandRelativeURLs(urls, pageURL)
					}
				}
				if !(t.Data == "a") {
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
				urls = append(urls, *u)
				continue
			}
			if !(t.Data == "div") {
				continue
			}

			id, err := extractID(t)
			if err != nil {
				continue
			}
			if id == elementID && tt == html.StartTagToken {
				depth++
				found = true
			}
		}
	}
	return expandRelativeURLs(urls, pageURL)
}

func extractID(t html.Token) (string, error) {
	for _, a := range t.Attr {
		if a.Key == "id" {
			return a.Val, nil
		}
	}
	return "", errors.New("ID not found")
}
func extractLink(t html.Token) (string, error) {
	for _, a := range t.Attr {
		if a.Key == "href" {
			return a.Val, nil
		}
	}
	return "", errors.New("Link not found")
}
func expandRelativeURLs(urls []url.URL, sourcePage url.URL) []url.URL {
	expandedURLs := urls
	for i, u := range expandedURLs {
		if u.Scheme == "" {
			expandedURLs[i].Scheme = sourcePage.Scheme
		}
		if u.Opaque == "" {
			expandedURLs[i].Opaque = sourcePage.Opaque
		}
		if u.User == nil {
			expandedURLs[i].User = sourcePage.User
		}
		if u.Host == "" {
			// Some websites specify protocol-relative URLs. It became an anti-pattern, but we still
			// need to support it. See https://www.paulirish.com/2010/the-protocol-relative-url/
			// for more info about those.
			if strings.HasPrefix(u.Path, "//") {
				// Moving actual host value into its place
				expandedURLs[i].Host = strings.Split(u.Path[2:], "/")[0]
				expandedURLs[i].Path = expandedURLs[i].Path[2+len(expandedURLs[i].Host):]

			} else {
				expandedURLs[i].Host = sourcePage.Host
			}
		}
	}
	return expandedURLs
}
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
