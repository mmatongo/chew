package chew

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"encoding/csv"
	"encoding/json"

	"github.com/PuerkitoBio/goquery"
	"github.com/unidoc/unipdf/v3/common/license"
	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
	"gopkg.in/yaml.v3"
)

type Chunk struct {
	Content string
	Source  string
}

const (
	contentTypeHTML     = "text/html"
	contentTypePDF      = "application/pdf"
	contentTypeCSV      = "text/csv"
	contentTypeJSON     = "application/json"
	contentTypeYAML     = "application/x-yaml"
	contentTypeMarkdown = "text/markdown"
)

var contentTypeProcessors = map[string]func(io.Reader, string) ([]Chunk, error){
	contentTypeHTML:     processHTML,
	contentTypePDF:      processPDF,
	contentTypeCSV:      processCSV,
	contentTypeJSON:     processJSON,
	contentTypeYAML:     processYAML,
	contentTypeMarkdown: processMarkdown,
}

func Process(ctx context.Context, urls []string) ([]Chunk, error) {
	var (
		result []Chunk
		wg     sync.WaitGroup
		mu     sync.Mutex
		errCh  = make(chan error, len(urls))
	)

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			chunks, err := processURL(ctx, url)
			if err != nil {
				errCh <- fmt.Errorf("processing %s, %w", url, err)
				return
			}
			mu.Lock()
			result = append(result, chunks...)
			mu.Unlock()
		}(url)
	}

	wg.Wait()
	close(errCh)

	if err := <-errCh; err != nil {
		return nil, err
	}

	return result, nil
}

func processURL(ctx context.Context, url string) ([]Chunk, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")

	var processor func(io.Reader, string) ([]Chunk, error)
	for key, proc := range contentTypeProcessors {
		if strings.Contains(contentType, key) {
			processor = proc
			break
		}
	}

	if processor == nil {
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}

	return processor(resp.Body, url)
}

func processHTML(r io.Reader, url string) ([]Chunk, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	var chunks []Chunk
	doc.Find("p, h1, h2, h3, h4, h5, h6").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		log.Println(text)
		if text != "" {
			chunks = append(chunks, Chunk{Content: text, Source: url})
		}
	})

	fmt.Println(doc)

	return chunks, nil
}

func processPDF(r io.Reader, url string) ([]Chunk, error) {
	if key := os.Getenv("UNIDOC_LICENSE_KEY"); key != "" {
		license.SetMeteredKey(key)
	}

	pdfData, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	pdfReader, err := model.NewPdfReader(bytes.NewReader(pdfData))
	if err != nil {
		return nil, err
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return nil, err
	}

	var chunks []Chunk

	for i := 0; i < numPages; i++ {
		page, err := pdfReader.GetPage(i + 1)
		if err != nil {
			return nil, err
		}

		ex, err := extractor.New(page)
		if err != nil {
			return nil, err
		}

		text, err := ex.ExtractText()
		if err != nil {
			return nil, err
		}

		// Split the text into paragraphs
		paragraphs := strings.Split(text, "\n\n")
		for _, paragraph := range paragraphs {
			trimmed := strings.TrimSpace(paragraph)
			if trimmed != "" {
				chunks = append(chunks, Chunk{
					Content: trimmed,
					Source:  url + "#page=" + strconv.Itoa(i),
				})
			}
		}

	}
	return chunks, nil
}

func processCSV(r io.Reader, url string) ([]Chunk, error) {
	csvReader := csv.NewReader(r)
	var records [][]string
	var err error

	records, err = csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	var chunks []Chunk
	for _, record := range records {
		chunks = append(chunks, Chunk{Content: strings.Join(record, ", "), Source: url})
	}

	return chunks, nil
}

func processJSON(r io.Reader, url string) ([]Chunk, error) {
	var data interface{}
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		return nil, err
	}

	jsonStr, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, err
	}

	return []Chunk{{Content: string(jsonStr), Source: url}}, nil
}

func processYAML(r io.Reader, url string) ([]Chunk, error) {
	var data interface{}
	if err := yaml.NewDecoder(r).Decode(&data); err != nil {
		return nil, err
	}

	yamlStr, err := yaml.Marshal(data)
	if err != nil {
		return nil, err
	}

	return []Chunk{{Content: string(yamlStr), Source: url}}, nil
}

func processMarkdown(r io.Reader, url string) ([]Chunk, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return []Chunk{{Content: string(content), Source: url}}, nil
}
