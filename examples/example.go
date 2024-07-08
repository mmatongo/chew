package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mmatongo/chew"
)

func main() {
	urls := []string{
		"https://example.com",
	}

	chunks, err := chew.Process(context.Background(), urls)
	if err != nil {
		log.Fatalf("Error processing URLs: %v", err)
	}

	for _, chunk := range chunks {
		fmt.Printf("Source: %s\nContent: %s\n\n", chunk.Source, chunk.Content)
	}
}
