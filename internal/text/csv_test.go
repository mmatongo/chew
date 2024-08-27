package text

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/mmatongo/chew/internal/common"
)

func TestProcessCSV(t *testing.T) {
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
				r:   strings.NewReader("Test content"),
				url: "https://example.com",
			},
			want: []common.Chunk{{
				Content: "Test content",
				Source:  "https://example.com",
			}},
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
		{
			name: "CSV with quoted fields",
			args: args{
				r:   strings.NewReader("\"header 1\",\"header 2\"\n\"value, with comma\",\"value2\""),
				url: "https://example.com/quoted.csv",
			},
			want: []common.Chunk{
				{Content: "header 1, header 2", Source: "https://example.com/quoted.csv"},
				{Content: "value, with comma, value2", Source: "https://example.com/quoted.csv"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ProcessCSV(tt.args.r, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessCSV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProcessCSV() = %v, want %v", got, tt.want)
			}
		})
	}
}
