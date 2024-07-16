package docx

import (
	"archive/zip"
	"bytes"
	"io"

	"github.com/mmatongo/chew/internal/utils"
)

func ProcessDocx(r io.Reader) ([]string, error) {
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
		if file.Name == "word/document.xml" {
			contents, err = utils.ExtractTextFromXML(file)
			if err != nil {
				return nil, err
			}
			break
		}
	}

	return contents, nil
}
