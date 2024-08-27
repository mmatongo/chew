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

	// The context is optional
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	chunks, err := chew.Process(urls, ctx)
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
