package transcribe

import (
	speech "cloud.google.com/go/speech/apiv1"
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"path/filepath"
)

type googleTranscriber struct{}

func (gt *googleTranscriber) process(ctx context.Context, filename string, opts TranscribeOptions) (string, error) {
	client, err := createSpeechClient(ctx, opts)
	if err != nil {
		return "", err
	}
	defer func(client *speech.Client) {
		err := client.Close()
		if err != nil {
			fmt.Printf("failed to close transcribe client: %v\n", err)
		}
	}(client)

	storageClient, err := createStorageClient(ctx, opts)
	if err != nil {
		return "", err
	}
	defer func(storageClient *storage.Client) {
		err := storageClient.Close()
		if err != nil {
			fmt.Printf("failed to close storage client: %v\n", err)
		}
	}(storageClient)

	audioInfo, err := getAudioInfo(filename)
	if err != nil {
		return "", fmt.Errorf("failed to process audio file: %w", err)
	}

	gcsURI, err := uploadToGCS(ctx, storageClient, opts.Bucket, filename)
	if err != nil {
		return "", fmt.Errorf("failed to upload to GCS: %w", err)
	}

	if opts.CleanupOnComplete {
		defer func(ctx context.Context, client *storage.Client, bucket, objectName string) {
			err := deleteFromGCS(ctx, client, bucket, objectName)
			if err != nil {
				fmt.Printf("failed to delete object from GCS: %v\n", err)
			}
		}(ctx, storageClient, opts.Bucket, filepath.Base(filename))
	}

	req := createRecognitionRequest(opts, audioInfo, gcsURI)

	op, err := client.LongRunningRecognize(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to start long running recognition: %w", err)
	}

	resp, err := op.Wait(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get long running recognition results: %w", err)
	}

	return extractTranscript(resp), nil
}
