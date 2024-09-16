package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mmatongo/chew/v1"
)

func main() {
	urls := []string{
		"https://example.com",
	}

	config := chew.Config{
		UserAgent:       "Chew/1.0 (+https://github.com/mmatongo/chew)",
		RetryLimit:      3,
		RetryDelay:      5 * time.Second,
		CrawlDelay:      10 * time.Second,
		ProxyList:       []string{}, // Add your proxies here, or leave empty
		RateLimit:       2 * time.Second,
		RateBurst:       3,
		IgnoreRobotsTxt: false,
	}

	haChew := chew.New(config)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	chunks, err := haChew.Process(ctx, urls)
	if err != nil {
		if err == context.DeadlineExceeded {
			log.Println("Operation timed out")
		} else {
			log.Printf("Error processing URLs: %v", err)
		}
		return
	}

	for _, chunk := range chunks {
		fmt.Printf("Source: %s\nContent: %s\n\n", chunk.Source, chunk.Content)
	}
}
