package config

import (
	"bytes"
	"os"
	"strings"
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

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantMode    ParseMode
		wantErr     bool
		errContains string
		stdoutHas   string
		stderrHas   string
		validateCfg func(t *testing.T, cfg *m3u8.DownloadConfig)
	}{
		{
			name:      "no args shows help",
			args:      []string{},
			wantMode:  ParseModeShowHelp,
			stdoutHas: "用法：",
		},
		{
			name:      "help subcommand shows help",
			args:      []string{"help"},
			wantMode:  ParseModeShowHelp,
			stdoutHas: "m3u8-download help",
		},
		{
			name:      "short help flag shows help",
			args:      []string{"-h"},
			wantMode:  ParseModeShowHelp,
			stdoutHas: "顯示說明",
		},
		{
			name:      "long help flag shows help",
			args:      []string{"--help"},
			wantMode:  ParseModeShowHelp,
			stdoutHas: "選項：",
		},
		{
			name:     "short version flag shows version mode",
			args:     []string{"-version"},
			wantMode: ParseModeShowVersion,
		},
		{
			name:     "long version flag shows version mode",
			args:     []string{"--version"},
			wantMode: ParseModeShowVersion,
		},
		{
			name:        "missing url returns error",
			args:        []string{"-workers", "2"},
			wantMode:    ParseModeRun,
			wantErr:     true,
			errContains: "請使用 -h、--help 或 help 查看說明",
		},
		{
			name:        "unknown flag returns error and parser message",
			args:        []string{"-unknown"},
			wantMode:    ParseModeRun,
			wantErr:     true,
			errContains: "請使用 -h、--help 或 help 查看說明",
			stderrHas:   "flag provided but not defined: -unknown",
		},
		{
			name:     "valid config and defaults applied",
			args:     []string{"-url", "http://example.com/video.m3u8", "-workers", "0", "-retries", "-1", "-timeout", "0"},
			wantMode: ParseModeRun,
			validateCfg: func(t *testing.T, cfg *m3u8.DownloadConfig) {
				t.Helper()
				if cfg.URL != "http://example.com/video.m3u8" {
					t.Fatalf("cfg.URL = %q, want %q", cfg.URL, "http://example.com/video.m3u8")
				}
				if cfg.Workers != 15 {
					t.Fatalf("cfg.Workers = %d, want 15", cfg.Workers)
				}
				if cfg.Retries != 3 {
					t.Fatalf("cfg.Retries = %d, want 3", cfg.Retries)
				}
				if cfg.Timeout != 30 {
					t.Fatalf("cfg.Timeout = %d, want 30", cfg.Timeout)
				}
				if cfg.UserAgent == "" {
					t.Fatal("cfg.UserAgent should not be empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			cfg, mode, err := ParseArgs(tt.args, &stdout, &stderr)
			if mode != tt.wantMode {
				t.Fatalf("mode = %v, want %v", mode, tt.wantMode)
			}

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.errContains)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.stdoutHas != "" && !strings.Contains(stdout.String(), tt.stdoutHas) {
				t.Fatalf("stdout %q does not contain %q", stdout.String(), tt.stdoutHas)
			}

			if tt.stderrHas != "" && !strings.Contains(stderr.String(), tt.stderrHas) {
				t.Fatalf("stderr %q does not contain %q", stderr.String(), tt.stderrHas)
			}

			if tt.validateCfg != nil {
				tt.validateCfg(t, cfg)
			}
		})
	}
}
