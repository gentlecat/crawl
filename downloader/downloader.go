package downloader

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

const (
	FILE_MODE = 0777
)

func DownloadURLs(urls []url.URL, location string) {
	fmt.Printf("Writing documents into %s directory...", location)
	for i, u := range urls {
		content, err := getHTML(u)
		if err != nil {
			continue
		}
		err = ioutil.WriteFile(fmt.Sprintf("%s/%d.html", location, i), content, FILE_MODE)
		check(err)
	}
	fmt.Print(" Done!\n")
}

func getHTML(url url.URL) ([]byte, error) {
	timeout := time.Duration(5 * time.Second)
	httpClient := http.Client{
		Timeout: timeout,
	}
	resp, err := httpClient.Get(url.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
