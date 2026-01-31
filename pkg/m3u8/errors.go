package m3u8

import (
	"fmt"
)

var (
	ErrInvalidURL     = fmt.Errorf("invalid M3U8 URL")
	ErrNoTSFiles      = fmt.Errorf("no TS files found in playlist")
	ErrDecryptFailed  = fmt.Errorf("AES decryption failed")
	ErrDownloadFailed = fmt.Errorf("download failed after retries")
	ErrMergeFailed    = fmt.Errorf("failed to merge files")
	ErrInvalidKey     = fmt.Errorf("invalid decryption key")
	ErrInvalidIV      = fmt.Errorf("invalid initialization vector")
)

type HTTPError struct {
	StatusCode int
	URL        string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP error: status %d for URL %s", e.StatusCode, e.URL)
}

func NewHTTPError(statusCode int, url string) *HTTPError {
	return &HTTPError{StatusCode: statusCode, URL: url}
}

type RetryExhaustedError struct {
	Attempts int
	LastErr  error
}

func (e *RetryExhaustedError) Error() string {
	return fmt.Errorf("failed after %d attempts: %w", e.Attempts, e.LastErr).Error()
}

func NewRetryExhaustedError(attempts int, err error) *RetryExhaustedError {
	return &RetryExhaustedError{Attempts: attempts, LastErr: err}
}
