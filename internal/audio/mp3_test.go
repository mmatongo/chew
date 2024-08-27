package audio

import (
	"reflect"
	"testing"
)

func Test_mp3Processor_process(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		p       *mp3Processor
		args    args
		want    *audioInfo
		wantErr bool
	}{
		{
			name: "success",
			p:    &mp3Processor{},
			args: args{
				filename: getRootPath(t) + "/testdata/audio/test.mp3",
			},
			want: &audioInfo{
				sampleRate:  44100,
				numChannels: 2,
				format:      "MP3",
			},
			wantErr: false,
		},
		{
			name: "file not found",
			p:    &mp3Processor{},
			args: args{
				filename: getRootPath(t) + "/testdata/audio/test_new.mp3",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid MP3 file",
			p:    &mp3Processor{},
			args: args{
				filename: getRootPath(t) + "/testdata/audio/test.flac",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &mp3Processor{}
			got, err := p.process(tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("mp3Processor.process() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mp3Processor.process() = %v, want %v", got, tt.want)
			}
		})
	}
}
