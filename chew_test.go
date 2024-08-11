package chew

import (
	"strings"
	"testing"
)

// type failReader struct{}

// func (f failReader) Read(p []byte) (n int, err error) {
// 	return 0, fmt.Errorf("intentional read failure")
// }

// func TestProcessHTMLError(t *testing.T) {
// 	// Use a reader that always fails
// 	failingReader := failReader{}

// 	_, err := processHTML(failingReader, "failing.html")
// 	if err == nil {
// 		t.Fatalf("expected an error, but got none")
// 	}
// }

// func TestProcessHTMLParseError(t *testing.T) {
// 	// Deliberately broken HTML or unparseable input
// 	badHTML := strings.NewReader("<<not even close to valid HTML>>")

// 	_, err := processHTML(badHTML, "broken.html")
// 	if err == nil {
// 		t.Fatalf("expected an error due to bad HTML, but got none")
// 	}
// }

func TestProcessHTML(t *testing.T) {

	// brokenHTMLContent := `
	// <html><p>Unclosed tag
	// `

	// invalidReader := strings.NewReader(brokenHTMLContent)

	// _, err := processHTML(invalidReader, "invalid.html")
	// if err == nil {
	// 	t.Fatalf("expected an error, but got none")
	// }

	htmlContent := `
    <html>
        <body>
            <header><h1>Header Content</h1></header>
            <nav>Navigation Content</nav>
            <h2>Introduction</h2>
            <p>Welcome to the test case!</p>
            <h3>Details</h3>
            <p>This is a test case for ProcessHTML Function.</p>
            <ul>
                <li>Chew Go Library</li>
                <li>Supports multiple content types, including HTML, PDF, CSV, JSON, YAML, DOCX, PPTX...</li>
            </ul>
            <footer>Footer Content</footer>
        </body>
    </html>`
	reader := strings.NewReader(htmlContent)

	expected := []Chunk{
		{Content: "Introduction", Source: "sample.html"},
		{Content: "Welcome to the test case!", Source: "sample.html"},
		{Content: "Details", Source: "sample.html"},
		{Content: "This is a test case for ProcessHTML Function.", Source: "sample.html"},
		{Content: "Chew Go Library", Source: "sample.html"},
		{Content: "Supports multiple content types, including HTML, PDF, CSV, JSON, YAML, DOCX, PPTX...", Source: "sample.html"},
	}

	chunks, err := processHTML(reader, "sample.html")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(chunks) != len(expected) {
		t.Fatalf("expected %d chunks, got %d", len(expected), len(chunks))
	}

	for i, chunk := range chunks {
		if chunk != expected[i] {
			t.Errorf("at index %d: expected %v, got %v", i, expected[i], chunk)
		}
	}
}

func BenchmarkProcessHTML(b *testing.B) {

	htmlContent := `
    <html>
        <body>
            <header><h1>Header Content</h1></header>
            <nav>Navigation Content</nav>
            <h2>Introduction</h2>
            <p>Welcome to the test case!</p>
            <h3>Details</h3>
            <p>This is a test case for ProcessHTML Function.</p>
            <ul>
                <li>Chew Go Library</li>
                <li>Supports multiple content types, including HTML, PDF, CSV, JSON, YAML, DOCX, PPTX...</li>
            </ul>
            <footer>Footer Content</footer>
        </body>
    </html>`

	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(htmlContent)
		processHTML(reader, "sample.html")
	}
}
