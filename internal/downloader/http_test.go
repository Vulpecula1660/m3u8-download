package downloader

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"m3u8-download/pkg/m3u8"
)

func TestNewHTTPClient(t *testing.T) {
	cfg := &m3u8.DownloadConfig{
		Timeout:   30,
		Retries:   3,
		UserAgent: "test-agent",
	}

	client := NewHTTPClient(cfg)

	if client == nil {
		t.Fatal("client is nil")
	}

	if client.timeout != 30*time.Second {
		t.Errorf("got timeout %v, want 30s", client.timeout)
	}

	if client.retries != 3 {
		t.Errorf("got retries %d, want 3", client.retries)
	}

	if client.userAgent != "test-agent" {
		t.Errorf("got user-agent %q, want %q", client.userAgent, "test-agent")
	}
}

func TestHTTPClient_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write([]byte("test data"))
	}))
	defer ts.Close()

	cfg := &m3u8.DownloadConfig{
		Timeout:   10,
		Retries:   2,
		UserAgent: "test-agent",
	}

	client := NewHTTPClient(cfg)

	body, err := client.Get(ts.URL)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if string(body) != "test data" {
		t.Errorf("got %q, want %q", string(body), "test data")
	}
}

func TestHTTPClient_GetRetry(t *testing.T) {
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write([]byte("test data"))
	}))
	defer ts.Close()

	cfg := &m3u8.DownloadConfig{
		Timeout:   10,
		Retries:   3,
		UserAgent: "test-agent",
	}

	client := NewHTTPClient(cfg)

	body, err := client.Get(ts.URL)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if string(body) != "test data" {
		t.Errorf("got %q, want %q", string(body), "test data")
	}

	if attempts != 3 {
		t.Errorf("got %d attempts, want 3", attempts)
	}
}

func TestHTTPClient_GetExhaustedRetries(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	cfg := &m3u8.DownloadConfig{
		Timeout:   10,
		Retries:   2,
		UserAgent: "test-agent",
	}

	client := NewHTTPClient(cfg)

	_, err := client.Get(ts.URL)
	if err == nil {
		t.Error("expected error but got none")
	}

	retryErr, ok := err.(*m3u8.RetryExhaustedError)
	if !ok {
		t.Errorf("got error type %T, want *RetryExhaustedError", err)
	}

	if retryErr == nil {
		t.Fatal("retryErr is nil")
	}

	if retryErr.Attempts != 2 {
		t.Errorf("got %d attempts, want 2", retryErr.Attempts)
	}
}

func TestNewDownloader(t *testing.T) {
	cfg := &m3u8.DownloadConfig{
		Timeout:   10,
		Retries:   2,
		UserAgent: "test-agent",
	}

	httpClient := NewHTTPClient(cfg)
	var logger *slog.Logger

	dl := NewDownloader(httpClient, logger)

	if dl == nil {
		t.Fatal("downloader is nil")
	}

	if dl.httpClient != httpClient {
		t.Error("httpClient not set correctly")
	}

	if dl.logger != logger {
		t.Error("logger not set correctly")
	}
}
