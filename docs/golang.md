Chew is native to Go and can be used as library in your Go project with ease. Here is a simple example of how to use Chew in your Go project.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mmatongo/chew"
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
```

The above code snippet demonstrates how to use Chew in your Go project. The `chew.Process` function takes a list of URLs and returns a list of `Chunk` objects. Each `Chunk` object contains the source URL and the content of the URL. The `context` parameter is optional and can be used to set a timeout for the operation. If the operation times out, the function will return a `context.DeadlineExceeded` error.

Markdown formatting is not enforced in the content of the `Chunk` object. However, the output is always going to be plain text so you can format it as you wish.
