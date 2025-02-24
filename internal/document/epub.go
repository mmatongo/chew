package document

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmatongo/chew/v1/internal/common"
	"github.com/taylorskalyo/goreader/epub"
)

func processEpubContent(r io.Reader) ([]common.Chunk, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read EPUB content: %w", err)
	}

	reader, err := epub.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return nil, fmt.Errorf("failed to create EPUB reader: %w", err)
	}

	if len(reader.Rootfiles) == 0 {
		return nil, fmt.Errorf("EPUB contains no content")
	}

	contents := reader.Rootfiles[0]
	var chunks []common.Chunk

	for _, item := range contents.Manifest.Items {
		if !strings.HasSuffix(item.HREF, ".xhtml") && !strings.HasSuffix(item.HREF, ".html") {
			continue
		}

		file, err := item.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open item %s: %w", item.HREF, err)
		}

		text, err := extractTextFromHTML(file)
		file.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to extract text from %s: %w", item.HREF, err)
		}

		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		chunks = append(chunks, common.Chunk{Content: text, Source: item.HREF})
	}

	return chunks, nil
}

func ProcessEpub(r io.Reader, url string) ([]common.Chunk, error) {
	chunks, err := processEpubContent(r)
	if err != nil {
		return nil, err
	}

	for i := range chunks {
		chunks[i].Source = url
	}

	return chunks, nil
}

func extractTextFromHTML(r io.Reader) (string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return "", err
	}

	doc.Find("script, style,nav, header, footer").Remove()

	var buf strings.Builder
	/*
		We're only interested in the text content of the HTML document
		however this is a very naive approach and might not work well
		for all HTML documents unfortunately.
		This is a known issue and I'm working on a better solution.
		see: https://github.com/mmatongo/chew/issues/22

		TODO: Allow users to specify a CSS selector to extract text from
	*/
	doc.Find("p, h1, h2, h3, h4, h5, h6, li").Each(func(_ int, s *goquery.Selection) {
		buf.WriteString(strings.TrimSpace(s.Text()))
		buf.WriteString("\n\n")
	})

	return strings.TrimSpace(buf.String()), nil
}
