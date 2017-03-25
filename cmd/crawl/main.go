package main

import (
	"flag"
	"fmt"
	"go.roman.zone/crawl/crawler"
	"log"
	"net/http"
	_ "net/http/pprof"
	"net/url"
)

var (
	seedURL     = flag.String("seed", "https://roman.zone", "URL of the page to retrieve")
	targetCount = flag.Int("index-target", 1000, "Number of unique pages to index")
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	flag.Parse()
	seedURLParsed, err := url.Parse(*seedURL)
	check(err)

	urls := crawler.Crawl(*seedURLParsed, *targetCount)
	fmt.Printf("Indexed %d pages.\n", len(urls))
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
