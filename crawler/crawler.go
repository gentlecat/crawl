package crawler

import (
	"bytes"
	"fmt"
	"go.roman.zone/crawl/crawler/classifier"
	"go.roman.zone/crawl/crawler/html_cleaner"
	"go.roman.zone/crawl/crawler/parser"
	"go.roman.zone/crawl/index"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	WORKER_COUNT          = 40
	WORKER_SLEEP_TIME_SEC = 4 // seconds
)

var (
	crawledPages   = make(map[url.URL]bool)
	retrievedPages = make(map[url.URL]bool)
	crawlMapLock   sync.Mutex

	crawlQueue = newCrawlQueue()

	crawlCountTotal = 0
	duplicateCount  = 0
	ignoredCount    = 0
	countLock       sync.Mutex
)

func Crawl(seedPage url.URL, topicKeywords []string, targetCount int, timeLimit time.Duration) []url.URL {
	wg := new(sync.WaitGroup)

	for i := 0; i <= WORKER_COUNT; i++ {
		go func(id int) {
			wg.Add(1)
			for {
				nextURL, err := crawlQueue.Pop()
				if err != nil {
					// TODO: Print status outside of this function. Perhaps some other one that just runs periodically.
					log.Printf("Worker %d: Queue is empty. Sleeping for %d seconds", id, WORKER_SLEEP_TIME_SEC)
					time.Sleep(time.Second * WORKER_SLEEP_TIME_SEC)
					continue
				}
				crawlMapLock.Lock()
				done := len(retrievedPages) >= targetCount
				crawlMapLock.Unlock()
				if done {
					wg.Done()
					return
				}
				crawlPage(nextURL, id, topicKeywords)
			}
			wg.Done()
		}(i + 1)
	}

	// Starting the process...
	crawlQueue.Push(seedPage)

	if timeLimit > 0 {
		time.Sleep(timeLimit)
		fmt.Println("Time limit has been reached. Stopping crawling.")
	} else {
		wg.Wait()
		fmt.Println("No more pages to crawl. Either limit has been reached or crawl queue is empty.")
	}

	index.Index.Export(index.STORAGE_FILE)
	return getRetrievedURLs()
}

func getRetrievedURLs() []url.URL {
	crawlMapLock.Lock()
	defer crawlMapLock.Unlock()
	crawledPageURLs := make([]url.URL, len(retrievedPages))
	i := 0
	for k := range retrievedPages {
		crawledPageURLs[i] = k
		i++
	}
	return crawledPageURLs
}

// TODO: Allow to pass a function for processing the pages. In the case of the final
// project we need to pass a page for topic checking and indexing (done separately).
func crawlPage(pageURL url.URL, workerID int, topicKeywords []string) {
	countLock.Lock()
	if crawlCountTotal%100 == 0 {
		crawlMapLock.Lock()
		log.Printf("Crawling item #%d (%d retrieved, %d ignored, %d duplicates skipped)",
			crawlCountTotal, len(retrievedPages), ignoredCount, duplicateCount)
		crawlMapLock.Unlock()
	}
	crawlCountTotal++
	countLock.Unlock()

	if isCrawled(pageURL) {
		return
	} else {
		crawlMapLock.Lock()
		crawledPages[pageURL] = true
		crawlMapLock.Unlock()
	}

	// Checking their robots.txt file. If error occurs then it's probably
	// fine to crawl anyway. 🤷
	shouldCrawl, err := ShouldCrawl(pageURL)
	if err != nil {
		log.Printf("Worker %d: Failed to get robots.txt from %s: %s\n",
			workerID, pageURL.Host, err)
	}
	if !shouldCrawl {
		countLock.Lock()
		ignoredCount++
		countLock.Unlock()
		return
	}

	// Retrieving the page, parsing, etc.
	pageContent, err := GetPage(pageURL)
	crawlMapLock.Lock()
	retrievedPages[pageURL] = true
	crawlMapLock.Unlock()
	if err != nil {
		log.Printf("Worker %d: Failed to crawl page %s: %s\n",
			workerID, pageURL.String(), err)
		return
	}
	linksToQueue(pageContent) // extracting links before indexing to not slow down the process
	if classifier.IsTopical(html_cleaner.Clean(pageContent), topicKeywords) {
		index.ProcessPage(index.Page{URL: pageURL, Content: pageContent})
	}
}

// linksToQueue does link extraction from an HTML page and puts all uncrawled
// URLs into the crawl queue.
func linksToQueue(pageContent string) {
	urls, err := parser.GetAllURLs(pageContent)
	if err != nil {
		log.Printf("Failed to extract links: %s\n", err)
		return
	}
	for _, u := range urls {
		if !isCrawled(u) {
			crawlQueue.Push(u)
		}
	}
}

func isCrawled(pageURL url.URL) bool {
	crawlMapLock.Lock()
	isCrawled, found := crawledPages[pageURL]
	crawlMapLock.Unlock()
	if found && isCrawled {
		countLock.Lock()
		duplicateCount++
		countLock.Unlock()
	}
	return found && isCrawled
}

func GetPage(pageURL url.URL) (string, error) {
	resp, err := http.Get(pageURL.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	return buf.String(), nil
}
