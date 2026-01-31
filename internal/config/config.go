package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"m3u8-download/pkg/m3u8"
)

func ParseFlags() (*m3u8.DownloadConfig, error) {
	var cfg m3u8.DownloadConfig

	flag.StringVar(&cfg.URL, "url", "", "M3U8 URL (required)")
	flag.StringVar(&cfg.Output, "output", "", "Output filename (.ts)")
	flag.IntVar(&cfg.Workers, "workers", 15, "Number of concurrent workers")
	flag.IntVar(&cfg.Retries, "retries", 3, "Number of retry attempts")
	flag.IntVar(&cfg.Timeout, "timeout", 30, "Request timeout in seconds")
	flag.StringVar(&cfg.UserAgent, "user-agent", "", "Custom User-Agent")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Enable verbose logging")
	flag.StringVar(&cfg.ProxyURL, "proxy", "", "Proxy URL")

	flag.Parse()

	if cfg.URL == "" {
		return nil, fmt.Errorf("-url flag is required")
	}

	if cfg.Workers <= 0 {
		cfg.Workers = 15
	}

	if cfg.Retries <= 0 {
		cfg.Retries = 3
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 30
	}

	if cfg.UserAgent == "" {
		cfg.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	}

	return &cfg, nil
}

func GetHTTPClient(cfg *m3u8.DownloadConfig) (*time.Duration, int, string) {
	timeout := time.Duration(cfg.Timeout) * time.Second
	retryCount := cfg.Retries
	userAgent := cfg.UserAgent

	return &timeout, retryCount, userAgent
}

func EnsureCacheDir(id string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	cacheDir := fmt.Sprintf("%s/cache/%s", wd, id)
	err = os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	return cacheDir, nil
}

func CleanupCacheDir(cacheDir string) error {
	err := os.RemoveAll(cacheDir)
	if err != nil {
		return fmt.Errorf("failed to cleanup cache directory: %w", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	cacheRoot := fmt.Sprintf("%s/cache", wd)
	if _, err := os.Stat(cacheRoot); err == nil {
		_ = os.Remove(cacheRoot)
	}

	return nil
}
