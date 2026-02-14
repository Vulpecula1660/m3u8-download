package main

import (
	"bytes"
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

func TestRunCLIPaths(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantCode    int
		wantStdout  string
		wantStderr  string
		avoidStderr string
		avoidStdout string
	}{
		{
			name:       "no args shows help",
			args:       []string{},
			wantCode:   0,
			wantStdout: "用法：",
		},
		{
			name:       "help subcommand shows help",
			args:       []string{"help"},
			wantCode:   0,
			wantStdout: "m3u8-download help",
		},
		{
			name:        "help flag shows help",
			args:        []string{"-h"},
			wantCode:    0,
			wantStdout:  "顯示說明",
			avoidStderr: "Error:",
		},
		{
			name:        "short version flag shows version",
			args:        []string{"-version"},
			wantCode:    0,
			wantStdout:  "m3u8-download version",
			avoidStderr: "Error:",
		},
		{
			name:        "long version flag shows version",
			args:        []string{"--version"},
			wantCode:    0,
			wantStdout:  "m3u8-download version",
			avoidStderr: "Error:",
		},
		{
			name:        "missing required url returns error",
			args:        []string{"-workers", "2"},
			wantCode:    1,
			wantStderr:  "請使用 -h、--help 或 help 查看說明",
			avoidStdout: "用法：",
		},
		{
			name:       "unknown flag returns error",
			args:       []string{"-unknown"},
			wantCode:   1,
			wantStderr: "請使用 -h、--help 或 help 查看說明",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			code := run(tt.args, &stdout, &stderr)
			if code != tt.wantCode {
				t.Fatalf("run() code = %d, want %d", code, tt.wantCode)
			}

			if tt.wantStdout != "" && !strings.Contains(stdout.String(), tt.wantStdout) {
				t.Errorf("stdout %q does not contain %q", stdout.String(), tt.wantStdout)
			}

			if tt.wantStderr != "" && !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Errorf("stderr %q does not contain %q", stderr.String(), tt.wantStderr)
			}

			if tt.avoidStdout != "" && strings.Contains(stdout.String(), tt.avoidStdout) {
				t.Errorf("stdout %q should not contain %q", stdout.String(), tt.avoidStdout)
			}

			if tt.avoidStderr != "" && strings.Contains(stderr.String(), tt.avoidStderr) {
				t.Errorf("stderr %q should not contain %q", stderr.String(), tt.avoidStderr)
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
