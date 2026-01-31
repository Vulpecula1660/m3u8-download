package downloader

import (
	"io"
	"net/http"
	"time"

	"m3u8-download/pkg/m3u8"
)

type HTTPClient struct {
	client    *http.Client
	timeout   time.Duration
	retries   int
	retryWait time.Duration
	userAgent string
}

func NewHTTPClient(cfg *m3u8.DownloadConfig) *HTTPClient {
	client := &HTTPClient{
		timeout:   time.Duration(cfg.Timeout) * time.Second,
		retries:   cfg.Retries,
		retryWait: 3 * time.Second,
		userAgent: cfg.UserAgent,
	}

	client.client = &http.Client{
		Timeout: client.timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return client
}

func (c *HTTPClient) Get(url string) ([]byte, error) {
	var body []byte
	var err error

	for attempt := 0; attempt <= c.retries; attempt++ {
		if attempt > 0 {
			waitTime := time.Duration(1<<uint(attempt-1)) * c.retryWait
			if waitTime > 30*time.Second {
				waitTime = 30 * time.Second
			}
			time.Sleep(waitTime)
		}

		body, err = c.doGet(url)
		if err == nil {
			return body, nil
		}

		if httpErr, ok := err.(*m3u8.HTTPError); ok {
			if httpErr.StatusCode >= 400 && httpErr.StatusCode < 500 {
				return nil, err
			}
		}
	}

	return nil, m3u8.NewRetryExhaustedError(c.retries, err)
}

func (c *HTTPClient) doGet(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, m3u8.NewHTTPError(resp.StatusCode, url)
	}

	return io.ReadAll(resp.Body)
}
