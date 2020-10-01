package scrapy

import (
	"context"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// ScrapyOptions...
type ScrapyOptions func(*Scrapy)
type RequestCallback func(*Request)
type ResponseCallback func(*http.Response)
type HtmlCallback func(*Request, *http.Response, *goquery.Document)
type ErrorCallback func(*http.Response, error)

type Scrapy struct {
	Client   *http.Client
	Threads  int
	Timeout  time.Duration
	Queue    *RequestQueue
	handlers []*Handler
	wg       *sync.WaitGroup
	cancel   context.CancelFunc
}

type Handler struct {
	id                  int
	onRequestCallbacks  []RequestCallback
	onResponseCallbacks []ResponseCallback
	onHtmlCallbacks     []HtmlCallback
	onErrorCallback     []ErrorCallback
	Priority            int
	scrapy              *Scrapy
}

// Request...
type Request struct {
	Request      *http.Request
	Ctx          context.Context
	priority     int
	handlerIndex int
}

// New...
func NewScrapy(opts ...ScrapyOptions) *Scrapy {
	var netTransport = &http.Transport{
		DisableKeepAlives: true,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 60 * time.Second,
	}

	client := &http.Client{Transport: netTransport}

	scrapy := &Scrapy{
		Client:  client,
		Timeout: time.Second,
		Queue:   NewQueue(100),
		Threads: 5,
	}

	for _, opt := range opts {
		opt(scrapy)
	}

	return scrapy
}

func SetTimeout(timeout time.Duration) ScrapyOptions {
	return func(s *Scrapy) {
		s.Timeout = timeout
	}
}

func (s *Scrapy) AddHandler(h *Handler) {
	id := len(s.handlers)
	h.id = id
	h.scrapy = s
	s.handlers = append(s.handlers, h)
}

func (s *Scrapy) Run() {
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.wg = wg
	s.runWorkers(ctx, wg)
}

func (s *Scrapy) Shutdown() {
	s.cancel()
}

func (s *Scrapy) Wait() {
	s.wg.Wait()
}

func (s *Scrapy) runWorkers(ctx context.Context, wg *sync.WaitGroup) {
	for i := 0; i < s.Threads; i++ {
		wg.Add(1)
		go s.work(ctx, wg)
	}
}

func (s *Scrapy) work(ctx context.Context, wg *sync.WaitGroup) {
	q := s.Queue
	client := s.Client
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if q.Len() > 0 {
				request := q.PopRequest()
				if request == nil {
					continue
				}
				handler := s.handlers[request.handlerIndex]
				for _, cb := range handler.onRequestCallbacks {
					cb(request)
				}

				RandomUserAgent(request)
				res, err := client.Do(request.Request)
				if err != nil {
					for _, cb := range handler.onErrorCallback {
						cb(res, err)
					}

					log.Println("Error in get response", err)
					continue
				}

				for _, cb := range handler.onResponseCallbacks {
					cb(res)
				}

				doc, err := goquery.NewDocumentFromReader(res.Body)
				if err != nil {
					log.Println("Error create html document ", err)
					//if err != res.Body.Close() {
					//	return err
					//}

					continue
				}

				for _, cb := range handler.onHtmlCallbacks {
					cb(request, res, doc)
				}

				res.Body.Close()
			}
		}

		time.Sleep(s.Timeout)
	}
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) AddRequest(req *Request) {
	if h.scrapy == nil {
		panic("Scrapy not found handler")
	}
	req.handlerIndex = h.id
	req.priority = h.Priority
	h.scrapy.Queue.PushRequest(req)
}

func (h *Handler) AddURL(ctx context.Context, url string) error {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req := &Request{
		Request:      request,
		Ctx:          ctx,
		priority:     h.Priority,
		handlerIndex: h.id,
	}
	h.AddRequest(req)
	return nil
}

func (h *Handler) AddURLWithContext(url string, ctx context.Context) error {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req := &Request{
		Request:      request,
		Ctx:          ctx,
		priority:     h.Priority,
		handlerIndex: h.id,
	}
	h.AddRequest(req)
	return nil
}

func (h *Handler) OnRequest(callback RequestCallback) {
	if h.onRequestCallbacks == nil {
		h.onRequestCallbacks = make([]RequestCallback, 0, 4)
	}

	h.onRequestCallbacks = append(h.onRequestCallbacks, callback)
}

func (h *Handler) OnResponse(callback ResponseCallback) {
	if h.onResponseCallbacks == nil {
		h.onResponseCallbacks = make([]ResponseCallback, 0, 4)
	}
	h.onResponseCallbacks = append(h.onResponseCallbacks, callback)
}

func (h *Handler) OnHtml(callback HtmlCallback) {
	if h.onHtmlCallbacks == nil {
		h.onHtmlCallbacks = make([]HtmlCallback, 0, 4)
	}

	h.onHtmlCallbacks = append(h.onHtmlCallbacks, callback)
}

func (h *Handler) OnError(callback ErrorCallback) {
	if h.onErrorCallback == nil {
		h.onErrorCallback = make([]ErrorCallback, 0, 4)
	}

	h.onErrorCallback = append(h.onErrorCallback, callback)
}
