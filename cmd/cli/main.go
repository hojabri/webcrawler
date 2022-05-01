package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"time"
	"webcrawler/crawler"
	"webcrawler/queue"

	"github.com/rs/zerolog"
)

func main() {
	// log setup
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log := zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
	// pretty log - should be disabled in production
	log = log.Output(zerolog.NewConsoleWriter())

	// setup crawler
	c := crawler.NewCrawler(&log)
	rawURL := "https://tomblomfield.com/"
	seedURL, err := url.Parse(rawURL)
	if err != nil {
		log.Fatal().Err(err).Str("url", rawURL).Msg("failed to parse seed URL")
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-signals
		log.Info().Msg("shutting down gracefully")
		cancel()
	}()
	t1 := time.Now()
	defer func() {
		log.Info().Str("time lapsed", time.Since(t1).String()).Int("unique page visited", len(c.VisitedList.Map())).Msg("crawling finished")
	}()
	go c.Crawl(ctx, seedURL, 0, 1)

	// read cloud pages
	ticker := time.NewTicker(100 * time.Millisecond)
	total := 0
	for {
		select {
		case <-c.Done:
			fmt.Printf("total: %d\n", total)
			ticker.Stop()
			return
		case <-ticker.C:
			total = total + readPrintCrawledPages(&c.Crawled)
			fmt.Printf("total links outer: %d\n", total)
		}
	}
}

func readPrintCrawledPages(c *queue.Queue) int {
	total := 0
	for {
		item := c.Pop()
		if item == nil {
			break
		}
		page, ok := item.(crawler.Page)
		if ok {
			fmt.Printf("page: %s\n", page.URL.String())
			for link := range page.Links {
				fmt.Printf("links: %s\n", link)
				total++
			}
			for link := range page.Statics {
				fmt.Printf("static: %s\n", link)
				total++
			}
		}
	}
	fmt.Printf("total links: %d\n", total)
	return total
}
