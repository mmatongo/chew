package speech

import (
	"reflect"
	"testing"

	"cloud.google.com/go/speech/apiv1/speechpb"
)

func Test_getEncoding(t *testing.T) {
	tests := []struct {
		name   string
		format string
		want   speechpb.RecognitionConfig_AudioEncoding
	}{
		{
			name:   "WAV format",
			format: "WAV",
			want:   speechpb.RecognitionConfig_LINEAR16,
		},
		{
			name:   "MP3 format",
			format: "MP3",
			want:   speechpb.RecognitionConfig_MP3,
		},
		{
			name:   "FLAC format",
			format: "FLAC",
			want:   speechpb.RecognitionConfig_FLAC,
		},
		{
			name:   "Unsupported format",
			format: "AAC",
			want:   speechpb.RecognitionConfig_ENCODING_UNSPECIFIED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getEncoding(tt.format); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getEncoding() = %v, want %v", got, tt.want)
			}
		})
	}
}
