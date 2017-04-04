package crawler

import (
	"github.com/temoto/robotstxt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var (
	urlCache      = make(map[url.URL]bool)
	urlCacheMutex sync.Mutex

	robotsDataCache  = make(map[string]*robotstxt.RobotsData)
	robotsCacheMutex sync.Mutex
)

const (
	USER_AGENT             = "Googlebot"
	ROBOTS_REQUEST_TIMEOUT = 2 // seconds
	ROBOTS_PATH            = "/robots.txt"
)

func ShouldCrawl(url url.URL) (bool, error) {
	urlCacheMutex.Lock()
	if isAllowed, ok := urlCache[url]; ok {
		urlCacheMutex.Unlock()
		return isAllowed, nil
	}
	urlCacheMutex.Unlock()
	robotsData, err := GetRobotsData(url.Host)
	if err != nil {
		return true, err
	}
	group := robotsData.FindGroup(USER_AGENT)
	isAllowed := group.Test(url.Path)
	urlCacheMutex.Lock()
	urlCache[url] = isAllowed
	urlCacheMutex.Unlock()
	return isAllowed, nil
}

func GetRobotsData(host string) (*robotstxt.RobotsData, error) {
	robotsCacheMutex.Lock()
	if data, ok := robotsDataCache[host]; ok {
		robotsCacheMutex.Unlock()
		return data, nil
	}
	robotsCacheMutex.Unlock()
	robotsURL := url.URL{
		Scheme: "https",
		Host:   host,
		Path:   ROBOTS_PATH,
	}
	var resp *http.Response
	noData, err := robotstxt.FromString("")
	if err != nil {
		log.Fatal(err)
	}
	httpClient := http.Client{
		Timeout: time.Duration(ROBOTS_REQUEST_TIMEOUT * time.Second),
	}
	resp, err = httpClient.Get(robotsURL.String())
	if err != nil {
		// Retrying with HTTP
		robotsURL.Scheme = "http"
		resp, err = httpClient.Get(robotsURL.String())
		if err != nil {
			robotsCacheMutex.Lock()
			robotsDataCache[host] = noData
			robotsCacheMutex.Unlock()
			return noData, err
		}
	}
	defer resp.Body.Close()

	data, err := robotstxt.FromResponse(resp)
	if err != nil {
		return noData, err
	}
	robotsCacheMutex.Lock()
	robotsDataCache[host] = data
	robotsCacheMutex.Unlock()
	return data, nil
}
