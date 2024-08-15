package chew

import (
	"encoding/json"
	"io"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestProcessHTML(t *testing.T) {

	t.Run("broken HTML", func(t *testing.T) {

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
	})

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

func TestProcessData(t *testing.T) {
	tests := []struct {
		name           string
		data           string
		expectedData   interface{}
		processFunc    func(io.Reader, string) ([]Chunk, error)
		expectedSource string
	}{
		{
			name: "JSON",
			data: `{
				"language": "Go",
				"role": "Backend Engineer",
				"location": "New York"
			}`,
			expectedData: map[string]interface{}{
				"language": "Go",
				"role":     "Backend Engineer",
				"location": "New York",
			},
			processFunc:    processJSON,
			expectedSource: "test.json",
		},
		{
			name: "YAML",
			data: `
language: Go
role: Backend Engineer
location: New York
`,
			expectedData: map[string]interface{}{
				"language": "Go",
				"role":     "Backend Engineer",
				"location": "New York",
			},
			processFunc:    processYAML,
			expectedSource: "test.yaml",
		},
		{
			name: "CSV",
			data: `Language,Role,Location
Go,Backend Engineer,New York
Rust,Systems Engineer,Los Angeles`,
			expectedData: []Chunk{
				{Content: "Language,Role,Location", Source: "test.csv"},
				{Content: "Go,Backend Engineer,New York", Source: "test.csv"},
				{Content: "Rust,Systems Engineer,Los Angeles", Source: "test.csv"},
			},
			processFunc:    processCSV,
			expectedSource: "test.csv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.data)
			chunks, err := tt.processFunc(reader, tt.expectedSource)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			switch expected := tt.expectedData.(type) {
			case map[string]interface{}:
				var resultData map[string]interface{}
				if tt.name == "JSON" {
					if err := json.Unmarshal([]byte(chunks[0].Content), &resultData); err != nil {
						t.Fatalf("failed to unmarshal result: %v", err)
					}
				} else if tt.name == "YAML" {
					if err := yaml.Unmarshal([]byte(chunks[0].Content), &resultData); err != nil {
						t.Fatalf("failed to unmarshal result: %v", err)
					}
				}

				if len(chunks) != 1 {
					t.Fatalf("expected 1 chunk, got %d", len(chunks))
				}

				if !equalMaps(resultData, expected) {
					t.Errorf("expected chunk content %v, got %v", expected, resultData)
				}

				if chunks[0].Source != tt.expectedSource {
					t.Errorf("expected chunk source %q, got %q", tt.expectedSource, chunks[0].Source)
				}

			case []Chunk:
				if len(chunks) != len(expected) {
					t.Fatalf("expected %d chunks, got %d", len(expected), len(chunks))
				}

				for i, chunk := range chunks {
					if chunk.Content != expected[i].Content {
						t.Errorf("expected chunk content %q, got %q", expected[i].Content, chunk.Content)
					}
					if chunk.Source != expected[i].Source {
						t.Errorf("expected chunk source %q, got %q", expected[i].Source, chunk.Source)
					}
				}
			default:
				t.Fatalf("unexpected expectedData type %T", expected)
			}
		})
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

/*
Comments:
- Both Docx and Pptx tests are skipped because they might require refactoring the processDocx and processPptx functions in chew.go so that each accept a processing function as one of their arguments.
- The Test for the getProcessor function is also skipped because every attempt to test it will require a refactoring of the getProcessor function in the chew.go code.
- The processURL test depends on the getProcessor function, so it is also skipped.
*/
