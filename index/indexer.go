package index

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

const (
	KEYWORD_EXCLUDE_REGEX = "[^\\w]"

	STORAGE_FILE = "index.csv"
)

var (
	Index = NewIndex(STORAGE_FILE)
)

type Page struct {
	URL     url.URL
	Content string
}

func ProcessPage(page Page) {
	fmt.Println("Indexed page:", page.URL.String())

	words := strings.Fields(page.Content)
	for _, w := range words {
		Index.AddItem(prepKeyword(w), IndexItem{
			URL: page.URL,
		})
	}
}

func prepKeyword(keyword string) string {
	keyword = strings.ToLower(keyword)
	re := regexp.MustCompile(KEYWORD_EXCLUDE_REGEX)
	return re.ReplaceAllString(keyword, "")
}
