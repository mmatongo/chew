Chew is a Go library for processing various content types into markdown.

## Features

- Supports multiple content types: HTML, PDF, CSV, JSON, YAML, DOCX, and Markdown
- Concurrent processing of multiple URLs
- Context-aware

## Installation

```bash
go get github.com/mmatongo/chew
```

## Usage

Here's a basic example of how to use Chew:

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

### Output

```bash
Source: https://example.com
Content: Example Domain

Source: https://example.com
Content: This domain is for use in illustrative examples in documents. You may use this domain in literature without prior coordination or asking for permission.

Source: https://example.com
Content: More information...
```

### TODO

- [ ] Add tests
- [ ] Improve error handling
- [ ] Add support for more content types
- [ ] Implement rate limiting for URL fetching
- [x] Use a free PDF processing library
- [ ] How to handle text/plain content type


## License
[MIT](./LICENSE)
