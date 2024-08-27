package transcribe

import (
	"context"
	"fmt"
	"sync"

	"github.com/mmatongo/chew/v1/internal/common"
)

/*
The TranscribeOptions struct contains the options for transcribing an audio file. It allows the user
to specify the Google Cloud credentials, the GCS bucket to upload the audio file to, the language code
to use for transcription, a potion to enable diarization including the min and max speakers and
an option to clean up the audio file from GCS after transcription is complete.

And also, it allows the user to specify whether to use the Whisper API for transcription, and if so,
the API key, model, and prompt to use.
*/
type TranscribeOptions = common.TranscribeOptions

// code is largely inspired by https://github.com/polyfact/polyfire-api

type transcribeOption func(*transcribeConfig)

type transcribeConfig struct {
	t transcriber
}

func WithTranscriber(t transcriber) transcribeOption {
	return func(config *transcribeConfig) {
		config.t = t
	}
}

/*
Transcribe uses the Google Cloud Speech-to-Text API to transcribe an audio file. It takes
a context, the filename of the audio file to transcribe, and a TranscribeOptions struct which
contains the Google Cloud credentials, the GCS bucket to upload the audio file to, the language code
to use for transcription, a potion to enable diarization including the min and max speakers and
an option to clean up the audio file from GCS after transcription is complete.
It returns the transcript of the audio file as a string and an error if the transcription fails.
*/
func Transcribe(ctx context.Context, filenames []string, opts TranscribeOptions, options ...transcribeOption) (map[string]string, error) {
	config := &transcribeConfig{}
	for _, option := range options {
		option(config)
	}

	if config.t == nil {
		if opts.UseWhisper {
			config.t = &whisperTranscriber{}
		} else {
			config.t = &googleTranscriber{}
		}
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

			transcript, err := config.t.process(ctx, filename, opts)
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
