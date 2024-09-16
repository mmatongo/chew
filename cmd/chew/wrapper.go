package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unsafe"

	"github.com/mmatongo/chew/v1"
)

//export Process
func Process(urls *C.char) *C.char {
	urlsSlice := strings.Split(C.GoString(urls), ",")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c := chew.New(chew.Config{
		UserAgent:       "Chew/1.0 (+https://github.com/mmatongo/chew)",
		RetryLimit:      3,
		RetryDelay:      time.Second,
		CrawlDelay:      time.Second,
		RateLimit:       time.Second,
		RateBurst:       1,
		IgnoreRobotsTxt: false,
	})

	chunks, err := c.Process(ctx, urlsSlice)
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

//export FreeString
func FreeString(ptr *C.char) {
	C.free(unsafe.Pointer(ptr))
}

func main() {}
