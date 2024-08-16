package utils

import (
	"archive/zip"
	"bytes"
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

func TestGetFileExtensionFromUrl(t *testing.T) {
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFileExtensionFromUrl(tt.args.rawUrl)
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
