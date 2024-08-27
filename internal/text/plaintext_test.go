package text

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/mmatongo/chew/internal/common"
)

func TestProcessText(t *testing.T) {
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ProcessText(tt.args.r, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProcessText() = %v, want %v", got, tt.want)
			}
		})
	}
}
