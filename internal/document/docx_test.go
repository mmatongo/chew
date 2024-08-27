package document

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"reflect"
	"testing"

	"github.com/mmatongo/chew/internal/common"
)

type errorReader struct{}

var errMockRead = errors.New("mock read error")

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, errMockRead
}

func createDocxWithContent(content string) io.Reader {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	f, _ := w.Create("word/document.xml")
	f.Write([]byte(content))
	w.Close()
	return bytes.NewReader(buf.Bytes())
}

func createEmptyDocx() io.Reader {
	return createDocxWithContent(`<?xml version="1.0" encoding="UTF-8"?><document></document>`)
}

func createSingleParagraphDocx(content string) io.Reader {
	return createDocxWithContent(`<?xml version="1.0" encoding="UTF-8"?><document><p>` + content + `</p></document>`)
}

func TestProcessDocx(t *testing.T) {
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
			name: "Empty docx file",
			args: args{
				r:   createEmptyDocx(),
				url: "http://example.com",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Single paragraph docx file",
			args: args{
				r:   createSingleParagraphDocx("Hello from chew!"),
				url: "http://example.com",
			},
			want: []common.Chunk{
				{
					Content: "Hello from chew! ",
					Source:  "http://example.com",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ProcessDocx(tt.args.r, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessDocx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProcessDocx() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessDocx_Error_ReadAll(t *testing.T) {
	_, err := processPptxContent(&errorReader{})
	if err == nil {
		t.Error("ProcessDocx() did not return an error, but one was expected")
	}
}
