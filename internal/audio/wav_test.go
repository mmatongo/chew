package audio

import (
	"reflect"
	"testing"
)

func Test_wavProcessor_process(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		p       *wavProcessor
		args    args
		want    *audioInfo
		wantErr bool
	}{
		{
			name: "success",
			p:    &wavProcessor{},
			args: args{
				filename: getRootPath() + "/testdata/audio/test.wav",
			},
			want: &audioInfo{
				sampleRate:  44100,
				numChannels: 2,
				format:      "WAV",
			},
		},
		{
			name: "file not found",
			p:    &wavProcessor{},
			args: args{
				filename: getRootPath() + "/testdata/audio/test_new.wav",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid WAV file",
			p:    &wavProcessor{},
			args: args{
				filename: getRootPath() + "/testdata/audio/test.flac",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &wavProcessor{}
			got, err := p.process(tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("wavProcessor.process() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("wavProcessor.process() = %v, want %v", got, tt.want)
			}
		})
	}
}
