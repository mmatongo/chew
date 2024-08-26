package chew

import (
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/mmatongo/chew/internal/common"
)

func mockProcessor(r io.Reader, url string) ([]common.Chunk, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return []common.Chunk{{Content: string(content), Source: url}}, nil
}

type mockTransport struct {
	response *http.Response
	err      error
}

func (m *mockTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return m.response, m.err
}

func Test_processURL(t *testing.T) {
	originalHTTPClient := http.DefaultClient
	originalContentTypeProcessors := contentTypeProcessors
	originalValidExtensions := validExtensions

	defer func() {
		http.DefaultClient = originalHTTPClient
		contentTypeProcessors = originalContentTypeProcessors
		validExtensions = originalValidExtensions
	}()

	mockTransport := &mockTransport{
		response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("Test content")),
			Header:     http.Header{"Content-Type": []string{"text/html"}},
		},
		err: nil,
	}

	http.DefaultClient = &http.Client{
		Transport: mockTransport,
	}

	contentTypeProcessors = map[string]func(io.Reader, string) ([]common.Chunk, error){
		"text/html": mockProcessor,
	}
	validExtensions = map[string]func(io.Reader, string) ([]common.Chunk, error){
		"html": mockProcessor,
	}

	tests := []struct {
		name    string
		url     string
		want    []common.Chunk
		wantErr bool
	}{
		{
			name:    "Valid URL",
			url:     "https://example.com/page.html",
			want:    []common.Chunk{{Content: "Test content", Source: "https://example.com/page.html"}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("processURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
