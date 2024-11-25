package text

import (
	"bytes"
	"encoding/xml"
	"io"

	"github.com/mmatongo/chew/v1/internal/common"
)

func ProcessXML(r io.Reader, url string) ([]common.Chunk, error) {
	decoder := xml.NewDecoder(r)
	var chunks []common.Chunk
	var currentElement string
	for {
		t, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch se := t.(type) {
		case xml.StartElement:
			currentElement = se.Name.Local
		case xml.CharData:
			content := string(bytes.TrimSpace(se))
			if content != "" && currentElement != "" {
				chunks = append(chunks, common.Chunk{
					Content: content,
					Source:  url,
				})
			}
		}
	}
	return chunks, nil
}
