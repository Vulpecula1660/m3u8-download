package config

import (
	"os"
	"testing"

	"m3u8-download/pkg/m3u8"
)

func TestEnsureCacheDir(t *testing.T) {
	id := "test-cache-id"

	cacheDir, err := EnsureCacheDir(id)
	if err != nil {
		t.Fatalf("EnsureCacheDir failed: %v", err)
	}

	if cacheDir == "" {
		t.Error("cacheDir is empty")
	}

	stat, err := os.Stat(cacheDir)
	if err != nil {
		t.Errorf("cache directory doesn't exist: %v", err)
	}

	if !stat.IsDir() {
		t.Error("cache path is not a directory")
	}

	err = CleanupCacheDir(cacheDir)
	if err != nil {
		t.Errorf("failed to cleanup cache directory: %v", err)
	}

	_, err = os.Stat(cacheDir)
	if !os.IsNotExist(err) {
		t.Error("cache directory still exists after cleanup")
	}
}

func TestCleanupCacheDir(t *testing.T) {
	id := "test-cleanup-cache"

	cacheDir, err := EnsureCacheDir(id)
	if err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}

	err = CleanupCacheDir(cacheDir)
	if err != nil {
		t.Errorf("CleanupCacheDir failed: %v", err)
	}

	_, err = os.Stat(cacheDir)
	if !os.IsNotExist(err) {
		t.Error("cache directory still exists after cleanup")
	}
}

func TestGetHTTPClient(t *testing.T) {
	cfg := &m3u8.DownloadConfig{
		Timeout:   30,
		Retries:   3,
		UserAgent: "test-user-agent",
	}

	timeout, retries, userAgent := GetHTTPClient(cfg)

	if timeout.Seconds() != 30 {
		t.Errorf("got timeout %v, want 30s", timeout)
	}

	if retries != 3 {
		t.Errorf("got retries %d, want 3", retries)
	}

	if userAgent != "test-user-agent" {
		t.Errorf("got user-agent %q, want %q", userAgent, "test-user-agent")
	}
}
