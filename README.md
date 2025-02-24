<div align="center">
<img
    width=40%
    src="https://raw.githubusercontent.com/mmatongo/chew/master/assets/gopher-eating.svg"
    alt="chew logo"
/>

[![Go Report Card](https://goreportcard.com/badge/github.com/mmatongo/chew)](https://goreportcard.com/report/github.com/mmatongo/chew)
[![GoDoc](https://godoc.org/github.com/mmatongo/chew?status.svg)](https://pkg.go.dev/github.com/mmatongo/chew)
[![Maintainability](https://api.codeclimate.com/v1/badges/441cfd36f310c0c48878/maintainability)](https://codeclimate.com/github/mmatongo/chew/maintainability)
[![codecov](https://codecov.io/github/mmatongo/chew/graph/badge.svg?token=6OOK91QQRC)](https://codecov.io/github/mmatongo/chew)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)
</div>

> <p align="center">A Go library for processing various content types into markdown/plaintext..</p>

## About <a id="about"></a>

*Chew* is a Go library that processes various content types into markdown or plaintext. It supports multiple content types, including HTML, PDF, CSV, JSON, YAML, DOCX, PPTX, Markdown, Plaintext, MP3, FLAC, and WAVE.

## Installation <a id="installation"></a>

```bash
go get github.com/mmatongo/chew
```

## Usage <a id="usage"></a>

Here's a basic example of how to use Chew:

```go
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

	// The context is optional, but can be used to cancel the operation after a certain time
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
```

Output

```bash
Source: https://example.com
Content: Example Domain

Source: https://example.com
Content: This domain is for use in illustrative examples in documents. You may use this domain in literature without prior coordination or asking for permission.

Source: https://example.com
Content: More information...
```

You can find more examples in the [examples](./examples) directory as well as instructions on how to use Chew with Ruby and Python.

## Contributing <a id="contributing"></a>

Contributions are welcome! Feel free to open an issue or submit a pull request if you have any suggestions or improvements.

## License <a id="license"></a>

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.

### Logo <a id="logo"></a>

The [logo](https://github.com/MariaLetta/free-gophers-pack) was made by the amazing [MariaLetta](https://github.com/MariaLetta).


### Similar Projects <a id="similar_projects"></a>
[docconv](https://github.com/sajari/docconv)

### Roadmap <a id="roadmap"></a>
The roadmap for this project is available [here](./TODO.md). It's meant more as a guide than a strict plan because I only work on this project in my free time.
