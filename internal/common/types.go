package common

import (
	"time"
)

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

type Config struct {
	UserAgent       string
	RetryLimit      int
	RetryDelay      time.Duration
	CrawlDelay      time.Duration
	ProxyList       []string
	RateLimit       time.Duration
	RateBurst       int
	IgnoreRobotsTxt bool
}

type Chunk struct {
	Content string
	Source  string
}
