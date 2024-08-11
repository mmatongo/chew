package speech

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
)

func uploadToGCS(ctx context.Context, client *storage.Client, bucket, filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Printf("failed to close file: %v\n", err)
		}
	}(f)

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
