package text

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/mmatongo/chew/v1/internal/common"
)

func TestProcessXML(t *testing.T) {
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
				r:   strings.NewReader("<root><child>Test content</child></root>"),
				url: "https://example.com",
			},
			want: []common.Chunk{{
				Content: "Test content",
				Source:  "https://example.com",
			}},

			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ProcessXML(tt.args.r, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessXML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProcessXML() = %v, want %v", got, tt.want)
			}
		})
	}
}
