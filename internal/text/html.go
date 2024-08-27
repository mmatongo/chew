package text

import (
	"fmt"
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmatongo/chew/internal/common"
)

func ProcessHTML(r io.Reader, url string) ([]common.Chunk, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var chunks []common.Chunk
	/*
		We're only interested in the text content of the HTML document
		so we're going to ignore the tags that don't contain useful text.
		This is a very naive approach and might not work for all HTML documents unfortunately
	*/

	doc.Find("nav, header, footer").Remove()

	doc.Find("p, h1, h2, h3, h4, h5, h6, li").Each(func(_ int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			chunks = append(chunks, common.Chunk{Content: text, Source: url})
		}
	})

	return chunks, nil
}
