package audio

import (
	"fmt"

	"github.com/mewkiz/flac"
)

type flacProcessor struct{}

func (p *flacProcessor) process(filename string) (*audioInfo, error) {
	file, err := flac.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open FLAC file: %w", err)
	}
	defer func(file *flac.Stream) {
		err := file.Close()
		if err != nil {
			fmt.Printf("failed to close FLAC file: %v\n", err)
		}
	}(file)

	return &audioInfo{
		sampleRate:  int(file.Info.SampleRate),
		numChannels: int(file.Info.NChannels),
		format:      "FLAC",
	}, nil
}
