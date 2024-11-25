/*
Package chew provides a simple way to process URLs and files. It allows you to process a list of URLs
and files, and returns the content of the URLs and files as a list of Chunks. It also provides a way to
transcribe audio files using the Google Cloud Speech-to-Text API or the OpenAI Whisper API.

The library respects rules defined in robots.txt file and crawl delays, and allows you to set a custom http.Client for making requests.

Note on Responsible Usage:

This library is designed for processing data from both local files and web sources. Users should be aware of the following considerations:

1. Web Scraping:
  - When scraping websites, ensure compliance with the target website's terms of service and robots.txt rules.
  - Respect rate limits and crawl delays to avoid overwhelming target servers.
  - Be aware that web scraping may be subject to legal restrictions in some jurisdictions.
  - While the library will attempt to respect robots.txt rules by default, users are responsible for ensuring
    that their usage complies with the target website's terms of service and legal requirements.

2. File Processing:
  - Exercise caution when processing files from untrusted sources.
  - Ensure you have appropriate permissions to access and process the files.
  - Be mindful of potential sensitive information in processed files and handle it securely.

3. Data Handling:
  - Properly secure and manage any data extracted or processed using this library, especially if it contains personal or sensitive information.
  - Comply with relevant data protection regulations (e.g., GDPR, CCPA) when handling personal data.

4. System Resource Usage:
  - Be aware that processing large files or numerous web pages can be resource-intensive. Monitor and manage system resources accordingly.

5. Have Fun

Users of this library are responsible for ensuring their usage complies with applicable laws, regulations, and ethical considerations in their jurisdiction and context of use.
*/
package chew

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/mmatongo/chew/v1/internal/common"
	"github.com/mmatongo/chew/v1/internal/document"
	"github.com/mmatongo/chew/v1/internal/text"
	"github.com/mmatongo/chew/v1/internal/transcribe"
	"github.com/mmatongo/chew/v1/internal/utils"
	"github.com/temoto/robotstxt"
	"golang.org/x/time/rate"
)

const (
	contentTypeHTML     = "text/html"
	contentTypeText     = "text/plain"
	contentTypeXML      = "application/xml"
	contentTypeTextXML  = "text/xml"
	contentTypePDF      = "application/pdf"
	contentTypeCSV      = "text/csv"
	contentTypeJSON     = "application/json"
	contentTypeYAML     = "application/x-yaml"
	contentTypeMarkdown = "text/markdown"
	contentTypeEPUB     = "application/epub+zip"
	contentTypeDocx     = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	contentTypePptx     = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
)

var contentTypeProcessors = map[string]func(io.Reader, string) ([]common.Chunk, error){
	contentTypeHTML:     text.ProcessHTML,
	contentTypeCSV:      text.ProcessCSV,
	contentTypeJSON:     text.ProcessJSON,
	contentTypeYAML:     text.ProcessYAML,
	contentTypeMarkdown: text.ProcessText,
	contentTypeText:     text.ProcessText,
	contentTypeXML:      text.ProcessXML,
	contentTypeTextXML:  text.ProcessXML,
	contentTypeDocx:     document.ProcessDocx,
	contentTypePptx:     document.ProcessPptx,
	contentTypePDF:      document.ProcessPDF,
	contentTypeEPUB:     document.ProcessEpub,
}

type Chew struct {
	config        common.Config
	httpClient    *http.Client
	rateLimiter   RateLimiter
	rateLimiterMu sync.RWMutex
	robotsCache   map[string]*robotstxt.RobotsData
	robotsMu      sync.RWMutex
	lastAccess    map[string]time.Time
	lastAccessMu  sync.Mutex
	proxyIndex    int
	proxyMu       sync.Mutex
}

type RateLimiter interface {
	Wait(context.Context) error
}

func (c *Chew) SetRateLimiter(rl RateLimiter) {
	c.rateLimiterMu.Lock()
	defer c.rateLimiterMu.Unlock()
	c.rateLimiter = rl
}

/*
NewConfig allows you to set the configuration options for URL processing. It takes a Config struct.

Usage:

	config := chew.Config{
		UserAgent:       "MyBot/1.0 (+https://example.com/bot)",
		RetryLimit:      3,
		RetryDelay:      5 * time.Second,
		CrawlDelay:      10 * time.Second,
		ProxyList:       []string{"http://proxy1.com", "http://proxy2.com"},
		RateLimit:       2 * time.Second,
		RateBurst:       3,
		IgnoreRobotsTxt: false,
	}

	chew.NewConfig(config)
*/
func New(config common.Config) *Chew {
	c := &Chew{
		config:      config,
		robotsCache: make(map[string]*robotstxt.RobotsData),
		lastAccess:  make(map[string]time.Time),
	}
	c.initHTTPClient()

	limit := rate.Every(config.RateLimit)
	c.rateLimiter = rate.NewLimiter(limit, config.RateBurst)

	return c
}

/*
Transcribe is a function that transcribes audio files using either the Google Cloud Speech-to-Text API
or the Whisper API. It handles uploading the audio file to Google Cloud Storage if necessary,
manages the transcription process, and returns the resulting transcript.

For detailed usage instructions, see the TranscribeOptions struct documentation.
*/
var Transcribe = transcribe.Transcribe

/*
The TranscribeOptions struct contains the options for transcribing an audio file. It allows the user
to specify the Google Cloud credentials, the GCS bucket to upload the audio file to, the language code
to use for transcription, an option to enable diarization (the process of separating and labeling
speakers in an audio stream) including the min and max speakers and
an option to clean up the audio file from Google Cloud Speech-to-Text (GCS) after transcription is complete.

And also, it allows the user to specify whether to use the Whisper API for transcription, and if so,
the API key, model, and prompt to use.

Usage:

	opts := chew.TranscribeOptions{
		CredentialsJSON:   []byte("..."),
		Bucket:            "my-bucket",
		LanguageCode:      "en-US",
		EnableDiarization: true,
		MinSpeakers:       2,
		MaxSpeakers:       4,
		CleanupOnComplete: true,
		UseWhisper:        true, // You can only have one of these enabled, by default it uses the Google Cloud Speech-to-Text API
		WhisperAPIKey:     "my-whisper-api-key",
		WhisperModel:      "whisper-1",
	}
*/
type TranscribeOptions = transcribe.TranscribeOptions

/*
Config struct contains the configuration options for URL processing.

Fields:
  - UserAgent: The user agent string to use for requests (e.g., "MyBot/1.0 (+https://example.com/bot)")
  - RetryLimit: Number of retries to attempt in case of failure (e.g., 3)
  - RetryDelay: Delay between retries (e.g., 5 * time.Second)
  - CrawlDelay: Delay between requests to the same domain (e.g., 10 * time.Second)
  - ProxyList: List of proxy URLs to use for requests (e.g., []string{"http://proxy1.com", "http://proxy2.com"})
  - RateLimit: Rate limit for requests (e.g., rate.Every(2 * time.Second))
  - RateBurst: Maximum burst size for rate limiting (e.g., 3)
  - IgnoreRobotsTxt: Whether to ignore robots.txt rules (e.g., false)

Usage:

	config := chew.Config{
	    UserAgent:       "MyBot/1.0 (+https://example.com/bot)",
	    RetryLimit:      3,
	    RetryDelay:      5 * time.Second,
	    CrawlDelay:      10 * time.Second,
	    ProxyList:       []string{"http://proxy1.com", "http://proxy2.com"},
	    RateLimit:       2 * time.Second,
	    RateBurst:       3,
	    IgnoreRobotsTxt: false,
	}
*/
type Config = common.Config

/*
This is meant as a fallback in case the content type is not recognized and to enforce
the content type based on the file extension instead of the content type
returned by the server. i.e. if the server returns text/plain but the file is a markdown file
the content types are the biggest culprits of this
*/
var validExtensions = map[string]func(io.Reader, string) ([]common.Chunk, error){
	".md":   text.ProcessText,
	".csv":  text.ProcessCSV,
	".json": text.ProcessJSON,
	".yaml": text.ProcessYAML,
	".html": text.ProcessHTML,
	".epub": document.ProcessEpub,
}

/*
SetHTTPClient allows you to set a custom http.Client to use for making requests.

This would be useful in the event custom logging, tracing, or other functionality is
required for the requests made by the library.

Usage:

	client := &http.Client{
		Transport: loggingRoundTripper{wrapped: http.DefaultTransport},
	}

	chew.SetHTTPClient(client)
*/

func (c *Chew) SetHTTPClient(client *http.Client) {
	c.httpClient = client
}

func (c *Chew) initHTTPClient() {
	transport := &http.Transport{
		Proxy: c.getProxy,
	}
	c.httpClient = &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}
}

/*
For content types that can also return text/plain as their content types we need to manually check
their extension to properly process them. I feel like this could be done better but this is my solution for now.
*/
func getProcessor(contentType, url string) (func(io.Reader, string) ([]common.Chunk, error), error) {
	for key, proc := range contentTypeProcessors {
		if strings.Contains(contentType, key) {
			return proc, nil
		}
	}

	ext, err := utils.GetFileExtension(url)
	if err != nil {
		return nil, fmt.Errorf("couldn't get file extension from url %s: %s", url, err)
	}

	if proc, ok := validExtensions[ext]; ok {
		return proc, nil
	}

	return nil, fmt.Errorf("unsupported content type: %s", contentType)
}

/*
Process takes a list of URLs and returns a list of Chunks

The slice of strings to be processed can be URLs or file paths
The context is optional and can be used to cancel the processing
of the URLs after a certain amount of time

This function is safe for concurrent use.

Usage:

	chunks, err := chew.Process([]string{"https://example.com", "file://path/to/file.txt"})
	if err != nil {
		log.Fatalf("Error processing URLs: %v", err)
	}

	for _, chunk := range chunks {
		log.Printf("Chunk: %s\n Source: %s\n", chunk.Content, chunk.Source)
	}
*/
func (c *Chew) Process(ctx context.Context, urls []string) ([]common.Chunk, error) {
	var (
		result []common.Chunk
		mu     sync.Mutex
		errCh  = make(chan error, len(urls))
		resCh  = make(chan []common.Chunk, len(urls))
	)

	for _, url := range urls {
		go func(url string) {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			default:
				c.rateLimiterMu.RLock()
				rateLimiter := c.rateLimiter
				c.rateLimiterMu.RUnlock()

				if err := rateLimiter.Wait(ctx); err != nil {
					errCh <- fmt.Errorf("rate limit exceeded for %s: %w", url, err)
					return
				}

				if !c.config.IgnoreRobotsTxt {
					allowed, crawlDelay, err := c.getRobotsTxtInfo(url)
					if err != nil {
						errCh <- fmt.Errorf("checking robots.txt for %s: %w", url, err)
						return
					}
					if !allowed {
						errCh <- fmt.Errorf("access to %s is disallowed by robots.txt", url)
						return
					}
					if err := c.respectCrawlDelay(ctx, url, crawlDelay); err != nil {
						errCh <- fmt.Errorf("respecting crawl delay for %s: %w", url, err)
						return
					}
				}

				chunks, err := c.processWithRetry(ctx, url)
				if err != nil {
					errCh <- fmt.Errorf("processing %s: %w", url, err)
					return
				}

				resCh <- chunks
			}
		}(url)
	}

	for i := 0; i < len(urls); i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case err := <-errCh:
			return nil, err
		case chunks := <-resCh:
			mu.Lock()
			result = append(result, chunks...)
			mu.Unlock()
		}
	}

	return result, nil
}

/*
processURL handles the actual processing of a single URL or file
file paths are processed directly while URLs are fetched and processed
*/
func (c *Chew) processURL(ctx context.Context, url string) ([]common.Chunk, error) {
	// if the url is a file path we can just open the file and process it directly
	if filePath, found := strings.CutPrefix(url, "file://"); found {
		file, err := utils.OpenFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("opening file: %w", err)
		}
		defer file.Close()

		ext, _ := utils.GetFileExtension(filePath)
		/*
			Will leave this in here for now, but I think it's better to just check the file extension
			instead of the content type returned.
		*/
		contentType := utils.GetFileContentType(file)

		proc, err := getProcessor(contentType, filePath)
		if err != nil {
			proc, ok := validExtensions[ext]
			if !ok {
				return nil, fmt.Errorf("unsupported file type: %s", ext)
			}
			return proc(file, url)
		}

		return proc(file, url)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", c.config.UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")

	processor, err := getProcessor(contentType, url)
	if err != nil {
		return nil, err
	}

	return processor(resp.Body, url)
}

func (c *Chew) getRobotsTxtInfo(urlStr string) (bool, time.Duration, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false, 0, err
	}

	robotsURL := fmt.Sprintf("%s://%s/robots.txt", parsedURL.Scheme, parsedURL.Host)

	c.robotsMu.RLock()
	robotsData, exists := c.robotsCache[robotsURL]
	c.robotsMu.RUnlock()

	if !exists {
		resp, err := http.Get(robotsURL)
		if err != nil {
			return true, c.config.CrawlDelay, nil
		}
		defer resp.Body.Close()

		robotsData, err = robotstxt.FromResponse(resp)
		if err != nil {
			return true, c.config.CrawlDelay, nil
		}

		c.robotsMu.Lock()
		c.robotsCache[robotsURL] = robotsData
		c.robotsMu.Unlock()
	}

	allowed := robotsData.TestAgent(parsedURL.Path, c.config.UserAgent)

	return allowed, c.config.CrawlDelay, nil
}

// respectCrawlDelay ensures that subsequent requests to the same domain respect the specified crawl delay.
func (c *Chew) respectCrawlDelay(ctx context.Context, urlStr string, delay time.Duration) error {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	domain := parsedURL.Hostname()

	c.lastAccessMu.Lock()
	lastAccess, exists := c.lastAccess[domain]
	if exists {
		timeToWait := time.Until(lastAccess.Add(delay))
		if timeToWait > 0 {
			c.lastAccessMu.Unlock()
			select {
			case <-time.After(timeToWait):
			case <-ctx.Done():
				return ctx.Err()
			}
			c.lastAccessMu.Lock()
		}
	}

	c.lastAccess[domain] = time.Now()
	c.lastAccessMu.Unlock()
	return nil
}

func (c *Chew) processWithRetry(ctx context.Context, url string) ([]common.Chunk, error) {
	var (
		chunks []common.Chunk
		err    error
	)

	var retries int
	for {
		chunks, err = c.processURL(ctx, url)
		if err == nil {
			return chunks, nil
		}
		if retries > c.config.RetryLimit {
			break
		}
		retries++
		c.wait(ctx, c.config.RetryDelay)
	}

	return nil, err
}

func (c *Chew) wait(ctx context.Context, d time.Duration) {
	select {
	case <-time.After(d):
	case <-ctx.Done():
	}
}

func (c *Chew) getProxy(req *http.Request) (*url.URL, error) {
	c.proxyMu.Lock()
	defer c.proxyMu.Unlock()

	if len(c.config.ProxyList) == 0 {
		return nil, nil
	}

	proxyURL, err := url.Parse(c.config.ProxyList[c.proxyIndex])
	if err != nil {
		return nil, err
	}

	c.proxyIndex = (c.proxyIndex + 1) % len(c.config.ProxyList)
	return proxyURL, nil
}
