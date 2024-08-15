package chew

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func createMockPPTXReader() io.Reader {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	var files = []struct {
		Name, Body string
	}{
		{"ppt/slides/slide1.xml", "<p:sld><p:txBody><a:p><a:r><a:t>Slide 1 content</a:t></a:r></a:p></p:txBody></p:sld>"},
		{"ppt/slides/slide2.xml", "<p:sld><p:txBody><a:p><a:r><a:t>Slide 2 content</a:t></a:r></a:p></p:txBody></p:sld>"},
	}
	for _, file := range files {
		f, err := w.Create(file.Name)
		if err != nil {
			panic(err)
		}
		_, err = f.Write([]byte(file.Body))
		if err != nil {
			panic(err)
		}
	}

	err := w.Close()
	if err != nil {
		panic(err)
	}

	return bytes.NewReader(buf.Bytes())
}

func createEmptyPPTXReader() io.Reader {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	err := w.Close()
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(buf.Bytes())
}

func createMockDOCXReader() io.Reader {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	f, err := w.Create("word/document.xml")
	if err != nil {
		panic(err)
	}
	_, err = f.Write([]byte(`
		<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
			<w:body>
				<w:p>
					<w:r>
						<w:t>This is a test document.</w:t>
					</w:r>
				</w:p>
				<w:p>
					<w:r>
						<w:t>It has multiple paragraphs.</w:t>
					</w:r>
				</w:p>
			</w:body>
		</w:document>
	`))
	if err != nil {
		panic(err)
	}

	err = w.Close()
	if err != nil {
		panic(err)
	}

	return bytes.NewReader(buf.Bytes())
}

func createEmptyDOCXReader() io.Reader {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	f, err := w.Create("word/document.xml")
	if err != nil {
		panic(err)
	}
	_, err = f.Write([]byte(`
		<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
			<w:body></w:body>
		</w:document>
	`))
	if err != nil {
		panic(err)
	}

	err = w.Close()
	if err != nil {
		panic(err)
	}

	return bytes.NewReader(buf.Bytes())
}

func mockProcessor(r io.Reader, url string) ([]Chunk, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return []Chunk{{Content: string(content), Source: url}}, nil
}

func Test_processText(t *testing.T) {
	type args struct {
		r   io.Reader
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    []Chunk
		wantErr bool
	}{
		{
			name: "Valid input",
			args: args{
				r:   strings.NewReader("Test content"),
				url: "https://example.com",
			},
			want: []Chunk{{
				Content: "Test content",
				Source:  "https://example.com",
			}},
			wantErr: false,
		},
		{
			name: "Empty input",
			args: args{
				r:   strings.NewReader(""),
				url: "https://example.com",
			},
			want: []Chunk{{
				Content: "",
				Source:  "https://example.com",
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processText(tt.args.r, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("processText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processText() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_processPptx(t *testing.T) {
	type args struct {
		r   io.Reader
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    []Chunk
		wantErr bool
	}{
		{
			name: "Valid PPTX input",
			args: args{
				r:   createMockPPTXReader(),
				url: "https://example.com/presentation.pptx",
			},
			want: []Chunk{
				{
					Content: "Slide 1 content Slide 2 content ",
					Source:  "https://example.com/presentation.pptx",
				},
			},
			wantErr: false,
		},
		{
			name: "Empty PPTX input",
			args: args{
				r:   createEmptyPPTXReader(),
				url: "https://example.com/empty.pptx",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Invalid input",
			args: args{
				r:   strings.NewReader("This is not a valid PPTX file"),
				url: "https://example.com/invalid.pptx",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processPptx(tt.args.r, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("processPptx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processPptx() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_processDocx(t *testing.T) {
	type args struct {
		r   io.Reader
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    []Chunk
		wantErr bool
	}{
		{
			name: "Valid DOCX input",
			args: args{
				r:   createMockDOCXReader(),
				url: "https://example.com/document.docx",
			},
			want: []Chunk{
				{
					Content: "This is a test document. It has multiple paragraphs. ",
					Source:  "https://example.com/document.docx",
				},
			},
			wantErr: false,
		},
		{
			name: "Empty DOCX input",
			args: args{
				r:   createEmptyDOCXReader(),
				url: "https://example.com/empty.docx",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Invalid input",
			args: args{
				r:   strings.NewReader("This is not a valid DOCX file"),
				url: "https://example.com/invalid.docx",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processDocx(tt.args.r, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("processDocx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processDocx() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_processYAML(t *testing.T) {
	type args struct {
		r   io.Reader
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    []Chunk
		wantErr bool
	}{
		{
			name: "Valid YAML input",
			args: args{
				r:   strings.NewReader("key: value\nkey2: value2"),
				url: "https://example.com/data.yaml",
			},
			want: []Chunk{
				{
					Content: "key: value\nkey2: value2\n",
					Source:  "https://example.com/data.yaml",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processYAML(tt.args.r, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("processYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processYAML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_processJSON(t *testing.T) {
	type args struct {
		r   io.Reader
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    []Chunk
		wantErr bool
	}{
		{
			name: "Valid JSON input",
			args: args{
				r:   strings.NewReader(`{"key": "value", "key2": "value2"}`),
				url: "https://example.com/data.json",
			},
			want: []Chunk{
				{
					Content: "{\n  \"key\": \"value\",\n  \"key2\": \"value2\"\n}",
					Source:  "https://example.com/data.json",
				},
			},
			wantErr: false,
		},
		{
			name: "Empty JSON input",
			args: args{
				r:   strings.NewReader(`{}`),
				url: "https://example.com/empty.json",
			},
			want: []Chunk{
				{
					Content: "{}",
					Source:  "https://example.com/empty.json",
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid JSON input",
			args: args{
				r:   strings.NewReader(`{"key": "value",}`),
				url: "https://example.com/invalid.json",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processJSON(tt.args.r, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("processJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_processCSV(t *testing.T) {
	type args struct {
		r   io.Reader
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    []Chunk
		wantErr bool
	}{
		{
			name: "Valid CSV input",
			args: args{
				r:   strings.NewReader("key1,key2\nvalue1,value2"),
				url: "https://example.com/data.csv",
			},
			want: []Chunk{
				{Content: "key1, key2", Source: "https://example.com/data.csv"},
				{Content: "value1, value2", Source: "https://example.com/data.csv"},
			},
			wantErr: false,
		},
		{
			name: "Empty CSV input",
			args: args{
				r:   strings.NewReader(""),
				url: "https://example.com/empty.csv",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "CSV with one column",
			args: args{
				r:   strings.NewReader("header\nvalue"),
				url: "https://example.com/single_column.csv",
			},
			want: []Chunk{
				{Content: "header", Source: "https://example.com/single_column.csv"},
				{Content: "value", Source: "https://example.com/single_column.csv"},
			},
			wantErr: false,
		},
		{
			name: "CSV with quoted fields",
			args: args{
				r:   strings.NewReader("\"header 1\",\"header 2\"\n\"value, with comma\",\"value2\""),
				url: "https://example.com/quoted.csv",
			},
			want: []Chunk{
				{Content: "header 1, header 2", Source: "https://example.com/quoted.csv"},
				{Content: "value, with comma, value2", Source: "https://example.com/quoted.csv"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processCSV(tt.args.r, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("processCSV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processCSV() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_processHTML(t *testing.T) {
	type args struct {
		r   io.Reader
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    []Chunk
		wantErr bool
	}{
		{
			name: "Valid HTML input",
			args: args{
				r: strings.NewReader(`
					<!DOCTYPE html>
					<html>
					<head>
						<title>Test HTML</title>
					</head>
					<body>
						<h1>Test content</h1>
						<p>This is a test paragraph.</p>
					</body>
					</html>
				`),
				url: "https://example.com/page.html",
			},
			want: []Chunk{
				{
					Content: "Test content",
					Source:  "https://example.com/page.html",
				},
				{
					Content: "This is a test paragraph.",
					Source:  "https://example.com/page.html",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processHTML(tt.args.r, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("processHTML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processHTML() = %v, want %v", got, tt.want)
			}
		})
	}
}

type mockTransport struct {
	response *http.Response
	err      error
}

func (m *mockTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return m.response, m.err
}

func Test_processURL(t *testing.T) {
	originalHTTPClient := http.DefaultClient
	originalContentTypeProcessors := contentTypeProcessors
	originalValidExtensions := validExtensions

	defer func() {
		http.DefaultClient = originalHTTPClient
		contentTypeProcessors = originalContentTypeProcessors
		validExtensions = originalValidExtensions
	}()

	mockTransport := &mockTransport{
		response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("Test content")),
			Header:     http.Header{"Content-Type": []string{"text/html"}},
		},
		err: nil,
	}

	http.DefaultClient = &http.Client{
		Transport: mockTransport,
	}

	contentTypeProcessors = map[string]func(io.Reader, string) ([]Chunk, error){
		"text/html": mockProcessor,
	}
	validExtensions = map[string]func(io.Reader, string) ([]Chunk, error){
		"html": mockProcessor,
	}

	tests := []struct {
		name    string
		url     string
		want    []Chunk
		wantErr bool
	}{
		{
			name:    "Valid URL",
			url:     "https://example.com/page.html",
			want:    []Chunk{{Content: "Test content", Source: "https://example.com/page.html"}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("processURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
