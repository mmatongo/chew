package transcribe

import (
	"context"
	"io"
	"net/http"
)

type transcriber interface {
	process(ctx context.Context, filename string, opts TranscribeOptions) (string, error)
}

type whisperTranscriber struct{}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type fileOpener func(name string) (io.ReadCloser, error)
