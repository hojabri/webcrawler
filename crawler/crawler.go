package crawler

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"webcrawler/queue"
	"webcrawler/visited"

	"golang.org/x/net/html"

	"github.com/rs/zerolog"
)

type Page struct {
	URL     *url.URL
	Links   map[string]string
	Statics map[string]string
}

type Crawler struct {
	log            *zerolog.Logger
	mu             sync.Mutex
	wg             sync.WaitGroup
	VisitedList    visited.Visited
	waitingToCrawl queue.Queue
	processing     int64
	Crawled        queue.Queue
	Done           chan struct{}
	isServerDown   bool
}

func NewCrawler(log *zerolog.Logger) Crawler {
	return Crawler{
		log:            log,
		VisitedList:    visited.NewVisitedList(),
		waitingToCrawl: queue.NewQueue(),
		Crawled:        queue.NewQueue(),
		Done:           make(chan struct{}),
	}
}

func (c *Crawler) Run(ctx context.Context, seedURL *url.URL, depth int, parallel int) {
	go c.Crawl(ctx, seedURL, 0, 1)
}

func (c *Crawler) Crawl(ctx context.Context, seedURL *url.URL, depth int, parallel int) {
	if parallel == 0 {
		parallel = 1
	}
	jobs := make(chan url.URL)

	for i := 0; i < parallel; i++ {
		c.wg.Add(1)
		go c.worker(ctx, jobs)
	}
	c.VisitedList.Store(seedURL.String(), struct{}{})
	c.waitingToCrawl.Add(*seedURL)

	for {
		if c.isServerDown {
			return
		}
		item := c.waitingToCrawl.Pop()
		if item != nil {
			if pickedURL, ok := item.(url.URL); ok {
				atomic.AddInt64(&c.processing, 1)
				// go c.crawl(&pickedURL)
				jobs <- pickedURL
			}
		} else if c.processing < 1 { // queue is empty
			c.shutdown()
		}
	}
	c.wg.Wait()
}

func (c *Crawler) shutdown() {
	if !c.isServerDown {
		c.mu.Lock()
		c.Done <- struct{}{}
		close(c.Done)
		c.isServerDown = true
		c.mu.Unlock()
	}
}

func (c *Crawler) worker(ctx context.Context, jobs <-chan url.URL) {
	defer c.wg.Done()
	for {
		select {
		case <-ctx.Done():
			c.log.Err(ctx.Err()).Msg("cancelled worker")
			c.shutdown()
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}
			c.crawl(&job)
		}
	}
}

func (c *Crawler) crawl(seedURL *url.URL) {
	defer atomic.AddInt64(&c.processing, -1)

	page := Page{
		URL:     seedURL,
		Links:   make(map[string]string),
		Statics: make(map[string]string),
	}

	err := c.fetch(&page)
	if err != nil {
		c.log.Err(err).Str("url", seedURL.String()).Msg("failed to fetch page contents for url")
	}
}

func (c *Crawler) fetch(page *Page) error {
	req, err := http.NewRequest("GET", page.URL.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "web-crawler")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = c.parse(page, resp.Body)
	if err != nil {
		return err
	}
	c.Crawled.Add(*page)
	return nil
}

func (c *Crawler) parse(page *Page, body io.ReadCloser) error {
	tokens := html.NewTokenizer(body)
	for {
		tokenType := tokens.Next()

		switch tokenType {
		case html.ErrorToken:
			return nil
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokens.Token()
			var relativeURL string
			switch token.DataAtom.String() {
			case "a", "link": // link tags
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						relativeURL = attr.Val
						_, ok := page.Links[relativeURL]
						if !ok {
							absURL, err := absoluteURL(page.URL, relativeURL)
							if err != nil {
								return err
							}
							if absURL != nil && absURL.String() != "" {

								page.Links[relativeURL] = absURL.String()
								if !c.VisitedList.Exist(absURL.String()) {
									c.VisitedList.Store(absURL.String(), struct{}{})
									c.waitingToCrawl.Add(*absURL)
								}

							}
						}
					}
				}
			case "img", "image", "script": // static tags
				for _, attr := range token.Attr {
					if attr.Key == "src" {
						relativeURL = attr.Val
						_, ok := page.Links[relativeURL]
						if !ok {
							absURL, err := absoluteURL(page.URL, relativeURL)
							if err != nil {
								return err
							}
							if absURL != nil && absURL.String() != "" {
								page.Statics[relativeURL] = absURL.String()
							}

						}
					}
				}
			}

		}

	}
}

func absoluteURL(u *url.URL, relativeURL string) (*url.URL, error) {
	relURL, err := url.Parse(relativeURL)
	if err != nil {
		return nil, err
	}

	newURL := u.ResolveReference(relURL)
	if newURL.Host != u.Host { // skip external URLs
		return nil, nil
	}

	return newURL, nil
}
