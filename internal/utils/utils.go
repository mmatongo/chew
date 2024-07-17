package utils

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"io"
	"net/url"
	"strings"
)

func GetFileExtensionFromUrl(rawUrl string) (string, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return "", err
	}
	pos := strings.LastIndex(u.Path, ".")
	if pos == -1 {
		return "", errors.New("couldn't find a period to indicate a file extension")
	}
	return u.Path[pos:len(u.Path)], nil
}

func ExtractTextFromXML(file *zip.File) ([]string, error) {
	fileReader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer fileReader.Close()

	decoder := xml.NewDecoder(fileReader)
	var contents []string
	var currentParagraph strings.Builder
	inParagraph := false

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
				inParagraph = true
				currentParagraph.Reset()
			}
		case xml.EndElement:
			if element.Name.Local == "p" {
				inParagraph = false
				if trimmed := strings.TrimSpace(currentParagraph.String()); trimmed != "" {
					contents = append(contents, trimmed)
				}
			}
		case xml.CharData:
			if inParagraph {
				currentParagraph.Write(element)
			}
		}
	}

	return contents, nil
}
