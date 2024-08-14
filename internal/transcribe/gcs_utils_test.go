package transcribe

import (
	"context"
	"os"
	"reflect"
	"testing"

	speech "cloud.google.com/go/speech/apiv1"
	"cloud.google.com/go/speech/apiv1/speechpb"
	"cloud.google.com/go/storage"
)

func Test_extractTranscript(t *testing.T) {
	type args struct {
		resp *speechpb.LongRunningRecognizeResponse
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty response",
			args: args{
				resp: &speechpb.LongRunningRecognizeResponse{},
			},
			want: "",
		},
		{
			name: "response with no results",
			args: args{
				resp: &speechpb.LongRunningRecognizeResponse{
					Results: []*speechpb.SpeechRecognitionResult{},
				},
			},
			want: "",
		},
		{
			name: "response with no alternatives",
			args: args{
				resp: &speechpb.LongRunningRecognizeResponse{
					Results: []*speechpb.SpeechRecognitionResult{
						{},
					},
				},
			},
		},
		{
			name: "response with result and alternative",
			args: args{
				resp: &speechpb.LongRunningRecognizeResponse{
					Results: []*speechpb.SpeechRecognitionResult{
						{
							Alternatives: []*speechpb.SpeechRecognitionAlternative{
								{
									Transcript: "hello world",
								},
							},
						},
					},
				},
			},
			want: "hello world",
		},
		{
			name: "response with multiple results and alternatives",
			args: args{
				resp: &speechpb.LongRunningRecognizeResponse{
					Results: []*speechpb.SpeechRecognitionResult{
						{
							Alternatives: []*speechpb.SpeechRecognitionAlternative{
								{
									Transcript: "hello world",
									Confidence: 0.9,
								},
								{
									Transcript: "hello world",
									Confidence: 0.8,
								},
							},
						},
						{
							Alternatives: []*speechpb.SpeechRecognitionAlternative{
								{
									Transcript: "hello world",
									Confidence: 0.7,
								},
								{
									Transcript: "hello world",
									Confidence: 0.6,
								},
							},
						},
					},
				},
			},
			want: "hello worldhello worldhello worldhello world",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractTranscript(tt.args.resp); got != tt.want {
				t.Errorf("extractTranscript() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createRecognitionRequest(t *testing.T) {
	type args struct {
		opts      TranscribeOptions
		audioInfo *speechpb.RecognitionConfig
		gcsURI    string
	}
	tests := []struct {
		name string
		args args
		want *speechpb.LongRunningRecognizeRequest
	}{
		{
			name: "create recognition request",
			args: args{
				opts: TranscribeOptions{
					EnableDiarization: true,
					MinSpeakers:       1,
					MaxSpeakers:       2,
					LanguageCode:      "en-US",
				},
				audioInfo: &speechpb.RecognitionConfig{
					Encoding:          speechpb.RecognitionConfig_ENCODING_UNSPECIFIED,
					SampleRateHertz:   44100,
					AudioChannelCount: 2,
				},
				gcsURI: "gs://bucket/object",
			},
			want: &speechpb.LongRunningRecognizeRequest{
				Config: &speechpb.RecognitionConfig{
					Encoding:                   speechpb.RecognitionConfig_ENCODING_UNSPECIFIED,
					SampleRateHertz:            44100,
					AudioChannelCount:          2,
					LanguageCode:               "en-US",
					EnableAutomaticPunctuation: true,
					UseEnhanced:                true,
					EnableWordConfidence:       true,
					Model:                      "latest_long",
					DiarizationConfig: &speechpb.SpeakerDiarizationConfig{
						EnableSpeakerDiarization: true,
						MinSpeakerCount:          1,
						MaxSpeakerCount:          2,
					},
				},
				Audio: &speechpb.RecognitionAudio{
					AudioSource: &speechpb.RecognitionAudio_Uri{
						Uri: "gs://bucket/object",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createRecognitionRequest(tt.args.opts, tt.args.audioInfo, tt.args.gcsURI); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createRecognitionRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

/*
All of the following tests are expected to fail because the credentials JSON is empty
and the functions are not written in a way that allows for mocking of the GCP client libraries.
This is a limitation of the current implementation and should be refactored in the future.
*/

func Test_createSpeechClient(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts TranscribeOptions
	}
	tests := []struct {
		name    string
		args    args
		want    *speech.Client
		wantErr bool
	}{
		{
			name: "create speech client",
			args: args{
				ctx: context.Background(),
				opts: TranscribeOptions{
					CredentialsJSON: nil,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			/*
				This test case is expected to fail because the credentials JSON is empty.

				TODO: Refactor to allow for mocking of the speech.NewClient function.
			*/
			name: "create speech client with credentials",
			args: args{
				ctx: context.Background(),
				opts: TranscribeOptions{
					CredentialsJSON: []byte(""),
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createSpeechClient(tt.args.ctx, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("createSpeechClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createSpeechClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createStorageClient(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts TranscribeOptions
	}
	tests := []struct {
		name    string
		args    args
		want    *storage.Client
		wantErr bool
	}{
		{
			name: "create storage client",
			args: args{
				ctx: context.Background(),
				opts: TranscribeOptions{
					CredentialsJSON: nil,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			/*
				This test case is expected to fail because the credentials JSON is empty.
				This does not affect the functionality of the createStorageClient function.

				TODO: Refactor to allow for mocking of the storage.NewClient function.
			*/

			name: "create storage client with credentials",
			args: args{
				ctx: context.Background(),
				opts: TranscribeOptions{
					CredentialsJSON: []byte(""),
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createStorageClient(tt.args.ctx, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("createStorageClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createStorageClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_uploadToGCS(t *testing.T) {
	store, _ := storage.NewClient(context.Background())

	type args struct {
		ctx      context.Context
		client   *storage.Client
		bucket   string
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "upload to GCS",
			args: args{
				ctx:      context.Background(),
				client:   nil,
				bucket:   "bucket",
				filename: "file",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "upload to GCS with empty filename",
			args: args{
				ctx:      context.Background(),
				client:   store,
				bucket:   "bucket",
				filename: "",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "upload to GCS with real filename of mock file" {
				defer os.Remove(tt.args.filename)
			}

			got, err := uploadToGCS(tt.args.ctx, tt.args.client, tt.args.bucket, tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("uploadToGCS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("uploadToGCS() = %v, want %v", got, tt.want)
			}
		})
	}
}
