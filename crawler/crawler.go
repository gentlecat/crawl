package crawler

import (
	"go.roman.zone/crawl/parser"
	"log"
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

func Crawl(seedPage url.URL, targetCount int) []url.URL {
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
				crawlPage(nextURL, id)

			}
			wg.Done()
		}(i + 1)
	}

	crawlQueue.Push(seedPage)

	wg.Wait()
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

func crawlPage(pageURL url.URL, workerID int) {
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
	urls, err := parser.GetAllURLs(pageURL)
	crawlMapLock.Lock()
	retrievedPages[pageURL] = true
	crawlMapLock.Unlock()
	if err != nil {
		log.Printf("Worker %d: Failed to crawl page %s: %s\n",
			workerID, pageURL.String(), err)
	}
	for _, u := range urls {
		if isCrawled(pageURL) {
			return
		}
		crawlQueue.Push(u)
	}
}

func isCrawled(pageURL url.URL) bool {
	crawlMapLock.Lock()
	isCrawled, notFound := crawledPages[pageURL]
	crawlMapLock.Unlock()
	if !(!notFound && isCrawled) {
		countLock.Lock()
		duplicateCount++
		countLock.Unlock()
	}
	return !notFound && isCrawled
}
