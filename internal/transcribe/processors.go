package transcribe

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/speech/apiv1/speechpb"
	"github.com/amanitaverna/go-mp3"
	"github.com/go-audio/wav"
	"github.com/mewkiz/flac"
)

var (
	defaultFactory = &defaultAudioProcessorFactory{}
	retriever      = newAudioInfoRetriever(defaultFactory)
)

var encodingMap = map[string]speechpb.RecognitionConfig_AudioEncoding{
	"WAV":  speechpb.RecognitionConfig_LINEAR16,
	"MP3":  speechpb.RecognitionConfig_MP3,
	"FLAC": speechpb.RecognitionConfig_FLAC,
}

func (f *defaultAudioProcessorFactory) createProcessor(ext string) (audioProcessor, error) {
	switch strings.ToLower(ext) {
	case ".mp3":
		return &mp3Processor{}, nil
	case ".flac":
		return &flacProcessor{}, nil
	case ".wav":
		return &wavProcessor{}, nil
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}

func newAudioInfoRetriever(factory audioProcessorFactory) *audioInfoRetriever {
	return &audioInfoRetriever{
		factory: factory,
	}
}

func getAudioInfo(filename string) (*speechpb.RecognitionConfig, error) {
	info, err := retriever.audioInfo(filename)
	if err != nil {
		return nil, err
	}

	return &speechpb.RecognitionConfig{
		Encoding:          getEncoding(info.format),
		SampleRateHertz:   int32(info.sampleRate),
		AudioChannelCount: int32(info.numChannels),
	}, nil
}

func (r *audioInfoRetriever) audioInfo(filename string) (*audioInfo, error) {
	ext := filepath.Ext(filename)
	processor, err := r.factory.createProcessor(ext)
	if err != nil {
		return nil, err
	}
	return processor.process(filename)
}

func getEncoding(format string) speechpb.RecognitionConfig_AudioEncoding {
	if encoding, ok := encodingMap[format]; ok {
		return encoding
	}
	return speechpb.RecognitionConfig_ENCODING_UNSPECIFIED
}

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
