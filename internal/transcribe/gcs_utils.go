package transcribe

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	speech "cloud.google.com/go/speech/apiv1"
	"cloud.google.com/go/speech/apiv1/speechpb"
	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

func uploadToGCS(ctx context.Context, client *storage.Client, bucket, filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("failed to close file: %w", cerr))
		}
	}()

	objectName := filepath.Base(filename)
	w := client.Bucket(bucket).Object(objectName).NewWriter(ctx)
	if _, err = io.Copy(w, f); err != nil {
		return "", fmt.Errorf("failed to copy file to GCS: %w", err)
	}
	if err := w.Close(); err != nil {
		return "", fmt.Errorf("failed to close GCS writer: %w", err)
	}

	return fmt.Sprintf("gs://%s/%s", bucket, objectName), nil
}

func deleteFromGCS(ctx context.Context, client *storage.Client, bucket, objectName string) error {
	if err := client.Bucket(bucket).Object(objectName).Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete object from GCS: %w", err)
	}
	return nil
}

func createStorageClient(ctx context.Context, opts TranscribeOptions) (*storage.Client, error) {
	var clientOpts []option.ClientOption
	if opts.CredentialsJSON != nil {
		clientOpts = append(clientOpts, option.WithCredentialsJSON(opts.CredentialsJSON))
	}
	return storage.NewClient(ctx, clientOpts...)
}

func createSpeechClient(ctx context.Context, opts TranscribeOptions) (*speech.Client, error) {
	var clientOpts []option.ClientOption
	if opts.CredentialsJSON != nil {
		clientOpts = append(clientOpts, option.WithCredentialsJSON(opts.CredentialsJSON))
	}
	return speech.NewClient(ctx, clientOpts...)
}

func createRecognitionRequest(opts TranscribeOptions, audioInfo *speechpb.RecognitionConfig, gcsURI string) *speechpb.LongRunningRecognizeRequest {
	diarizationConfig := &speechpb.SpeakerDiarizationConfig{
		EnableSpeakerDiarization: opts.EnableDiarization,
		MinSpeakerCount:          int32(opts.MinSpeakers),
		MaxSpeakerCount:          int32(opts.MaxSpeakers),
	}

	return &speechpb.LongRunningRecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			Encoding:                   audioInfo.Encoding,
			SampleRateHertz:            audioInfo.SampleRateHertz,
			AudioChannelCount:          audioInfo.AudioChannelCount,
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
}

func extractTranscript(resp *speechpb.LongRunningRecognizeResponse) string {
	var transcript string
	for _, result := range resp.Results {
		for _, alt := range result.Alternatives {
			transcript += alt.Transcript
		}
	}
	return transcript
}
