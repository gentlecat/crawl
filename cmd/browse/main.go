package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"go.roman.zone/crawl/index"
	"log"
	"net/http"
	"sort"
)

var (
	listenHost = flag.String("host", "127.0.0.1", "Host to listen on")
	listenPort = flag.Int("port", 8080, "Port to listen on")
)

func main() {
	idx, err := index.Import(index.STORAGE_FILE)
	check(err)
	index.Index = idx

	listenAddr := fmt.Sprintf("%s:%d", *listenHost, *listenPort)
	log.Printf("Starting server on %s...\n", listenAddr)
	check(http.ListenAndServe(listenAddr, makeRouter()))
}

func makeRouter() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)

	// Attach new handlers here:
	r.HandleFunc("/", queryHandler)

	return r
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()
	query, ok := queries["q"]
	if !ok {
		http.Error(w, "No query", http.StatusBadRequest)
		return
	}
	queryStr := query[0]

	items := index.Index.GetItem(queryStr)
	itemsRanked := make(map[index.IndexItem]int)

	for _, item := range items {
		if _, ok := itemsRanked[item]; ok {
			itemsRanked[item]++
		} else {
			itemsRanked[item] = 1
		}
	}

	results := make(SearchResults, len(itemsRanked))
	i := 0
	for k, v := range itemsRanked {
		results[i] = SearchResult{k, v}
		i++
	}
	sort.Sort(sort.Reverse(results))

	resultsOut := make([]SearchResultOutput, len(results))
	i = 0
	for _, item := range results {
		resultsOut[i] = SearchResultOutput{
			URL:  item.Item.URL.String(),
			Rank: item.Rank,
		}
		i++
	}

	b, err := json.Marshal(resultsOut)
	if err != nil {
		http.Error(w, "Internal error.", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	callback, ok := queries["callback"]
	if ok {
		w.Header().Set("Content-Type", "application/javascript")
		fmt.Fprintf(w, "%s(%s)", callback[0], b)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	}
}

type SearchResultOutput struct {
	URL  string `json:"url"`
	Rank int    `json:"rank"`
}

type SearchResult struct {
	Item index.IndexItem
	Rank int
}

type SearchResults []SearchResult

func (p SearchResults) Len() int           { return len(p) }
func (p SearchResults) Less(i, j int) bool { return p[i].Rank < p[j].Rank }
func (p SearchResults) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
