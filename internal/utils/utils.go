package utils

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"io"
	"net/url"
	"regexp"
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

/*
Wondering if this is even necessary but I can see how it can be useful
as it also removes links, images, and code blocks.

I'm not sure if this is the best way to remove markdown syntax.
Inspired by https://github.com/mmatongo/site/blob/master/cmd/dnlm/helpers.go#L62-L87
*/

/* RemoveMarkdownSyntax removes markdown syntax from a string */
func RemoveMarkdownSyntax(text string) string {
	patterns := []string{
		"(```[\\s\\S]*?```)",                      // Code blocks
		"(`[^`\n]+`)",                             // Inline code
		"!\\[([^\\]]*?)\\]\\(([^)]+)\\)",          // Images
		"\\[([^\\]]+)\\]\\(([^)]+)\\)",            // Links
		"(__|\\*\\*|_|\\*)(.+?)(__|\\*\\*|_|\\*)", // Bold and Italic
		"~~(.+?)~~",                               // Strikethrough
		"^#{1,6}\\s(.*)$",                         // Headers
		"^>\\s(.*)$",                              // Blockquotes
		"^-{3,}$",                                 // Horizontal rules
		"^\\s*[\\*\\-+]\\s+(.+)$",                 // Unordered lists
		"^\\s*\\d+\\.\\s+(.+)$",                   // Ordered lists
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile("(?m)" + pattern)
		switch {
		case strings.HasPrefix(pattern, "(```"):
			text = re.ReplaceAllString(text, "$1")
		case strings.HasPrefix(pattern, "(`"):
			text = re.ReplaceAllString(text, "$1")
		case strings.HasPrefix(pattern, "!\\["):
			text = re.ReplaceAllString(text, "$1 ($2)")
		case strings.HasPrefix(pattern, "\\["):
			text = re.ReplaceAllString(text, "$1 ($2)")
		case strings.Contains(pattern, "(__|\\*\\*|_|\\*)"):
			text = re.ReplaceAllString(text, "$2")
		case strings.Contains(pattern, "~~"):
			text = re.ReplaceAllString(text, "$1")
		case strings.HasPrefix(pattern, "^#"):
			text = re.ReplaceAllString(text, "$1")
		case strings.HasPrefix(pattern, "^>"):
			text = re.ReplaceAllString(text, "$1")
		case strings.HasPrefix(pattern, "^\\s*[\\*\\-+]"):
			text = re.ReplaceAllString(text, "$1")
		case strings.HasPrefix(pattern, "^\\s*\\d+"):
			text = re.ReplaceAllString(text, "$1")
		default:
			text = re.ReplaceAllString(text, "")
		}
	}

	// Remove any remaining Markdown characters
	text = strings.NewReplacer(
		"*", "",
		"_", "",
		"`", "",
		"#", "",
		">", "",
		"+", "",
		"-", "",
	).Replace(text)

	return strings.TrimSpace(text)
}
