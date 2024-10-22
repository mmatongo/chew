package utils

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func createMockZipFile(content string) *zip.File {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	var files = []struct {
		Name, Body string
	}{
		{"document.xml", content},
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

	r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		panic(err)
	}

	return r.File[0]
}

func TestRemoveMarkdownSyntax(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test 1",
			args: args{
				text: "This is a **bold** text",
			},
			want: "This is a bold text",
		},
		{
			name: "Test 2",
			args: args{
				text: "This is a *italic* text",
			},
			want: "This is a italic text",
		},
		{
			name: "Test 3",
			args: args{
				text: "This is a [link](https://example.com) text",
			},
			want: "This is a link (https://example.com) text",
		},
		{
			name: "Test 4",
			args: args{
				text: "This is a ![image](https://example.com/image.png) text",
			},
			want: "This is a image (https://example.com/image.png) text",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveMarkdownSyntax(tt.args.text); got != tt.want {
				t.Errorf("RemoveMarkdownSyntax() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFileExtension(t *testing.T) {
	type args struct {
		rawUrl string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test 1",
			args: args{
				rawUrl: "https://example.com/test.csv",
			},
			want:    ".csv",
			wantErr: false,
		},
		{
			name: "Test 2",
			args: args{
				rawUrl: "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Test 3",
			args: args{
				rawUrl: "https://example.com/test",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Test 4",
			args: args{
				rawUrl: "file:///test.csv",
			},
			want:    ".csv",
			wantErr: false,
		},
		{
			name: "Test 5",
			args: args{
				rawUrl: "file:///test",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Test 6",
			args: args{
				rawUrl: string([]byte{0x01, 0x02, 0x03, 0x04, 0x05}),
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFileExtension(tt.args.rawUrl)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFileExtensionFromUrl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetFileExtensionFromUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractTextFromXML(t *testing.T) {
	type args struct {
		file *zip.File
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "valid XML with paragraphs",
			args: args{
				file: createMockZipFile(`
					<?xml version="1.0" encoding="UTF-8"?>
					<document>
						<p>First paragraph</p>
						<p>Second paragraph</p>
						<p>Third paragraph</p>
					</document>
				`),
			},
			want:    []string{"First paragraph", "Second paragraph", "Third paragraph"},
			wantErr: false,
		},
		{
			name: "XML with empty paragraphs",
			args: args{
				file: createMockZipFile(`
					<?xml version="1.0" encoding="UTF-8"?>
					<document>
						<p>First paragraph</p>
						<p></p>
						<p>Third paragraph</p>
					</document>
				`),
			},
			want:    []string{"First paragraph", "Third paragraph"},
			wantErr: false,
		},
		{
			name: "invalid XML",
			args: args{
				file: createMockZipFile(`
					<?xml version="1.0" encoding="UTF-8"?>
					<document>
						<p>Unclosed paragraph
					</document>
				`),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractTextFromXML(tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractTextFromXML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractTextFromXML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOpenFile(t *testing.T) {
	type args struct {
		filePath string
	}
	tests := []struct {
		name    string
		args    args
		want    *os.File
		wantErr bool
	}{
		{
			name: "valid file",
			args: args{
				filePath: "testdata/test.pdf",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := OpenFile(tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("OpenFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OpenFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFileContentType(t *testing.T) {
	tempDir := t.TempDir()
	testHTMLPath := filepath.Join(tempDir, "test.html")

	err := os.WriteFile(testHTMLPath, []byte("html content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test html file: %v", err)
	}

	filepath, _ := OpenFile(testHTMLPath)
	type args struct {
		file *os.File
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test 1",
			args: args{
				file: filepath,
			},
			want: "text/html; charset=utf-8",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetFileContentType(tt.args.file); got != tt.want {
				t.Errorf("GetFileContentType() = %v, want %v", got, tt.want)
			}
		})
	}
}
