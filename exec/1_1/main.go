// TASK 1.1
// Download the links that appear in the “On this day...” section of the
// English version of Wikipedia: https://en.wikipedia.org/wiki/Main_Page
package main

import (
	"flag"
	"fmt"
	"go.roman.zone/crawl/downloader"
	"go.roman.zone/crawl/parser"
	"log"
	"net/url"
	"os"
)

var (
	location = flag.String("location", "http://example.com", "URL of the page to retrieve")
)

const (
	OUTPUT_DIR = "./out"
	FILE_MODE  = 0777
)

func main() {
	flag.Parse()
	u, err := url.Parse(*location)
	check(err)

	fmt.Printf("Retrieving all URLs from %s...", u.String())
	urls, err := parser.GetAllURLs(*u)
	check(err)
	fmt.Print(" Done!\n")

	outDir := fmt.Sprintf("%s/%s", OUTPUT_DIR, u.Host)
	err = os.MkdirAll(outDir, FILE_MODE)
	check(err)
	downloader.DownloadURLs(urls, outDir)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
