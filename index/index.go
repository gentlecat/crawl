package index

import (
	"encoding/csv"
	"errors"
	"log"
	"net/url"
	"os"
	"sync"
)

var (
	ErrMissingIndex   = errors.New("can't find the index file")
	ErrIndexFileIsDir = errors.New("index file is a directory")
)

// Since there's no Queue type in Go, we can just use this. Kind of hacky, but
// that's ok. See https://github.com/golang/go/wiki/SliceTricks for more info.
type indexType struct {
	// Map of keywords to items
	mapping map[string][]IndexItem
	mutex   sync.Mutex
}

type IndexItem struct {
	// FIXME: There might a problem with this:
	// Same page might appear multiple times under one keyword. Though, this might
	// be beneficial for determining how often a word appears...
	URL url.URL
}

func NewIndex(filename string) *indexType {
	index, err := importFromFile(filename)
	if err != nil {
		if err == ErrMissingIndex {
			return &indexType{
				mapping: make(map[string][]IndexItem, 0),
			}
		} else {
			log.Fatal(err)
		}
	}
	return index
}

func importFromFile(filename string) (*indexType, error) {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrMissingIndex
		} else {
			return nil, err
		}
	}
	if fileInfo.IsDir() {
		return nil, ErrIndexFileIsDir
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	csvReader := csv.NewReader(f)
	csvReader.FieldsPerRecord = -1 // We have variable number of fields, so no need to do checking
	lines, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	indexMapping := make(map[string][]IndexItem, 0)

	for _, line := range lines {
		if len(line) < 1 {
			continue
		}
		keyword := line[0]
		indexMapping[keyword] = make([]IndexItem, 0)
		for _, urlStr := range line[1:] {
			parsedURL, err := url.Parse(urlStr)
			if err != nil {
				return nil, err
			}
			indexMapping[keyword] = append(indexMapping[keyword], IndexItem{URL: *parsedURL})
		}
	}

	return &indexType{
		mapping: indexMapping,
	}, nil
}

func (i *indexType) AddItem(keyword string, item IndexItem) {
	i.mutex.Lock()
	if _, ok := i.mapping[keyword]; ok {
		i.mapping[keyword] = append(i.mapping[keyword], item)
	} else {
		i.mapping[keyword] = make([]IndexItem, 0)
		i.mapping[keyword] = append(i.mapping[keyword], item)
	}
	i.mutex.Unlock()
}

func (i *indexType) GetItem(keyword string) []IndexItem {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	if _, ok := i.mapping[keyword]; ok {
		return i.mapping[keyword]
	} else {
		return make([]IndexItem, 0)
	}
}

func (i *indexType) Length() int {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	return len(i.mapping)
}

func (i *indexType) Export(filename string) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	for keyword, items := range i.mapping {
		row := make([]string, 0)
		row = append(row, keyword)
		for _, item := range items {
			row = append(row, item.URL.String())
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	return nil
}
