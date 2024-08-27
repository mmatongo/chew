package text

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/mmatongo/chew/internal/common"
)

func TestProcessHTML(t *testing.T) {
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
				r: strings.NewReader(`
					<!DOCTYPE html>
					<html>
					<head>
						<title>Test HTML</title>
					</head>
					<body>
						<h1>Test content</h1>
						<p>This is a test paragraph.</p>
					</body>
					</html>
				`),
				url: "https://example.com/page.html",
			},
			want: []common.Chunk{
				{
					Content: "Test content",
					Source:  "https://example.com/page.html",
				},
				{
					Content: "This is a test paragraph.",
					Source:  "https://example.com/page.html",
				},
			},
			wantErr: false,
		},
		{
			name: "empty",
			args: args{
				r:   strings.NewReader(""),
				url: "https://example.com",
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ProcessHTML(tt.args.r, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("processHTML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processHTML() = %v, want %v", got, tt.want)
			}
		})
	}
}
