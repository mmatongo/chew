package chew

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/mmatongo/chew/v1/internal/common"
	"github.com/mmatongo/chew/v1/internal/document"
	"github.com/mmatongo/chew/v1/internal/text"
	"github.com/mmatongo/chew/v1/internal/transcribe"
	"github.com/mmatongo/chew/v1/internal/utils"
)

const (
	contentTypeHTML     = "text/html"
	contentTypeText     = "text/plain"
	contentTypePDF      = "application/pdf"
	contentTypeCSV      = "text/csv"
	contentTypeJSON     = "application/json"
	contentTypeYAML     = "application/x-yaml"
	contentTypeMarkdown = "text/markdown"
	contentTypeDocx     = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	contentTypePptx     = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
)

var contentTypeProcessors = map[string]func(io.Reader, string) ([]common.Chunk, error){
	contentTypeHTML:     text.ProcessHTML,
	contentTypeCSV:      text.ProcessCSV,
	contentTypeJSON:     text.ProcessJSON,
	contentTypeYAML:     text.ProcessYAML,
	contentTypeMarkdown: text.ProcessText,
	contentTypeText:     text.ProcessText,
	contentTypeDocx:     document.ProcessDocx,
	contentTypePptx:     document.ProcessPptx,
	contentTypePDF:      document.ProcessPDF,
}

/*
Transcribe uses the Google Cloud Speech-to-Text API to transcribe an audio file. It takes
a context, the filename of the audio file to transcribe, and a TranscribeOptions struct which
contains the Google Cloud credentials, the GCS bucket to upload the audio file to, and the language code
to use for transcription. It returns the transcript of the audio file as a string and an error if the
transcription fails.
*/
var Transcribe = transcribe.Transcribe

type TranscribeOptions = transcribe.TranscribeOptions

/*
This is meant as a fallback in case the content type is not recognized and to enforce
the content type based on the file extension instead of the content type
returned by the server. i.e. if the server returns text/plain but the file is a markdown file
the content types are the biggest culprits of this
*/
var validExtensions = map[string]func(io.Reader, string) ([]common.Chunk, error){
	".md":   text.ProcessText,
	".csv":  text.ProcessCSV,
	".json": text.ProcessJSON,
	".yaml": text.ProcessYAML,
	".html": text.ProcessHTML,
}

/*
For content types that can also return text/plain as their content types we need to manually check
their extension to properly process them. I feel like this could be done better but this is my solution for now.
*/
func getProcessor(contentType, url string) (func(io.Reader, string) ([]common.Chunk, error), error) {
	for key, proc := range contentTypeProcessors {
		if strings.Contains(contentType, key) {
			return proc, nil
		}
	}

	ext, err := utils.GetFileExtensionFromUrl(url)
	if err != nil {
		return nil, fmt.Errorf("couldn't get file extension from url: %s", err)
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
func Process(urls []string, ctxs ...context.Context) ([]common.Chunk, error) {
	ctx := context.Background()
	if len(ctxs) > 0 {
		ctx = ctxs[0]
	}

	var (
		result []common.Chunk
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

func processURL(url string, ctxs ...context.Context) ([]common.Chunk, error) {
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
