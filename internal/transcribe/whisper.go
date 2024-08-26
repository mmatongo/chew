package transcribe

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func processWhisper(ctx context.Context, filename string, opts TranscribeOptions, client httpClient, opener fileOpener) (string, error) {
	if client == nil {
		client = &http.Client{}
	}
	if opener == nil {
		opener = func(name string) (io.ReadCloser, error) {
			return os.Open(name)
		}
	}

	file, err := opener(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			err = fmt.Errorf("failed to close file: %v (original error: %w)", cerr, err)
		}
	}()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err = io.Copy(part, file); err != nil {
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}

	if err = writeFields(writer, opts); err != nil {
		return "", err
	}

	if err = writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/audio/transcriptions", body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+opts.WhisperAPIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = fmt.Errorf("failed to close response body: %v (original error: %w)", cerr, err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Text string `json:"text"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Text, nil
}

func writeFields(writer *multipart.Writer, opts TranscribeOptions) error {
	fields := map[string]string{
		"model":    opts.WhisperModel,
		"language": opts.LanguageCode,
		"prompt":   opts.WhisperPrompt,
	}

	for key, value := range fields {
		if value != "" {
			if err := writer.WriteField(key, value); err != nil {
				return fmt.Errorf("failed to write %s field: %w", key, err)
			}
		}
	}

	return nil
}

func (wt *whisperTranscriber) process(ctx context.Context, filename string, opts TranscribeOptions) (string, error) {
	return processWhisper(ctx, filename, opts, nil, nil)
}
