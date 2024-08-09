package speech

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
)

func uploadToGCS(ctx context.Context, bucket, filename string) (string, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create storage client: %v", err)
	}
	defer client.Close()

	f, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer f.Close()

	objectName := filepath.Base(filename)
	w := client.Bucket(bucket).Object(objectName).NewWriter(ctx)
	if _, err = io.Copy(w, f); err != nil {
		return "", fmt.Errorf("failed to copy file to GCS: %v", err)
	}
	if err := w.Close(); err != nil {
		return "", fmt.Errorf("failed to close GCS writer: %v", err)
	}

	return fmt.Sprintf("gs://%s/%s", bucket, objectName), nil
}