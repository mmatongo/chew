package document

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"

	"github.com/mmatongo/chew/v1/internal/common"
	"github.com/mmatongo/chew/v1/internal/utils"
)

func processPptxContent(r io.Reader) ([]string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	var contents []string

	for _, file := range zipReader.File {
		if strings.HasPrefix(file.Name, "ppt/slides/") {
			slideText, err := utils.ExtractTextFromXML(file)
			if err != nil {
				return nil, err
			}
			contents = append(contents, slideText...)
		}
	}

	var allContent strings.Builder
	for _, content := range contents {
		allContent.WriteString(content)
		allContent.WriteString(" ")
	}

	return []string{allContent.String()}, nil

	/*
		// In the event we just want chunks we can just return contents
		return contents, nil
	*/
}

func ProcessPptx(r io.Reader, url string) ([]common.Chunk, error) {
	content, err := processPptxContent(r)
	if err != nil {
		return nil, err
	}

	var chunks []common.Chunk
	for _, chunk := range content {
		if strings.TrimSpace(string(chunk)) != "" {
			chunks = append(chunks, common.Chunk{Content: string(chunk), Source: url})
		}
	}

	return chunks, nil
}
