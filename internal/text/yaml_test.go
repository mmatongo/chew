package text

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/mmatongo/chew/v1/internal/common"
)

func TestProcessYAML(t *testing.T) {
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
				r:   strings.NewReader("key: value\nkey2: value2"),
				url: "https://example.com/data.yaml",
			},
			want: []common.Chunk{
				{
					Content: "key: value\nkey2: value2\n",
					Source:  "https://example.com/data.yaml",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ProcessYAML(tt.args.r, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProcessYAML() = %v, want %v", got, tt.want)
			}
		})
	}
}
