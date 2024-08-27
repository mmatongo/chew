package audio

import (
	"reflect"
	"testing"
)

func Test_flacProcessor_process(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		p       *flacProcessor
		args    args
		want    *audioInfo
		wantErr bool
	}{
		{
			name: "success",
			p:    &flacProcessor{},
			args: args{
				filename: getRootPath(t) + "/testdata/audio/test.flac",
			},
			want: &audioInfo{
				sampleRate:  96000,
				numChannels: 2,
				format:      "FLAC",
			},
			wantErr: false,
		},
		{
			name: "file not found",
			p:    &flacProcessor{},
			args: args{
				filename: getRootPath(t) + "/testdata/audio/test_new.flac",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &flacProcessor{}
			got, err := p.process(tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("flacProcessor.process() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("flacProcessor.process() = %v, want %v", got, tt.want)
			}
		})
	}
}
