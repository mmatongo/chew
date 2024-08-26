package common

type TranscribeOptions struct {
	CredentialsJSON   []byte
	Bucket            string
	LanguageCode      string
	EnableDiarization bool
	MinSpeakers       int
	MaxSpeakers       int
	CleanupOnComplete bool
	UseWhisper        bool
	WhisperAPIKey     string
	WhisperModel      string
	WhisperPrompt     string
}

type Chunk struct {
	Content string
	Source  string
}
