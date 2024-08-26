package audio

type audioInfo struct {
	sampleRate  int
	numChannels int
	format      string
}

type audioProcessor interface {
	process(filename string) (*audioInfo, error)
}

type audioProcessorFactory interface {
	createProcessor(fileExtension string) (audioProcessor, error)
}

type defaultAudioProcessorFactory struct{}

type audioInfoRetriever struct {
	factory audioProcessorFactory
}
