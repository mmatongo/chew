package audio

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-audio/wav"
)

type wavProcessor struct{}

func (p *wavProcessor) process(filename string) (*audioInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open WAV file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("failed to close WAV file: %v\n", err)
		}
	}(file)

	decoder := wav.NewDecoder(file)
	if !decoder.IsValidFile() {
		return nil, errors.New("invalid WAV file")
	}

	return &audioInfo{
		sampleRate:  int(decoder.SampleRate),
		numChannels: int(decoder.NumChans),
		format:      "WAV",
	}, nil
}
