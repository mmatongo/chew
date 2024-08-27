package document

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/mmatongo/chew/internal/common"
)

func getRootPath(t *testing.T) string {
	t.Helper()
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting current folder: %s", err)
	}
	pwd = filepath.Dir(filepath.Dir(pwd))
	return pwd
}

func TestProcessPDF(t *testing.T) {
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
					f, _ := os.Open(filepath.Join(getRootPath(t), "testdata", "files", "test.pdf"))
					return f
				}(),
				url: "https://example.com/test.pdf",
			},
			want: []common.Chunk{
				{
					Content: "Apdffortesting",
					Source:  "https://example.com/test.pdf#page=1",
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
				r:   func() io.Reader { f, _ := os.Open("nonexistent.pdf"); return f }(),
				url: "https://example.com/nonexistent.pdf",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ProcessPDF(tt.args.r, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessPDF() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProcessPDF() = %v, want %v", got, tt.want)
			}
		})
	}
}
