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

	"github.com/mmatongo/chew/internal/docx"
	"github.com/mmatongo/chew/internal/markdown"
	"github.com/mmatongo/chew/internal/utils"
)

type Chunk struct {
	Content string
	Source  string
}

const (
	contentTypeHTML     = "text/html"
	contentTypeText     = "text/plain"
	contentTypePDF      = "application/pdf"
	contentTypeCSV      = "text/csv"
	contentTypeJSON     = "application/json"
	contentTypeYAML     = "application/x-yaml"
	contentTypeMarkdown = "text/markdown"
	contentTypeDocx     = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
)

var contentTypeProcessors = map[string]func(io.Reader, string) ([]Chunk, error){
	contentTypeHTML:     processHTML,
	contentTypePDF:      processPDF,
	contentTypeCSV:      processCSV,
	contentTypeJSON:     processJSON,
	contentTypeYAML:     processYAML,
	contentTypeMarkdown: processMarkdown,
	contentTypeDocx:     processDocx,
	contentTypeText:     processText,
}

/*
This is meant as a fallback in case the content type is not recognized and to enforce
the content type based on the file extension instead of the content type
returned by the server. i.e. if the server returns text/plain but the file is a markdown file
the content types are the biggest culprits of this
*/
var validExtensions = map[string]func(io.Reader, string) ([]Chunk, error){
	".md":   processMarkdown,
	".csv":  processCSV,
	".json": processJSON,
	".yaml": processYAML,
	".html": processHTML,
}

func getProcessor(contentType, url string) (func(io.Reader, string) ([]Chunk, error), error) {
	for key, proc := range contentTypeProcessors {
		if strings.Contains(contentType, key) {
			return proc, nil
		}
	}

	ext, err := utils.GetFileExtensionFromUrl(url)
	if err != nil {
		log.Printf("couldn't get file extension from url: %v", err)
	}

	if proc, ok := validExtensions[ext]; ok {
		return proc, nil
	}

	return nil, fmt.Errorf("unsupported content type: %s", contentType)
}

/*
Process takes a list of URLs and returns a list of Chunks

The slice of strings to be processed can be URLs or file paths
The context is optional and can be used to cancel the processing
of the URLs after a certain amount of time
*/
func Process(urls []string, ctxs ...context.Context) ([]Chunk, error) {
	ctx := context.Background()
	if len(ctxs) > 0 {
		ctx = ctxs[0]
	}

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
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			default:
				chunks, err := processURL(url, ctx)
				if err != nil {
					errCh <- fmt.Errorf("processing %s, %w", url, err)
					return
				}
				mu.Lock()
				result = append(result, chunks...)
				mu.Unlock()
			}
		}(url)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return nil, err
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return result, nil
}

func processURL(url string, ctxs ...context.Context) ([]Chunk, error) {
	ctx := context.Background()
	if len(ctxs) > 0 {
		ctx = ctxs[0]
	}

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

	processor, err := getProcessor(contentType, url)
	if err != nil {
		return nil, err
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
		if text != "" {
			chunks = append(chunks, Chunk{Content: text, Source: url})
		}
	})

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
	sanitizedMarkdown := markdown.RemoveMarkdownSyntax(string(content))

	return []Chunk{{Content: sanitizedMarkdown, Source: url}}, nil
}

func processDocx(r io.Reader, url string) ([]Chunk, error) {
	content, err := docx.ProcessDocx(r)
	if err != nil {
		return nil, err
	}

	var chunks []Chunk
	for _, chunk := range content {
		if strings.TrimSpace(string(chunk)) != "" {
			chunks = append(chunks, Chunk{Content: string(chunk), Source: url})
		}
	}

	return chunks, nil
}

/*
Not entirely sure about this one because I've had instances where the
content type is text/plain but the content is actually HTML
So I'm just going to leave it here for now
*/
func processText(r io.Reader, url string) ([]Chunk, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return []Chunk{{Content: string(content), Source: url}}, nil
}
