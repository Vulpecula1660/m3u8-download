package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"m3u8-download/internal/config"
	"m3u8-download/internal/downloader"
	"m3u8-download/internal/parser"

	"github.com/twinj/uuid"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	cfg, mode, err := config.ParseArgs(args, stdout, stderr)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	if mode == config.ParseModeShowHelp {
		return 0
	}

	logger := setupLogger(cfg.Verbose)

	id := uuid.NewV4().String()

	cacheDir, err := config.EnsureCacheDir(id)
	if err != nil {
		logger.Error("Failed to create cache directory", "error", err)
		return 1
	}

	httpClient := downloader.NewHTTPClient(cfg)
	dl := downloader.NewDownloader(httpClient, logger)

	logger.Info("Fetching M3U8 playlist", "url", cfg.URL)
	body, err := httpClient.Get(cfg.URL)
	if err != nil {
		logger.Error("Failed to fetch M3U8", "error", err)
		return 1
	}

	logger.Info("Parsing playlist")
	playlist, err := parser.ParsePlaylist(string(body), cfg.URL)
	if err != nil {
		logger.Error("Failed to parse playlist", "error", err)
		return 1
	}

	logger.Info("Playlist parsed", "segments", len(playlist.Segments), "encrypted", playlist.IsEncrypted)

	if cfg.Output == "" {
		cfg.Output = fmt.Sprintf("%s.ts", id)
	}

	logger.Info("Starting download", "output", cfg.Output, "workers", cfg.Workers)
	startTime := time.Now()

	stats, err := dl.DownloadSegments(playlist, cacheDir, cfg.Workers)
	if err != nil {
		logger.Error("Download failed", "error", err)
		return 1
	}

	logger.Info("Merging files")
	if err := dl.MergeFiles(cacheDir, cfg.Output); err != nil {
		logger.Error("Failed to merge files", "error", err)
		return 1
	}

	if err := config.CleanupCacheDir(cacheDir); err != nil {
		logger.Warn("Failed to cleanup cache directory", "error", err)
	}

	elapsed := time.Since(startTime)
	logger.Info("Download completed",
		"file", cfg.Output,
		"segments", stats.Total,
		"completed", stats.Completed,
		"failed", stats.Failed,
		"duration", elapsed,
	)

	return 0
}

func setupLogger(verbose bool) *slog.Logger {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
}
