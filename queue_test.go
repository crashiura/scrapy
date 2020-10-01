package scrapy

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestQueue(t *testing.T) {
	reqs := []*Request{
		{
			priority: 1,
		},
		{
			priority: 1,
		},
		{
			priority: 1,
		},
		{
			priority: 0,
		},
	}

	q := NewQueue(len(reqs))

	go func() {
		for i := 0; i < 10000; i++ {
			req := &Request{
				priority: i,
			}
			q.PushRequest(req)
		}
	}()

	for _, r := range reqs {
		q.PushRequest(r)
	}

	f := func() {
		for {
			if q.Len() > 0 {
				time.Sleep(time.Microsecond * 50)
				req := q.PopRequest()
				if req != nil {
					fmt.Println(req.priority)
				}
			}
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	for i := 0; i < 10; i++ {
		go f()
	}
	wg.Wait()
}
