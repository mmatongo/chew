package chew

import (
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestProcessHTML(t *testing.T) {

	//the goquery parser couldn't catch a broken HTML content,
	//so the test fails because it expects an error but got none
	brokenHTMLContent := `
	<html><p>Unclosed tag
	`

	invalidReader := strings.NewReader(brokenHTMLContent)

	_, err := processHTML(invalidReader, "invalid.html")
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

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

// just a little playaround here
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

func TestProcessCSV(t *testing.T) {
	// Define the test input (CSV data) and expected output
	csvData := `Language, Role, Location
Go, Backend Engineer, New York
Rust, Systems Engineer, Los Angeles`
	expectedChunks := []Chunk{
		{Content: "Language, Role, Location", Source: "test.csv"},
		{Content: "Go, Backend Engineer, New York", Source: "test.csv"},
		{Content: "Rust, Systems Engineer, Los Angeles", Source: "test.csv"},
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
		"language": "Go",
		"role": "Backend Engineer",
		"location": "New York"
	}`

	expectedData := map[string]interface{}{
		"language": "Go",
		"role":     "Backend Engineer",
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

func TestProcessYAML(t *testing.T) {
	// Define the test input (YAML data)
	yamlData := `
language: Go
role: Backend Engineer
location: New York
`

	expectedData := map[string]interface{}{
		"language": "Go",
		"role":     "Backend Engineer",
		"location": "New York",
	}

	// Create a reader from the YAML data
	reader := strings.NewReader(yamlData)

	// Call the processYAML function
	chunks, err := processYAML(reader, "test.yaml")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Decode the result into a map
	var resultData map[string]interface{}
	if err := yaml.Unmarshal([]byte(chunks[0].Content), &resultData); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	// Compare the actual output with the expected output
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}

	// Compare the resulting map with the expected map
	if !equalMaps(resultData, expectedData) {
		t.Errorf("expected chunk content %v, got %v", expectedData, resultData)
	}

	if chunks[0].Source != "test.yaml" {
		t.Errorf("expected chunk source %q, got %q", "test.yaml", chunks[0].Source)
	}
}

/*
Comments:
- Both Docx and Pptx tests are skipped because they might require refactoring the processDocx and processPptx functions in chew.go so that each accept a processing function as one of their arguments.
- The Test for the getProcessor function is also skipped because every attempt to test it will require a refactoring of the getProcessor function in the chew.go code.
- The processURL test depends on the getProcessor function, so it is also skipped.
*/
