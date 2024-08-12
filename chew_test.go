package chew

import (
	"bytes"
	"encoding/json"
	"log"
	"strings"
	"testing"

	"github.com/jung-kurt/gofpdf"
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
func TestProcessPDF(t *testing.T) {
	// Sample PDF content in bytes
	pdfContent := generateSamplePDF() // Assume this function returns []byte of a simple PDF

	// Create an io.Reader from the PDF content
	reader := bytes.NewReader(pdfContent)

	// Expected output
	expectedChunks := []Chunk{
		{Content: "Chew is a Go library that processes various content types into markdown or plaintext.\nIt supports multiple content types, including HTML, PDF, CSV, JSON, YAML, DOCX, PPTX, Markdown, Plaintext, MP3, FLAC, and WAVE.", Source: "sample.pdf#page=1"},
	}

	// Call the processPDF function
	chunks, err := processPDF(reader, "sample.pdf")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check the number of chunks
	if len(chunks) != len(expectedChunks) {
		t.Fatalf("expected %d chunks, got %d", len(expectedChunks), len(chunks))
	}

	// Check the content of each chunk
	for i, chunk := range chunks {
		if chunk != expectedChunks[i] {
			t.Errorf("expected chunk %v, got %v", expectedChunks[i], chunk)
		}
	}
}

// Helper function to generate a sample PDF
func generateSamplePDF() []byte {
	var buf bytes.Buffer

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.MultiCell(0, 10, "Chew is a Go library that processes various content types into markdown or plaintext.\n\nIt supports multiple content types, including HTML, PDF, CSV, JSON, YAML, DOCX, PPTX, Markdown, Plaintext, MP3, FLAC, and WAVE.", "", "C", false)

	// Write PDF content to the buffer
	err := pdf.Output(&buf)
	if err != nil {
		log.Fatalf("Failed to generate PDF: %v", err)
	}

	return buf.Bytes()
}

func TestProcessPDF_Error(t *testing.T) {
	// Create an invalid PDF content (e.g., empty content)
	reader := bytes.NewReader([]byte{})

	_, err := processPDF(reader, "sample.pdf")
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
}

func TestProcessCSV(t *testing.T) {
	// Define the test input (CSV data) and expected output
	csvData := `Name, Age, Location
John Doe, 30, New York
Jane Smith, 25, Los Angeles`
	expectedChunks := []Chunk{
		{Content: "Name, Age, Location", Source: "test.csv"},
		{Content: "John Doe, 30, New York", Source: "test.csv"},
		{Content: "Jane Smith, 25, Los Angeles", Source: "test.csv"},
	}

	// Create a reader from the CSV data
	reader := strings.NewReader(csvData)

	// Call the processCSV function
	chunks, err := processCSV(reader, "test.csv")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Compare the actual output with the expected output
	if len(chunks) != len(expectedChunks) {
		t.Fatalf("expected %d chunks, got %d", len(expectedChunks), len(chunks))
	}

	for i, chunk := range chunks {
		if chunk.Content != expectedChunks[i].Content {
			t.Errorf("expected chunk content %q, got %q", expectedChunks[i].Content, chunk.Content)
		}
		if chunk.Source != expectedChunks[i].Source {
			t.Errorf("expected chunk source %q, got %q", expectedChunks[i].Source, chunk.Source)
		}
	}
}

func TestProcessJSON(t *testing.T) {
	jsonData := `{
		"name": "John Doe",
		"age": 30,
		"location": "New York"
	}`

	expectedData := map[string]interface{}{
		"name":     "John Doe",
		"age":      float64(30),
		"location": "New York",
	}

	// Create a reader from the JSON data
	reader := strings.NewReader(jsonData)

	// Call the processJSON function
	chunks, err := processJSON(reader, "test.json")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// The order of a JSON object is not guaranteed in Go
	// so it is safe to decode it into a map and compare
	// the expected map with the resulting map)
	var resultData map[string]interface{}
	if err := json.Unmarshal([]byte(chunks[0].Content), &resultData); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}

	if !equalMaps(resultData, expectedData) {
		t.Errorf("expected chunk content %v, got %v", expectedData, resultData)
	}

	if chunks[0].Source != "test.json" {
		t.Errorf("expected chunk source %q, got %q", "test.json", chunks[0].Source)
	}
}

func equalMaps(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if vb, ok := b[k]; !ok || vb != v {
			return false
		}
	}
	return true
}
