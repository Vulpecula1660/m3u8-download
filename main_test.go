package main

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"m3u8-download/internal/config"
	"m3u8-download/internal/downloader"
	"m3u8-download/internal/parser"
	"m3u8-download/pkg/m3u8"
)

func TestIntegrationFullDownload(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".m3u8") {
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			w.Write([]byte(`#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:10
#EXTINF:10.0,
segment1.ts
#EXTINF:10.0,
segment2.ts
#EXT-X-ENDLIST`))
		} else if strings.HasSuffix(r.URL.Path, ".ts") {
			w.Header().Set("Content-Type", "video/mp2t")
			w.Write([]byte{0x47, 0x00, 0x00, 0x00})
		} else if strings.HasSuffix(r.URL.Path, ".key") {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write([]byte("0123456789012345"))
		}
	}))
	defer ts.Close()

	cfg := &m3u8.DownloadConfig{
		URL:     ts.URL + "/video.m3u8",
		Output:  "test_output.ts",
		Workers: 2,
		Retries: 2,
		Timeout: 10,
		Verbose: false,
	}

	httpClient := downloader.NewHTTPClient(cfg)
	logger := setupLogger(false)
	dl := downloader.NewDownloader(httpClient, logger)

	body, err := httpClient.Get(cfg.URL)
	if err != nil {
		t.Fatalf("Failed to fetch M3U8: %v", err)
	}

	playlist, err := parser.ParsePlaylist(string(body), cfg.URL)
	if err != nil {
		t.Fatalf("Failed to parse playlist: %v", err)
	}

	id := "test-integration"
	cacheDir, err := config.EnsureCacheDir(id)
	if err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}
	defer config.CleanupCacheDir(cacheDir)

	stats, err := dl.DownloadSegments(playlist, cacheDir, cfg.Workers)
	if err != nil {
		t.Fatalf("Failed to download segments: %v", err)
	}

	if stats.Total != 2 {
		t.Errorf("got %d segments, want 2", stats.Total)
	}

	if stats.Completed != 2 {
		t.Errorf("got %d completed, want 2", stats.Completed)
	}

	err = dl.MergeFiles(cacheDir, cfg.Output)
	if err != nil {
		t.Fatalf("Failed to merge files: %v", err)
	}

	_, err = os.Stat(cfg.Output)
	if err != nil {
		t.Errorf("output file doesn't exist: %v", err)
	}

	os.Remove(cfg.Output)
}

func TestIntegrationEncryptedDownload(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".m3u8") {
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			w.Write([]byte(`#EXTM3U
#EXT-X-VERSION:3
#EXT-X-KEY:METHOD=AES-128,URI="key.key"
#EXT-X-TARGETDURATION:10
#EXTINF:10.0,
segment1.ts
#EXT-X-ENDLIST`))
		} else if strings.HasSuffix(r.URL.Path, ".key") {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write([]byte("0123456789012345"))
		} else if strings.HasSuffix(r.URL.Path, ".ts") {
			w.Header().Set("Content-Type", "video/mp2t")
			w.Write([]byte{0x47, 0x00, 0x00, 0x00})
		}
	}))
	defer ts.Close()

	cfg := &m3u8.DownloadConfig{
		URL:     ts.URL + "/video.m3u8",
		Output:  "test_encrypted_output.ts",
		Workers: 2,
		Retries: 2,
		Timeout: 10,
		Verbose: false,
	}

	httpClient := downloader.NewHTTPClient(cfg)
	logger := setupLogger(false)
	dl := downloader.NewDownloader(httpClient, logger)

	body, err := httpClient.Get(cfg.URL)
	if err != nil {
		t.Fatalf("Failed to fetch M3U8: %v", err)
	}

	playlist, err := parser.ParsePlaylist(string(body), cfg.URL)
	if err != nil {
		t.Fatalf("Failed to parse playlist: %v", err)
	}

	id := "test-encrypted-integration"
	cacheDir, err := config.EnsureCacheDir(id)
	if err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}
	defer config.CleanupCacheDir(cacheDir)

	_, err = dl.DownloadSegments(playlist, cacheDir, cfg.Workers)
	if err != nil {
		t.Errorf("Failed to download segments: %v", err)
	}

	os.Remove(cfg.Output)
}

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid url",
			args:    []string{"-url", "http://example.com/video.m3u8"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)

			var url string
			fs.StringVar(&url, "url", "", "M3U8 URL (required)")

			err := fs.Parse(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if url == "" {
				t.Error("URL is empty")
			}
		})
	}
}

func TestEnsureCacheDirCreatesDirectory(t *testing.T) {
	id := "test-cache-dir"

	cacheDir, err := config.EnsureCacheDir(id)
	if err != nil {
		t.Fatalf("EnsureCacheDir failed: %v", err)
	}

	info, err := os.Stat(cacheDir)
	if err != nil {
		t.Errorf("cache directory doesn't exist: %v", err)
	}

	if !info.IsDir() {
		t.Error("path is not a directory")
	}

	if !strings.Contains(cacheDir, id) {
		t.Errorf("cache directory path %q doesn't contain %q", cacheDir, id)
	}

	config.CleanupCacheDir(cacheDir)
}
