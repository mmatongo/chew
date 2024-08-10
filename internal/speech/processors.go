package speech

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/speech/apiv1/speechpb"
	"github.com/amanitaverna/go-mp3"
	"github.com/go-audio/wav"
	"github.com/mewkiz/flac"
)

type audioInfo struct {
	sampleRate  int
	numChannels int
	bitDepth    int
	format      string
}

func getAudioInfo(filename string) (*audioInfo, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".mp3":
		return processMp3(filename)
	case ".flac":
		return processFlac(filename)
	case ".wav":
		return processWav(filename)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}

func getEncoding(format string) speechpb.RecognitionConfig_AudioEncoding {
	switch format {
	case "WAV":
		return speechpb.RecognitionConfig_LINEAR16
	case "MP3":
		return speechpb.RecognitionConfig_MP3
	case "FLAC":
		return speechpb.RecognitionConfig_FLAC
	default:
		return speechpb.RecognitionConfig_ENCODING_UNSPECIFIED
	}
}

func processMp3(filename string) (*audioInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open MP3 file: %w", err)
	}

	defer file.Close()

	decoder, err := mp3.NewDecoder(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create MP3 decoder: %w", err)
	}

	return &audioInfo{
		sampleRate:  decoder.SampleRate(),
		numChannels: 2,
		bitDepth:    0, // http://blog.bjrn.se/2008/10/lets-build-mp3-decoder.html
		format:      "MP3",
	}, nil
}

func processFlac(filename string) (*audioInfo, error) {
	file, err := flac.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open FLAC file: %w", err)
	}
	defer file.Close()

	return &audioInfo{
		sampleRate:  int(file.Info.SampleRate),
		numChannels: int(file.Info.NChannels),
		bitDepth:    int(file.Info.BitsPerSample),
		format:      "FLAC",
	}, nil
}

func processWav(filename string) (*audioInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open WAV file: %w", err)
	}
	defer file.Close()

	decoder := wav.NewDecoder(file)
	if !decoder.IsValidFile() {
		return nil, errors.New("invalid WAV file")
	}

	return &audioInfo{
		sampleRate:  int(decoder.SampleRate),
		numChannels: int(decoder.NumChans),
		bitDepth:    int(decoder.BitDepth),
		format:      "WAV",
	}, nil
}
