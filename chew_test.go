package chew

import (
	"strings"
	"testing"
)

func TestProcessHTML(t *testing.T) {
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
