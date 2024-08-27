package transcribe

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"cloud.google.com/go/storage"

	"github.com/mmatongo/chew/internal/audio"
	"github.com/mmatongo/chew/internal/utils/gcs"
)

type googleTranscriber struct{}

/*
This relies too heavily on external dependencies and is not easily testable. A refactor is needed to make it more testable and is currently in progress.
*/
func (gt *googleTranscriber) process(ctx context.Context, filename string, opts TranscribeOptions) (string, error) {
	client, err := gcs.NewSpeechClient(ctx, opts)
	if err != nil {
		return "", fmt.Errorf("failed to create speech client: %w", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("failed to close transcribe client: %w", cerr))
		}
	}()

	storageClient, err := gcs.NewStorageClient(ctx, opts)
	if err != nil {
		return "", err
	}
	defer func(storageClient *storage.Client) {
		err := storageClient.Close()
		if err != nil {
			fmt.Printf("failed to close storage client: %v\n", err)
		}
	}(storageClient)

	audioInfo, err := audio.GetAudioInfo(filename)
	if err != nil {
		return "", fmt.Errorf("failed to process audio file: %w", err)
	}

	gcsURI, err := gcs.UploadToGCS(ctx, storageClient, opts.Bucket, filename)
	if err != nil {
		return "", fmt.Errorf("failed to upload to GCS: %w", err)
	}

	if opts.CleanupOnComplete {
		defer func(ctx context.Context, client *storage.Client, bucket, objectName string) {
			err := gcs.DeleteFromGCS(ctx, client, bucket, objectName)
			if err != nil {
				fmt.Printf("failed to delete object from GCS: %v\n", err)
			}
		}(ctx, storageClient, opts.Bucket, filepath.Base(filename))
	}

	req := gcs.NewRecognitionRequest(opts, audioInfo, gcsURI)

	op, err := client.LongRunningRecognize(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to start long running recognition: %w", err)
	}

	resp, err := op.Wait(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get long running recognition results: %w", err)
	}

	return gcs.ExtractTranscript(resp), nil
}
