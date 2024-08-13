package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/mmatongo/chew"
)

func main() {
	key := os.Getenv("OPEN_AI_API_KEY")
	if key == "" {
		log.Fatalf("Please set the OPEN_AI_API_KEY environment variable")
	}

	whisperOpts := chew.TranscribeOptions{
		UseWhisper:    true,
		WhisperAPIKey: key,
		WhisperModel:  "whisper-1",
	}

	log.Println("transcribing files...")
	/*
		The whisper model is a bit faster than the google cloud speech-to-text api
		so the timeout can be set to a lower value.

		In a test with MLK Jr's transcribe it took about 32s to complete

		The two audio files used in this example can be obtained from the following links:
		- Conference.wav: https://storage.googleapis.com/chew-go/audio/Conference.wav
		- MLKDream_64kb.mp3: https://voiceage.com/wbsamples/in_stereo/Conference.wav
	*/

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	audioFiles := []string{
		"audio/Conference.wav",
		"audio/MLKDream_64kb.mp3",
	}

	results, err := chew.Transcribe(ctx, audioFiles, whisperOpts)

	if err != nil {
		log.Fatalf("Error transcribing with OpenAI Whisper: %v", err)
	}

	for filename, transcript := range results {
		log.Printf("Transcript for %s: %s\n", filename, transcript)
	}
}
