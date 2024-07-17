# Chew

[![Go Report Card](https://goreportcard.com/badge/github.com/mmatongo/chew)](https://goreportcard.com/report/github.com/mmatongo/chew)
[![GoDoc](https://godoc.org/github.com/mmatongo/chew?status.svg)](https://pkg.go.dev/github.com/mmatongo/chew)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)

Chew is a Go library for processing various content types into markdown/plaintext.

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

## Features

- Supports multiple content types: HTML, PDF, CSV, JSON, YAML, DOCX, and Markdown
- Concurrent processing of multiple URLs
- Context-aware

### Similar Projects
[docconv](https://github.com/sajari/docconv)

### License
[MIT](./LICENSE)

### Roadmap
[TODO](./TODO.md)
