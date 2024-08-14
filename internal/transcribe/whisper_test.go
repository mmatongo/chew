package transcribe

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func Test_processWhisper(t *testing.T) {
	/* mocks */

	tempDir, err := os.MkdirTemp("", "whisper_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFilePath := filepath.Join(tempDir, "test.mp3")
	if err := os.WriteFile(testFilePath, []byte("dummy audio content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	unreadableFilePath := filepath.Join(tempDir, "unreadable.mp3")
	if err := os.WriteFile(unreadableFilePath, []byte("unreadable content"), 0000); err != nil {
		t.Fatalf("failed to create unreadable file: %v", err)
	}

	successfulMockClient := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"text": "this is a test transcription."
				}`)),
			}, nil
		},
	}

	errorMockClient := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("mock HTTP error")
		},
	}

	badResponseMockClient := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 400,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error": "Bad Request"}`)),
			}, nil
		},
	}

	invalidJSONMockClient := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`invalid JSON`)),
			}, nil
		},
	}

	type args struct {
		ctx      context.Context
		filename string
		opts     TranscribeOptions
		client   httpClient
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "successful transcription",
			args: args{
				ctx:      context.Background(),
				filename: testFilePath,
				opts: TranscribeOptions{
					WhisperAPIKey: "test-api-key",
					WhisperModel:  "test-model",
					LanguageCode:  "en-US",
					WhisperPrompt: "test-prompt",
				},
				client: successfulMockClient,
			},
			want:    "this is a test transcription.",
			wantErr: false,
		},
		{
			name: "file open error",
			args: args{
				ctx:      context.Background(),
				filename: "non-existent-file.mp3",
				opts:     TranscribeOptions{},
				client:   successfulMockClient,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "file read error",
			args: args{
				ctx:      context.Background(),
				filename: unreadableFilePath,
				opts:     TranscribeOptions{},
				client:   successfulMockClient,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "HTTP client error",
			args: args{
				ctx:      context.Background(),
				filename: testFilePath,
				opts:     TranscribeOptions{},
				client:   errorMockClient,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "bad response from API",
			args: args{
				ctx:      context.Background(),
				filename: testFilePath,
				opts:     TranscribeOptions{},
				client:   badResponseMockClient,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "invalid JSON response",
			args: args{
				ctx:      context.Background(),
				filename: testFilePath,
				opts:     TranscribeOptions{},
				client:   invalidJSONMockClient,
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processWhisper(tt.args.ctx, tt.args.filename, tt.args.opts, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("processWhisper() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("processWhisper() = %v, want %v", got, tt.want)
			}
		})
	}
}
