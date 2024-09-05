package chew

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/mmatongo/chew/v1/internal/common"
	"github.com/mmatongo/chew/v1/internal/text"
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

	mockClient := &http.Client{
		Transport: &mockTransport{
			response: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("Test content")),
				Header:     http.Header{"Content-Type": []string{"text/html"}},
			},
		},
	}
	SetHTTPClient(mockClient)
	defer SetHTTPClient(nil)

	contentTypeProcessors = map[string]func(io.Reader, string) ([]common.Chunk, error){
		"text/html":  mockProcessor,
		"text/plain": mockProcessor,
	}
	validExtensions = map[string]func(io.Reader, string) ([]common.Chunk, error){
		".html": mockProcessor,
		".txt":  mockProcessor,
	}

	tempDir := t.TempDir()
	testHTMLPath := filepath.Join(tempDir, "test.html")
	testTXTPath := filepath.Join(tempDir, "test.txt")
	testUnsupportedPath := filepath.Join(tempDir, "test.unsupported")

	err := os.WriteFile(testHTMLPath, []byte("html content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test html file: %v", err)
	}

	err = os.WriteFile(testTXTPath, []byte("text content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test text file: %v", err)
	}

	err = os.WriteFile(testUnsupportedPath, []byte("unsupported content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test unsupported file: %v", err)
	}

	tests := []struct {
		name    string
		url     string
		want    []common.Chunk
		wantErr bool
	}{
		{
			name:    "success",
			url:     "https://example.com/page.html",
			want:    []common.Chunk{{Content: "Test content", Source: "https://example.com/page.html"}},
			wantErr: false,
		},
		{
			name:    "success html",
			url:     "file://" + testHTMLPath,
			want:    []common.Chunk{{Content: "html content", Source: "file://" + testHTMLPath}},
			wantErr: false,
		},
		{
			name:    "success txt",
			url:     "file://" + testTXTPath,
			want:    []common.Chunk{{Content: "text content", Source: "file://" + testTXTPath}},
			wantErr: false,
		},
		{
			name:    "unsupported file type",
			url:     "file://" + testUnsupportedPath,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "non-existent file",
			url:     "file:///non-existent.md",
			want:    nil,
			wantErr: true,
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

func Test_getProcessor(t *testing.T) {
	type args struct {
		contentType string
		url         string
	}
	tests := []struct {
		name    string
		args    args
		want    func(io.Reader, string) ([]common.Chunk, error)
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				contentType: "text/html",
				url:         "https://example.com/page.html",
			},
			want:    mockProcessor,
			wantErr: false,
		},
		{
			name: "unknown content type",
			args: args{
				contentType: "octet/stream",
				url:         "https://example.com/page.html",
			},
			want:    text.ProcessHTML,
			wantErr: false,
		},
		{
			name: "unsupported content type",
			args: args{
				contentType: "application/octet-stream",
				url:         "https://example.com/page.htt",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "no extension",
			args: args{
				contentType: "octet/stream",
				url:         "https://example.com/page",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getProcessor(tt.args.contentType, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("getProcessor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got == nil && tt.want != nil {
				t.Errorf("getProcessor() returned nil, want non-nil")
			} else if got != nil && tt.want == nil {
				t.Errorf("getProcessor() returned non-nil, want nil")
			} else if got != nil {
				gotType := reflect.TypeOf(got)
				wantType := reflect.TypeOf(tt.want)
				if gotType != wantType {
					t.Errorf("getProcessor() returned function of type %v, want %v", gotType, wantType)
				}
			}
		})
	}
}

func TestProcess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/robots.txt":
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("User-agent: *\nDisallow: /disallowed\nCrawl-delay: 1"))
		case "/text":
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("A plain text file."))
		case "/html":
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte("<html><body><p>An HTML file.</p></body></html>"))
		case "/markdown":
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("# A Markdown file"))
		case "/disallowed":
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("This page is disallowed by robots.txt"))
		case "/rate-limited":
			time.Sleep(2 * time.Second)
			w.Write([]byte("Rate limited content"))
		}
	}))
	defer server.Close()

	type args struct {
		urls []string
		ctxs []context.Context
	}
	tests := []struct {
		name            string
		args            args
		want            []common.Chunk
		wantErr         bool
		expectedErrText string
		ignoreRobotsTxt bool
	}{
		{
			name: "plain text",
			args: args{
				urls: []string{server.URL + "/text"},
			},
			want: []common.Chunk{
				{Content: "A plain text file.", Source: server.URL + "/text"},
			},
			wantErr: false,
		},
		{
			name: "HTML",
			args: args{
				urls: []string{server.URL + "/html"},
			},
			want: []common.Chunk{
				{Content: "An HTML file.", Source: server.URL + "/html"},
			},
			wantErr: false,
		},
		{
			name: "markdown",
			args: args{
				urls: []string{server.URL + "/markdown"},
			},
			want: []common.Chunk{
				{Content: "# A Markdown file", Source: server.URL + "/markdown"},
			},
			wantErr: false,
		},
		{
			name: "multiple URLs",
			args: args{
				urls: []string{server.URL + "/text", server.URL + "/html"},
			},
			want: []common.Chunk{
				{Content: "An HTML file.", Source: server.URL + "/html"},
				{Content: "A plain text file.", Source: server.URL + "/text"},
			},
			wantErr: false,
		},
		{
			name: "invalid URL",
			args: args{
				urls: []string{"ftp://invalid.url"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "context cancellation",
			args: args{
				urls: []string{server.URL + "/text"},
				ctxs: []context.Context{func() context.Context {
					ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
					defer cancel()
					return ctx
				}()},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "with more than one context",
			args: args{
				urls: []string{server.URL + "/text"},
				ctxs: []context.Context{context.Background(), context.Background()},
			},
			want:    []common.Chunk{{Content: "A plain text file.", Source: server.URL + "/text"}},
			wantErr: false,
		},
		{
			name: "respects robots.txt",
			args: args{
				urls: []string{server.URL + "/disallowed"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ignores robots.txt when configured",
			args: args{
				urls: []string{server.URL + "/disallowed"},
			},
			want: []common.Chunk{
				{Content: "This page is disallowed by robots.txt", Source: server.URL + "/disallowed"},
			},
			wantErr:         false,
			ignoreRobotsTxt: true,
		},
		{
			name: "robots.txt disallowed",
			args: args{
				urls: []string{server.URL + "/disallowed"},
			},
			want:            nil,
			wantErr:         true,
			expectedErrText: "access to",
		},
		{
			name: "respects crawl delay",
			args: args{
				urls: []string{server.URL + "/text", server.URL + "/html"},
			},
			want: []common.Chunk{
				{Content: "A plain text file.", Source: server.URL + "/text"},
				{Content: "An HTML file.", Source: server.URL + "/html"},
			},
			wantErr: false,
		},
		{
			name: "rate limiting error",
			args: args{
				urls: []string{server.URL + "/rate-limited"},
				ctxs: []context.Context{func() context.Context {
					ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
					defer cancel()
					return ctx
				}()},
			},
			want:            nil,
			wantErr:         true,
			expectedErrText: "rate limiting for",
		},
		{
			name: "crawl delay respect error",
			args: args{
				urls: []string{server.URL + "/text", server.URL + "/text"},
				ctxs: []context.Context{func() context.Context {
					ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
					defer cancel()
					return ctx
				}()},
			},
			want:            nil,
			wantErr:         true,
			expectedErrText: "respecting crawl delay for",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldState := ignoreRobotsTxt
			ignoreRobotsTxt = tt.ignoreRobotsTxt
			defer func() { ignoreRobotsTxt = oldState }()

			start := time.Now()
			got, err := Process(tt.args.urls, tt.args.ctxs...)
			duration := time.Since(start)

			if (err != nil) != tt.wantErr {
				t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Process() = %v, want %v", got, tt.want)
			}

			if tt.name == "respects crawl delay" && duration < time.Second {
				t.Errorf("Crawl delay not respected. Duration: %v", duration)
			}
		})
	}
}

func Test_getNextProxy(t *testing.T) {
	tests := []struct {
		name string
		want *url.URL
	}{
		{
			name: "no proxies",
			want: nil,
		},
		{
			name: "single proxy",
			want: &url.URL{Scheme: "http", Host: "localhost:8080"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "single proxy" {
				NewConfig(Config{
					ProxyList: []string{"http://localhost:8080"},
				})
			}
			if got := getNextProxy(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getNextProxy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRespectCrawlDelay(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		delay    time.Duration
		wantWait bool
	}{
		{
			name:     "first access",
			url:      "https://example.com",
			delay:    time.Second,
			wantWait: false,
		},
		{
			name:     "second access",
			url:      "https://example.com",
			delay:    time.Second,
			wantWait: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			err := respectCrawlDelay(tt.url, tt.delay)
			duration := time.Since(start)

			if err != nil {
				t.Errorf("respectCrawlDelay() error = %v", err)
				return
			}

			if tt.wantWait && duration < tt.delay {
				t.Errorf("respectCrawlDelay() didn't wait long enough. Duration: %v, Expected: %v", duration, tt.delay)
			}
			if !tt.wantWait && duration >= tt.delay {
				t.Errorf("respectCrawlDelay() waited unnecessarily. Duration: %v", duration)
			}
		})
	}
}
