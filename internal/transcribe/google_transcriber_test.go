package transcribe

import (
	"context"
	"testing"
)

func Test_googleTranscriber_process(t *testing.T) {
	type args struct {
		ctx      context.Context
		filename string
		opts     TranscribeOptions
	}
	tests := []struct {
		name    string
		gt      *googleTranscriber
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "failed to create speech client",
			gt:   &googleTranscriber{},
			args: args{
				ctx:      context.Background(),
				filename: "test.mp3",
				opts:     TranscribeOptions{},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gt := &googleTranscriber{}
			got, err := gt.process(tt.args.ctx, tt.args.filename, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("googleTranscriber.process() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("googleTranscriber.process() = %v, want %v", got, tt.want)
			}
		})
	}
}
