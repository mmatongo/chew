package document

import (
	"archive/zip"
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/mmatongo/chew/v1/internal/common"
)

func createPptxWithContent(content string) io.Reader {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	f, _ := w.Create("ppt/slides/slide1.xml")
	f.Write([]byte(content))
	w.Close()
	return bytes.NewReader(buf.Bytes())
}

func createEmptyPptx() io.Reader {
	return createPptxWithContent(`<?xml version="1.0" encoding="UTF-8"?><document></document>`)
}

func createSingleParagraphPptx(content string) io.Reader {
	return createPptxWithContent(`<?xml version="1.0" encoding="UTF-8"?><document><p>` + content + `</p></document>`)
}

func TestProcessPptx(t *testing.T) {
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
			name:    "Empty pptx file",
			args:    args{r: createEmptyPptx(), url: "http://example.com"},
			want:    nil,
			wantErr: false,
		},
		{
			name:    "Single paragraph pptx file",
			args:    args{r: createSingleParagraphPptx("Hello from chew!"), url: "http://example.com"},
			want:    []common.Chunk{{Content: "Hello from chew! ", Source: "http://example.com"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ProcessPptx(tt.args.r, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessPptx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProcessPptx() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessPptx_Error_ReadAll(t *testing.T) {
	_, err := processPptxContent(&errorReader{})
	if err == nil {
		t.Error("ProcessPptx() did not return an error, but one was expected")
	}
}
