//go:build ignore

package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/mmatongo/chew/v1"
)

func main() {
	credentialsFile := "chew-go.json"
	credentialsJSON, err := os.ReadFile(credentialsFile)
	if err != nil {
		log.Fatalf("Failed to read credentials file: %v", err)
	}

	err = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credentialsFile)
	if err != nil {
		log.Fatalf("Failed to set environment variable: %v", err)
	}

	config := chew.TranscribeOptions{
		CredentialsJSON: credentialsJSON,
		Bucket:          "chew-go",
		LanguageCode:    "en-US",
	}

	log.Println("transcribing files...")
	/*
		Transcriptions can take a bit of time so ensure that the timeout you set
		is enough for the process to finish

		In a test with MLK Jr's speech it took about 3min to complete

		The two audio files used in this example can be obtained from the following links:
		- Conference.wav: https://voiceage.com/wbsamples/in_stereo/Conference.wav
		- MLKDream_64kb.mp3: https://archive.org/details/MLKDream
	*/

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	filenames := []string{
		"audio/Conference.wav",
		"audio/MLKDream_64kb.mp3",
	}

	results, err := chew.Transcribe(ctx, filenames, config)
	if err != nil {
		log.Fatalf("failed to transcribe: %v", err)
	}

	for filename, transcript := range results {
		log.Printf("Transcript for %s: %s\n", filename, transcript)
	}
}
