package speech

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"cloud.google.com/go/speech/apiv1/speechpb"
)

type mockProcessor struct {
	info *audioInfo
	err  error
}

func (m *mockProcessor) process(string) (*audioInfo, error) {
	return m.info, m.err
}

type mockFactory struct {
	processor audioProcessor
	err       error
}

func (m *mockFactory) createProcessor(string) (audioProcessor, error) {
	return m.processor, m.err
}

func getRootPath() string {
	pwd, _ := os.Getwd()
	pwd = filepath.Dir(filepath.Dir(pwd))
	return pwd
}

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

func Test_audioInfoRetriever_audioInfo(t *testing.T) {
	tests := []struct {
		name     string
		factory  audioProcessorFactory
		filename string
		want     *audioInfo
		wantErr  bool
		errMsg   string
	}{
		{
			name: "MP3 file - successful processing",
			factory: &mockFactory{
				processor: &mockProcessor{
					info: &audioInfo{
						sampleRate:  44100,
						numChannels: 2,
						format:      "MP3",
					},
					err: nil,
				},
				err: nil,
			},
			filename: "test.mp3",
			want: &audioInfo{
				sampleRate:  44100,
				numChannels: 2,
				format:      "MP3",
			},
			wantErr: false,
		},
		{
			name: "FLAC file - successful processing",
			factory: &mockFactory{
				processor: &mockProcessor{
					info: &audioInfo{
						sampleRate:  96000,
						numChannels: 2,
						format:      "FLAC",
					},
					err: nil,
				},
				err: nil,
			},
			filename: "test.flac",
			want: &audioInfo{
				sampleRate:  96000,
				numChannels: 2,
				format:      "FLAC",
			},
			wantErr: false,
		},
		{
			name: "WAV file - processing error",
			factory: &mockFactory{
				processor: &mockProcessor{
					info: nil,
					err:  errors.New("failed to process WAV file"),
				},
				err: nil,
			},
			filename: "test.wav",
			want:     nil,
			wantErr:  true,
			errMsg:   "failed to process WAV file",
		},
		{
			name: "Unsupported file format",
			factory: &mockFactory{
				processor: nil,
				err:       errors.New("unsupported file format: .aac"),
			},
			filename: "test.aac",
			want:     nil,
			wantErr:  true,
			errMsg:   "unsupported file format: .aac",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newAudioInfoRetriever(tt.factory)
			got, err := r.audioInfo(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("audioInfoRetriever.audioInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("audioInfoRetriever.audioInfo() error = %v, expected error %v", err, tt.errMsg)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("audioInfoRetriever.audioInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newAudioInfoRetriever(t *testing.T) {
	factory := &defaultAudioProcessorFactory{}
	tests := []struct {
		name string
		want *audioInfoRetriever
	}{
		{
			name: "Create new audioInfoRetriever",
			want: &audioInfoRetriever{
				factory: factory,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newAudioInfoRetriever(factory); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newAudioInfoRetriever() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getAudioInfo(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    *speechpb.RecognitionConfig
		wantErr bool
	}{
		{
			name: "MP3 file",
			args: args{
				filename: getRootPath() + "/testdata/audio/test.mp3",
			},
			want: &speechpb.RecognitionConfig{
				Encoding:          speechpb.RecognitionConfig_MP3,
				SampleRateHertz:   44100,
				AudioChannelCount: 2,
			},
			wantErr: false,
		},
		{
			name: "FLAC file",
			args: args{
				filename: getRootPath() + "/testdata/audio/test.flac",
			},
			want: &speechpb.RecognitionConfig{
				Encoding:          speechpb.RecognitionConfig_FLAC,
				SampleRateHertz:   96000,
				AudioChannelCount: 2,
			},
			wantErr: false,
		},
		{
			name: "WAV file",
			args: args{
				filename: getRootPath() + "/testdata/audio/test.wav",
			},
			want: &speechpb.RecognitionConfig{
				Encoding:          speechpb.RecognitionConfig_LINEAR16,
				SampleRateHertz:   44100,
				AudioChannelCount: 2,
			},
			wantErr: false,
		},
		{
			name: "Unsupported file format",
			args: args{
				filename: getRootPath() + "/testdata/audio/test.ogg",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getAudioInfo(tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAudioInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getAudioInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
