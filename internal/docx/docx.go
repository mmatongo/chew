package docx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"strings"
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
			contents, err = extractTextFromXML(file)
			if err != nil {
				return nil, err
			}
			break
		}
	}

	return contents, nil
}

func extractTextFromXML(file *zip.File) ([]string, error) {
	fileReader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer fileReader.Close()

	decoder := xml.NewDecoder(fileReader)
	var contents []string
	var currentParagraph strings.Builder

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch element := token.(type) {
		case xml.StartElement:
			if element.Name.Local == "p" {
				currentParagraph.Reset()
			}
		case xml.EndElement:
			if element.Name.Local == "p" {
				content := strings.TrimSpace(currentParagraph.String())
				if content != "" {
					contents = append(contents, content)
				}
			}
		case xml.CharData:
			currentParagraph.Write(element)
		}
	}

	return contents, nil
}
