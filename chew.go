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
	contentTypeDocx:     document.ProcessDocx,
	contentTypePptx:     document.ProcessPptx,
	contentTypePDF:      document.ProcessPDF,
	contentTypeEPUB:     document.ProcessEpub,
}

var (
	rateLimiter       *rate.Limiter
	robotsCache       = make(map[string]*robotstxt.RobotsData)
	robotsMu          sync.RWMutex
	userAgent         = "Chew/1.0 (+https://github.com/mmatongo/chew)"
	lastAccessTimes   = make(map[string]time.Time)
	lastAccessMu      sync.Mutex
	retryLimit        = 3
	retryDelay        = 5 * time.Second
	crawlDelay        = 10 * time.Second
	proxyList         []string
	currentProxyIndex int
	proxyMu           sync.Mutex
	ignoreRobotsTxt   bool
)

/*
NewConfig allows you to set the configuration options for URL processing. It takes a Config struct.
*/
func NewConfig(config common.Config) {
	userAgent = config.UserAgent
	retryLimit = config.RetryLimit
	retryDelay = config.RetryDelay
	crawlDelay = config.CrawlDelay
	proxyList = config.ProxyList
	rateLimiter = rate.NewLimiter(config.RateLimit, config.RateBurst)
	ignoreRobotsTxt = config.IgnoreRobotsTxt
}

func init() {
	NewConfig(common.Config{
		UserAgent:       userAgent,
		RetryLimit:      3,
		RetryDelay:      5 * time.Second,
		CrawlDelay:      10 * time.Second,
		ProxyList:       []string{},
		RateLimit:       rate.Every(2 * time.Second),
		RateBurst:       3,
		IgnoreRobotsTxt: false,
	})
}

/*
Transcribe uses the Google Cloud Speech-to-Text API to transcribe an audio file. It takes
a context, the filename of the audio file to transcribe, and a TranscribeOptions struct which
contains the Google Cloud credentials, the GCS bucket to upload the audio file to, and the language code
to use for transcription. It returns the transcript of the audio file as a string and an error if the
transcription fails.
*/
var Transcribe = transcribe.Transcribe

type TranscribeOptions = transcribe.TranscribeOptions

/*
The Config struct contains the configuration options for the library. You can specify the
user agent to use for requests, number of retries to attempt in case of failure, delay between retries,
delay between requests, a list of proxies to use for requests, a rate limit for requests, and an option to
ignore the robots.txt file.
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

var httpClient *http.Client

// Obviously, this is not the best way to do this but it's a start

/*
SetHTTPClient allows you to set a custom http.Client to use for making requests.

This would be useful in the event custom logging, tracing, or other functionality is
required for the requests made by the library.
*/
func SetHTTPClient(client *http.Client) {
	httpClient = client
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
*/
func Process(urls []string, ctxs ...context.Context) ([]common.Chunk, error) {
	ctx := context.Background()
	if len(ctxs) > 0 {
		ctx = ctxs[0]
	}

	var (
		result []common.Chunk
		wg     sync.WaitGroup
		mu     sync.Mutex
		errCh  = make(chan error, len(urls))
	)

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			default:
				err := rateLimiter.Wait(ctx)
				if err != nil {
					errCh <- fmt.Errorf("rate limiting for %s: %w", url, err)
					return
				}

				if !ignoreRobotsTxt {
					allowed, crawlDelay, err := getRobotsTxtInfo(url)
					if err != nil {
						errCh <- fmt.Errorf("checking robots.txt for %s: %w", url, err)
						return
					}
					if !allowed {
						errCh <- fmt.Errorf("access to %s is disallowed by robots.txt", url)
						return
					}
					if err := respectCrawlDelay(url, crawlDelay); err != nil {
						errCh <- fmt.Errorf("respecting crawl delay for %s: %w", url, err)
						return
					}
				}
				chunks, err := processWithRetry(url, ctx)
				if err != nil {
					errCh <- fmt.Errorf("processing %s, %w", url, err)
					return
				}
				mu.Lock()
				result = append(result, chunks...)
				mu.Unlock()
			}
		}(url)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return nil, err
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return result, nil
}

func processURL(url string, ctxs ...context.Context) ([]common.Chunk, error) {
	ctx := context.Background()
	if len(ctxs) > 0 {
		ctx = ctxs[0]
	}

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

	req.Header.Set("User-Agent", userAgent)

	var client *http.Client
	if httpClient != nil {
		client = httpClient
	} else {
		transport := http.DefaultTransport
		if proxyURL := getNextProxy(); proxyURL != nil {
			transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
		}
		client = &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		}
	}

	resp, err := client.Do(req)
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

func getRobotsTxtInfo(urlStr string) (bool, time.Duration, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false, 0, err
	}

	robotsURL := fmt.Sprintf("%s://%s/robots.txt", parsedURL.Scheme, parsedURL.Host)

	robotsMu.RLock()
	robotsData, exists := robotsCache[robotsURL]
	robotsMu.RUnlock()

	/*
		This bit of code makes a lot of very unhealthy assumptions, in an ideal world this
		would be more straightforward but unfortunately it's not.

		Use this with caution and respect for site owners.
	*/
	if !exists {
		resp, err := http.Get(robotsURL)
		if err != nil {
			return true, crawlDelay, nil
		}
		defer resp.Body.Close()

		robotsData, err = robotstxt.FromResponse(resp)
		if err != nil {
			return true, crawlDelay, nil
		}

		robotsMu.Lock()
		robotsCache[robotsURL] = robotsData
		robotsMu.Unlock()
	}

	allowed := robotsData.TestAgent(parsedURL.Path, userAgent)

	return allowed, crawlDelay, nil
}

func respectCrawlDelay(urlStr string, delay time.Duration) error {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	domain := parsedURL.Hostname()

	lastAccessMu.Lock()
	defer lastAccessMu.Unlock()

	lastAccess, exists := lastAccessTimes[domain]
	if exists {
		timeToWait := time.Until(lastAccess.Add(delay))
		if timeToWait > 0 {
			time.Sleep(timeToWait)
		}
	}

	lastAccessTimes[domain] = time.Now()
	return nil
}

func processWithRetry(url string, ctx context.Context) ([]common.Chunk, error) {
	var (
		chunks []common.Chunk
		err    error
	)

	for i := 0; i < retryLimit; i++ {
		chunks, err = processURL(url, ctx)
		if err == nil {
			return chunks, nil
		}

		if i < retryLimit-1 {
			time.Sleep(retryDelay)
		}
	}

	return nil, err
}

func getNextProxy() *url.URL {
	proxyMu.Lock()
	defer proxyMu.Unlock()

	if len(proxyList) == 0 {
		return nil
	}

	proxy := proxyList[currentProxyIndex]
	currentProxyIndex = (currentProxyIndex + 1) % len(proxyList)

	proxyURL, _ := url.Parse(proxy)
	return proxyURL
}
