Chew is a Go library for processing various content types into markdown.

## Features

- Supports multiple content types: HTML, PDF, CSV, JSON, YAML, and Markdown
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

## Caveats

I used [unipdf](https://github.com/unidoc/unipdf) for PDF processing as it was the most straightforward library I could find (I'm still learning Go). The library is not free, so you will need to purchase a license if you plan to use it in a commercial project. They do offer a free trial, so you can test it out before purchasing. If you have any suggestions for a better library, please let me know.

To use the PDF processing feature, you need to pass the `UNIDOC_LICENSE_KEY` environment variable with your license key.

```bash
export UNIDOC_LICENSE_KEY=your_license_key
```

### TODO

- [ ] Add tests
- [ ] Improve error handling
- [ ] Add support for more content types
- [ ] Implement rate limiting for URL fetching
- [ ] Use a free PDF processing library
- [ ] How to handle text/plain content type


## License
[MIT](./LICENSE)
