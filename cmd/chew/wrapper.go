package main

import (
	"C"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mmatongo/chew"
)

//export Process
func Process(urls *C.char) *C.char {
	urlsSlice := strings.Split(C.GoString(urls), ",")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	chunks, err := chew.Process(urlsSlice, ctx)
	if err != nil {
		if err == context.DeadlineExceeded {
			return C.CString("Operation timed out")
		}
		return C.CString(fmt.Sprintf("Error processing URLs: %v", err))
	}

	var result strings.Builder
	for _, chunk := range chunks {
		result.WriteString(fmt.Sprintf("Source: %s\nContent: %s\n\n", chunk.Source, chunk.Content))
	}

	return C.CString(result.String())
}

func main() {}
