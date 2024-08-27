package audio

import (
	"fmt"
	"path/filepath"
	"strings"

	"cloud.google.com/go/speech/apiv1/speechpb"
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

func GetAudioInfo(filename string) (*speechpb.RecognitionConfig, error) {
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
