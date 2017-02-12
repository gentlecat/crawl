// TASK 2
// * Collect up to 1000 URL from one seed of your choice, following links.
// * Download just the first 10 webpages.
package main

import (
	"flag"
	"fmt"
	"go.roman.zone/crawl/crawler"
	"go.roman.zone/crawl/downloader"
	"log"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
)

var (
	seedURL       = flag.String("seed", "https://www.google.com/?q=test", "URL of the page to retrieve")
	targetCount   = flag.Int("index-target", 1000, "Number of unique pages to index")
	downloadLimit = flag.Int("download-limit", 10, "Maximum number of pages to download")
)

const (
	OUTPUT_DIR = "./out"
	FILE_MODE  = 0777
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

	outDir := fmt.Sprintf("%s/pages", OUTPUT_DIR)
	err = os.MkdirAll(outDir, FILE_MODE)
	check(err)
	dumpURLs(urls, fmt.Sprintf("%s/urls.txt", OUTPUT_DIR))

	if len(urls) < *downloadLimit {
		downloader.DownloadURLs(urls, outDir)
	} else {
		downloader.DownloadURLs(urls[:*downloadLimit], outDir)
	}
}

func dumpURLs(urls []url.URL, fileLocation string) {
	fmt.Printf("Writing list of indexed URLs into %s...", fileLocation)
	f, err := os.Create(fileLocation)
	check(err)
	defer f.Close()
	for _, u := range urls {
		_, err := f.WriteString(u.String() + "\n")
		check(err)
	}
	fmt.Print(" Done!\n")
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
