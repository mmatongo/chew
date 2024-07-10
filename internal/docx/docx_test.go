package docx

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createEmptyDocx() io.Reader {
	return createDocxWithContent(`<?xml version="1.0" encoding="UTF-8"?><document></document>`)
}

func createSingleParagraphDocx(content string) io.Reader {
	return createDocxWithContent(`<?xml version="1.0" encoding="UTF-8"?><document><p>` + content + `</p></document>`)
}

func createMultiParagraphDocx(paragraphs []string) io.Reader {
	content := `<?xml version="1.0" encoding="UTF-8"?><document>`
	for _, p := range paragraphs {
		content += `<p>` + p + `</p>`
	}
	content += `</document>`
	return createDocxWithContent(content)
}

func createDocxWithContent(content string) io.Reader {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	f, _ := w.Create("word/document.xml")
	f.Write([]byte(content))
	w.Close()
	return bytes.NewReader(buf.Bytes())
}

func createZipFileWithXML(content string) *zip.File {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	f, _ := w.Create("test.xml")
	f.Write([]byte(content))
	w.Close()

	r, _ := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	return r.File[0]
}

type errorReader struct{}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, assert.AnError
}

func TestProcessDocx(t *testing.T) {
	tests := []struct {
		name     string
		input    io.Reader
		expected []string
		wantErr  bool
	}{
		{
			name:     "Empty docx file",
			input:    createEmptyDocx(),
			expected: []string{},
			wantErr:  false,
		},
		{
			name:     "Single paragraph docx file",
			input:    createSingleParagraphDocx("Hello from chew!"),
			expected: []string{"Hello from chew!"},
			wantErr:  false,
		},
		{
			name:     "Multiple paragraphs docx file",
			input:    createMultiParagraphDocx([]string{"Paragraph one", "Paragraph two", "Paragraph three"}),
			expected: []string{"Paragraph one", "Paragraph two", "Paragraph three"},
			wantErr:  false,
		},
		{
			name:     "Invalid docx file",
			input:    strings.NewReader("Invalid docx file"),
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ProcessDocx(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessDocx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.ElementsMatch(t, tt.expected, got, "ProcessDocx() returned unexpected result")
		})
	}
}

func TestExtractTextFromXML(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		expected []string
		wantErr  bool
	}{
		{
			name:     "Empty XML",
			xml:      `<?xml version="1.0" encoding="UTF-8"?><document></document>`,
			expected: []string{},
			wantErr:  false,
		},
		{
			name:     "Single paragraph",
			xml:      `<?xml version="1.0" encoding="UTF-8"?><document><p>Hello from chew!</p></document>`,
			expected: []string{"Hello from chew!"},
			wantErr:  false,
		},
		{
			name:     "Multiple paragraphs",
			xml:      `<?xml version="1.0" encoding="UTF-8"?><document><p>First</p><p>Second</p><p>Third</p></document>`,
			expected: []string{"First", "Second", "Third"},
			wantErr:  false,
		},
		{
			name:     "Nested elements",
			xml:      `<?xml version="1.0" encoding="UTF-8"?><document><p>Hello <b>bold</b> text</p></document>`,
			expected: []string{"Hello bold text"},
			wantErr:  false,
		},
		{
			name:     "Empty paragraphs",
			xml:      `<?xml version="1.0" encoding="UTF-8"?><document><p></p><p>Stuff</p><p> </p></document>`,
			expected: []string{"Stuff"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := createZipFileWithXML(tt.xml)
			got, err := extractTextFromXML(file)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractTextFromXML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.ElementsMatch(t, tt.expected, got, "extractTextFromXML() returned unexpected result")
		})
	}
}

func TestExtractTextFromXML_Error(t *testing.T) {
	file := createZipFileWithXML(`<?xxml version="1.0" encoding="UTF-8"?><document><p>Hello</p>`)
	_, err := extractTextFromXML(file)
	assert.NotNil(t, err, "extractTextFromXML() did not return an error")
}

func TestProcessDocx_Error_ReadAll(t *testing.T) {
	_, err := ProcessDocx(&errorReader{})
	assert.NotNil(t, err, "ProcessDocx() did not return an error")
}
