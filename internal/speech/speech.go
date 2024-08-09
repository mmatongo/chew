package speech

import (
	"context"
	"fmt"

	speech "cloud.google.com/go/speech/apiv1"
	"cloud.google.com/go/speech/apiv1/speechpb"
	"google.golang.org/api/option"
)

type TranscribeOptions struct {
	CredentialsJSON   []byte
	Bucket            string
	LanguageCode      string
	EnableDiarization bool
	MinSpeakers       int
	MaxSpeakers       int
}

// code is largely inspired by https://github.com/polyfact/polyfire-api

/*
Transcribe uses the Google Cloud Speech-to-Text API to transcribe an audio file. It takes
a context, the filename of the audio file to transcribe, and a TranscribeOptions struct which
contains the Google Cloud credentials, the GCS bucket to upload the audio file to, and the language code
to use for transcription. It returns the transcript of the audio file as a string and an error if the
transcription fails.
*/

func Transcribe(ctx context.Context, filename string, opts TranscribeOptions) (string, error) {
	var clientOpts []option.ClientOption
	if opts.CredentialsJSON != nil {
		clientOpts = append(clientOpts, option.WithCredentialsJSON(opts.CredentialsJSON))
	}

	client, err := speech.NewClient(ctx, clientOpts...)
	if err != nil {
		return "", fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	audioInfo, err := getAudioInfo(filename)
	if err != nil {
		return "", fmt.Errorf("failed to process audio file: %v", err)
	}

	gcsURI, err := uploadToGCS(ctx, opts.Bucket, filename)
	if err != nil {
		return "", fmt.Errorf("failed to upload to GCS: %v", err)
	}

	diarizationConfig := &speechpb.SpeakerDiarizationConfig{
		EnableSpeakerDiarization: opts.EnableDiarization,
		MinSpeakerCount:          int32(opts.MinSpeakers),
		MaxSpeakerCount:          int32(opts.MaxSpeakers),
	}

	req := &speechpb.LongRunningRecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			Encoding:                   getEncoding(audioInfo.format),
			SampleRateHertz:            int32(audioInfo.sampleRate),
			AudioChannelCount:          int32(audioInfo.numChannels),
			LanguageCode:               opts.LanguageCode,
			EnableAutomaticPunctuation: true,
			UseEnhanced:                true,
			EnableWordConfidence:       true,
			Model:                      "latest_long",
			DiarizationConfig:          diarizationConfig,
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Uri{
				Uri: gcsURI,
			},
		},
	}

	op, err := client.LongRunningRecognize(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to start long running recognition: %v", err)
	}

	resp, err := op.Wait(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get long running recognition results: %v", err)
	}

	var transcript string
	for _, result := range resp.Results {
		for _, alt := range result.Alternatives {
			transcript += alt.Transcript
		}
	}

	return transcript, nil
}
