package transcribe

import (
	"context"
	"fmt"
	"sync"
)

/*
The TranscribeOptions struct contains the options for transcribing an audio file. It allows the user
to specify the Google Cloud credentials, the GCS bucket to upload the audio file to, the language code
to use for transcription, a potion to enable diarization including the min and max speakers and
an option to clean up the audio file from GCS after transcription is complete.

And also, it allows the user to specify whether to use the Whisper API for transcription, and if so,
the API key, model, and prompt to use.
*/
type TranscribeOptions struct {
	CredentialsJSON   []byte
	Bucket            string
	LanguageCode      string
	EnableDiarization bool
	MinSpeakers       int
	MaxSpeakers       int
	CleanupOnComplete bool
	UseWhisper        bool
	WhisperAPIKey     string
	WhisperModel      string
	WhisperPrompt     string
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
	var t transcriber
	if opts.UseWhisper {
		t = &whisperTranscriber{}
	} else {
		t = &googleTranscriber{}
	}

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

			transcript, err := t.process(ctx, filename, opts)
			if err != nil {
				select {
				case errCh <- fmt.Errorf("transcribing %s: %w", filename, err):
				default:
				}
				return
			}

			mu.Lock()
			results[filename] = transcript
			mu.Unlock()
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
