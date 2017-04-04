package main

import (
	"flag"
	"fmt"
	"go.roman.zone/crawl/crawler"
	"log"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"strings"
)

var (
	seedURL     = flag.String("seed", "https://example.com", "URL of the page to use as a seed")
	keywordsStr    = flag.String("keywords", "", "Comma-separated list of keywords that define a topic")
	targetCount = flag.Int("index-target", 1000, "Number of unique pages to index")
	timeLimit   = flag.Duration("time-limit", 0, "Maximum time the crawler should run for")
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	flag.Parse()
	seedURLParsed, err := url.Parse(*seedURL)
	check(err)
	keywords := strings.Split(*keywordsStr, ",")

	fmt.Printf("Crawling using %s as a seed with keywords: %s.\n",
		seedURLParsed.String(), strings.Join(keywords, ","))
	urls := crawler.Crawl(*seedURLParsed, keywords, *targetCount, *timeLimit)
	fmt.Printf("Indexed %d pages.\n", len(urls))
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
