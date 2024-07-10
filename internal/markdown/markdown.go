package markdown

import (
	"regexp"
	"strings"
)

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
