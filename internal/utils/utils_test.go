package utils

import "testing"

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
				rawUrl: "https://example.com/big-file.csv",
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
