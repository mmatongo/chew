package markdown

import "testing"

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
