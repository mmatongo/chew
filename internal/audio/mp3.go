package audio

import (
	"fmt"
	"os"

	"github.com/amanitaverna/go-mp3"
)

type mp3Processor struct{}

func (p *mp3Processor) process(filename string) (*audioInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open MP3 file: %w", err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("failed to close MP3 file: %v\n", err)
		}
	}(file)

	decoder, err := mp3.NewDecoder(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create MP3 decoder: %w", err)
	}

	return &audioInfo{
		sampleRate: decoder.SampleRate(),
		/*
			This is a terrible assumption but seeing as the MP3 decoder
			doesn't expose this information, we'll have to live with it for now.
		*/
		numChannels: 2,
		format:      "MP3",
	}, nil
}
