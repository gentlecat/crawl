package crawler

import (
	"errors"
	"net/url"
	"sync"
)

// Since there's no Queue type in Go, we can just use this. Kind of hacky, but
// that's ok. See https://github.com/golang/go/wiki/SliceTricks for more info.
type crawlQueueType struct {
	queue []url.URL
	m     sync.Mutex
}

func newCrawlQueue() *crawlQueueType {
	return &crawlQueueType{
		queue: make([]url.URL, 0),
	}
}

func (q *crawlQueueType) Push(u url.URL) {
	q.m.Lock()
	q.queue = append(q.queue, u)
	q.m.Unlock()
}

func (q *crawlQueueType) Pop() (url.URL, error) {
	q.m.Lock()
	defer q.m.Unlock()
	if len(q.queue) == 0 {
		return url.URL{}, errors.New("The queue is empty")
	}
	val := q.queue[0]
	q.queue = q.queue[1:] // Discard top element
	return val, nil
}

func (q *crawlQueueType) Length() int {
	q.m.Lock()
	defer q.m.Unlock()
	return len(q.queue)
}
