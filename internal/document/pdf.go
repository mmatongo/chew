package document

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/mmatongo/chew/internal/common"
)

func ProcessPDF(r io.Reader, url string) ([]common.Chunk, error) {
	pdfData, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	f, err := pdf.NewReader(bytes.NewReader(pdfData), int64(len(pdfData)))
	if err != nil {
		return nil, err
	}

	var chunks []common.Chunk
	for i := 1; i <= f.NumPage(); i++ {
		p := f.Page(i)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err != nil {
			log.Printf("Error extracting text from page %d: %v\n\n", i, err)
			continue
		}

		text = strings.TrimSpace(text)
		text = strings.ReplaceAll(text, "\n", "\n\n")

		chunks = append(chunks, common.Chunk{
			Content: text,
			Source:  fmt.Sprintf("%s#page=%d", url, i),
		})
	}

	if len(chunks) == 0 {
		return nil, err
	}

	return chunks, nil
}
