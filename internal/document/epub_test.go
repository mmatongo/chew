package document

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/mmatongo/chew/v1/internal/common"
	"github.com/mmatongo/chew/v1/internal/utils"
)

func Test_processEpubContent(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    []common.Chunk
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				r: func() io.Reader {
					f, _ := os.Open(filepath.Join(getRootPath(t), "testdata", "files", "test.epub"))
					return f
				}(),
			},
			want: []common.Chunk{
				{
					Content: "A pdf for testing",
					Source:  "index.html",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processEpubContent(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("processEpubContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processEpubContent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessEpub(t *testing.T) {
	type args struct {
		r   io.Reader
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    []common.Chunk
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				r: func() io.Reader {
					f, _ := os.Open(filepath.Join(getRootPath(t), "testdata", "files", "test.epub"))
					return f
				}(),
				url: "https://example.com/test.epub",
			},
			want: []common.Chunk{
				{
					Content: "A pdf for testing",
					Source:  "https://example.com/test.epub",
				},
			},
			wantErr: false,
		},
		{
			name: "error",
			args: args{
				r:   strings.NewReader("key: value, key2: value2"),
				url: "https://example.com/data.yaml",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "empty",
			args: args{
				r:   strings.NewReader(""),
				url: "https://example.com",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "unreadable",
			args: args{
				r:   func() io.Reader { f, _ := os.Open("nonexistent.epub"); return f }(),
				url: "https://example.com/nonexistent.epub",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ProcessEpub(tt.args.r, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessEpub() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProcessEpub() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractTextFromHTML(t *testing.T) {
	file, _ := utils.OpenFile("testdata/invalid.html")
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				r: strings.NewReader("<html><body><h1>some content</h1></body></html>"),
			},
			want:    "some content",
			wantErr: false,
		},
		{
			name: "error",
			args: args{
				r: file,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractTextFromHTML(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractTextFromHTML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractTextFromHTML() = %v, want %v", got, tt.want)
			}
		})
	}
}
