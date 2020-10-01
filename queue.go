package scrapy

import (
	heap2 "container/heap"
	"sync"
)

type requestHeap []*Request

type RequestQueue struct {
	mux         sync.Mutex
	requestHeap requestHeap
}

func NewQueue(cap int) *RequestQueue {
	heap := make([]*Request, 0, cap)

	rHeap := &RequestQueue{
		requestHeap: heap,
		mux:         sync.Mutex{},
	}

	heap2.Init(&rHeap.requestHeap)
	return rHeap
}

func (r *RequestQueue) PushRequest(req *Request) {
	r.mux.Lock()
	defer r.mux.Unlock()
	heap2.Push(&r.requestHeap, req)
}

func (r *RequestQueue) PopRequest() *Request {
	r.mux.Lock()
	defer r.mux.Unlock()

	item := heap2.Pop(&r.requestHeap)
	if item == nil {
		return nil
	}

	return item.(*Request)
}

func (r *RequestQueue) Len() int {
	return r.requestHeap.Len()
}

func (r requestHeap) Len() int {
	return len(r)
}

func (r requestHeap) Less(i, j int) bool {
	return r[i].priority < r[j].priority
}

func (r requestHeap) Swap(i, j int) {
	if i > 0 && j > 0 {
		r[i], r[j] = r[j], r[i]
	}
}

func (r *requestHeap) Push(x interface{}) {
	item := x.(*Request)
	*r = append(*r, item)
}

func (r *requestHeap) Pop() interface{} {
	old := *r
	n := len(*r)
	if n == 0 {
		return nil
	}

	item := old[n-1]
	*r = old[0 : n-1]
	return item
}
