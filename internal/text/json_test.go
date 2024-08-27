package text

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/mmatongo/chew/v1/internal/common"
)

func TestProcessJSON(t *testing.T) {
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
				r:   strings.NewReader(`{"key": "value"}`),
				url: "https://example.com/data.json",
			},
			want: []common.Chunk{{
				Content: "{\n  \"key\": \"value\"\n}",
				Source:  "https://example.com/data.json",
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
			wantErr: true,
		},
		{
			name: "valid empty json",
			args: args{
				r:   strings.NewReader("{}"),
				url: "https://example.com",
			},
			want: []common.Chunk{{
				Content: "{}",
				Source:  "https://example.com",
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ProcessJSON(tt.args.r, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProcessJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}
