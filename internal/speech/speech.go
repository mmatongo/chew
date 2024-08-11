package speech

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	speech "cloud.google.com/go/speech/apiv1"
	"cloud.google.com/go/speech/apiv1/speechpb"
	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type TranscribeOptions struct {
	CredentialsJSON   []byte
	Bucket            string
	LanguageCode      string
	EnableDiarization bool
	MinSpeakers       int
	MaxSpeakers       int
	CleanupOnComplete bool
}

// code is largely inspired by https://github.com/polyfact/polyfire-api

/*
Transcribe uses the Google Cloud Speech-to-Text API to transcribe an audio file. It takes
a context, the filename of the audio file to transcribe, and a TranscribeOptions struct which
contains the Google Cloud credentials, the GCS bucket to upload the audio file to, the language code
to use for transcription, a potion to enable diarization including the min and max speakers and
an option to clean up the audio file from GCS after transcription is complete.
It returns the transcript of the audio file as a string and an error if the transcription fails.
*/
func Transcribe(ctx context.Context, filenames []string, opts TranscribeOptions) (map[string]string, error) {
	var (
		results = make(map[string]string)
		wg      sync.WaitGroup
		mu      sync.Mutex
		errCh   = make(chan error, len(filenames))
	)

	for _, filename := range filenames {
		wg.Add(1)
		go func(filename string) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			default:
				transcript, err := processFile(ctx, filename, opts)
				if err != nil {
					errCh <- fmt.Errorf("transcribing %s: %w", filename, err)
					return
				}
				mu.Lock()
				results[filename] = transcript
				mu.Unlock()
			}
		}(filename)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return nil, err
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return results, nil
}

func processFile(ctx context.Context, filename string, opts TranscribeOptions) (string, error) {
	var clientOpts []option.ClientOption
	if opts.CredentialsJSON != nil {
		clientOpts = append(clientOpts, option.WithCredentialsJSON(opts.CredentialsJSON))
	}

	client, err := speech.NewClient(ctx, clientOpts...)
	if err != nil {
		return "", fmt.Errorf("failed to create client: %w", err)
	}
	defer func(client *speech.Client) {
		err := client.Close()
		if err != nil {
			fmt.Printf("failed to close client: %v\n", err)
		}
	}(client)

	storageClient, err := storage.NewClient(ctx, clientOpts...)
	if err != nil {
		return "", fmt.Errorf("failed to create storage client: %w", err)
	}
	defer func(storageClient *storage.Client) {
		err := storageClient.Close()
		if err != nil {
			fmt.Printf("failed to close storage client: %v\n", err)
		}
	}(storageClient)

	audioInfo, err := getAudioInfo(filename)
	if err != nil {
		return "", fmt.Errorf("failed to process audio file: %e", err)
	}

	gcsURI, err := uploadToGCS(ctx, storageClient, opts.Bucket, filename)
	if err != nil {
		return "", fmt.Errorf("failed to upload to GCS: %e", err)
	}

	defer func() {
		if opts.CleanupOnComplete {
			objectName := filepath.Base(filename)
			err := deleteFromGCS(ctx, storageClient, opts.Bucket, objectName)
			if err != nil {
				fmt.Printf("failed to delete from GCS: %v\n", err)
			}
		}
	}()

	diarizationConfig := &speechpb.SpeakerDiarizationConfig{
		EnableSpeakerDiarization: opts.EnableDiarization,
		MinSpeakerCount:          int32(opts.MinSpeakers),
		MaxSpeakerCount:          int32(opts.MaxSpeakers),
	}

	req := &speechpb.LongRunningRecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			Encoding:                   audioInfo.Encoding,
			SampleRateHertz:            audioInfo.SampleRateHertz,
			AudioChannelCount:          int32(audioInfo.AudioChannelCount),
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
		return "", fmt.Errorf("failed to start long running recognition: %w", err)
	}

	resp, err := op.Wait(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get long running recognition results: %w", err)
	}

	var transcript string
	for _, result := range resp.Results {
		for _, alt := range result.Alternatives {
			transcript += alt.Transcript
		}
	}

	return transcript, nil
}
