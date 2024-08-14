package transcribe

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

type mockTranscriber struct {
	processFn func(ctx context.Context, filename string, opts TranscribeOptions) (string, error)
}

func (m *mockTranscriber) process(ctx context.Context, filename string, opts TranscribeOptions) (string, error) {
	if m.processFn != nil {
		return m.processFn(ctx, filename, opts)
	}
	return "", nil
}

func TestTranscribe(t *testing.T) {
	type args struct {
		ctx       context.Context
		filenames []string
		opts      TranscribeOptions
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
		mockFn  func(ctx context.Context, filename string, opts TranscribeOptions) (string, error)
	}{
		{
			name: "Test Transcribe",
			args: args{
				ctx:       context.Background(),
				filenames: []string{"test1.mp3", "test2.mp3"},
				opts: TranscribeOptions{
					CredentialsJSON:   []byte("``"),
					Bucket:            "test-bucket",
					LanguageCode:      "en-US",
					EnableDiarization: false,
					MinSpeakers:       0,
					MaxSpeakers:       0,
					CleanupOnComplete: false,
					UseWhisper:        false,
					WhisperAPIKey:     "",
					WhisperModel:      "",
					WhisperPrompt:     "",
				},
			},
			want: map[string]string{
				"test1.mp3": "transcript for test1.mp3",
				"test2.mp3": "transcript for test2.mp3",
			},
			wantErr: false,
			mockFn: func(ctx context.Context, filename string, opts TranscribeOptions) (string, error) {
				return "transcript for " + filename, nil
			},
		},
		{
			name: "Test Transcribe Error",
			args: args{
				ctx:       context.Background(),
				filenames: []string{"test1.mp3", "test2.mp3"},
				opts:      TranscribeOptions{},
			},
			want:    nil,
			wantErr: true,
			mockFn: func(ctx context.Context, filename string, opts TranscribeOptions) (string, error) {
				return "", fmt.Errorf("mock error")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockT := &mockTranscriber{
				processFn: tt.mockFn,
			}
			got, err := Transcribe(tt.args.ctx, tt.args.filenames, tt.args.opts, WithTranscriber(mockT))
			if (err != nil) != tt.wantErr {
				t.Errorf("Transcribe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Transcribe() = %v, want %v", got, tt.want)
			}
		})
	}
}
