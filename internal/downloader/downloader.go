package downloader

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"m3u8-download/internal/decrypt"
	"m3u8-download/pkg/m3u8"

	progressbar "github.com/schollz/progressbar/v3"
)

type Downloader struct {
	httpClient *HTTPClient
	logger     *slog.Logger
}

func NewDownloader(httpClient *HTTPClient, logger *slog.Logger) *Downloader {
	return &Downloader{
		httpClient: httpClient,
		logger:     logger,
	}
}

func (d *Downloader) DownloadSegments(playlist *m3u8.Playlist, cacheDir string, workers int) (*m3u8.DownloadStats, error) {
	stats := &m3u8.DownloadStats{
		Total:     len(playlist.Segments),
		StartTime: 0,
	}

	bar := progressbar.Default(int64(len(playlist.Segments)))

	var wg sync.WaitGroup
	ch := make(chan struct{}, workers)

	var mu sync.Mutex
	var keyData []byte
	var keyDataErr error
	var keyDataLoaded bool

	if playlist.IsEncrypted && playlist.Key != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			keyData, keyDataErr = d.httpClient.Get(playlist.Key)
			keyDataLoaded = true
		}()
	}

	wg.Add(len(playlist.Segments))
	for i, segment := range playlist.Segments {
		ch <- struct{}{}

		go func(idx int, seg *m3u8.TSInfo) {
			defer func() {
				bar.Add(1)
				<-ch
				wg.Done()
			}()

			filePath := fmt.Sprintf("%s/%s", cacheDir, seg.Name)

			err := d.downloadSegment(seg.Url, filePath, playlist.IsEncrypted, keyData, keyDataLoaded, &mu)
			if err != nil {
				d.logger.Error("Failed to download segment", "index", idx, "url", seg.Url, "error", err)
				mu.Lock()
				stats.Failed++
				mu.Unlock()
				return
			}

			mu.Lock()
			stats.Completed++
			mu.Unlock()
		}(i, segment)
	}

	wg.Wait()

	if keyDataErr != nil {
		return stats, fmt.Errorf("failed to download encryption key: %w", keyDataErr)
	}

	return stats, nil
}

func (d *Downloader) downloadSegment(url, filePath string, isEncrypted bool, key []byte, keyLoaded bool, mu *sync.Mutex) error {
	data, err := d.httpClient.Get(url)
	if err != nil {
		return err
	}

	if isEncrypted {
		mu.Lock()
		ready := keyLoaded
		mu.Unlock()

		if !ready {
			return fmt.Errorf("encryption key not loaded")
		}

		mu.Lock()
		decrypted, err := decrypt.Decrypt(data, key, nil)
		mu.Unlock()

		if err != nil {
			return fmt.Errorf("decryption failed: %w", err)
		}

		data = decrypt.RemoveSyncBytePrefix(decrypted)
	} else {
		data = decrypt.RemoveSyncBytePrefix(data)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	writer.Flush()

	return nil
}

func (d *Downloader) MergeFiles(cacheDir, output string) error {
	outFile, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := fmt.Sprintf("%s/%s", cacheDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			d.logger.Warn("Failed to read file", "path", filePath, "error", err)
			continue
		}

		if _, err := outFile.Write(data); err != nil {
			return fmt.Errorf("failed to write to output file: %w", err)
		}

		if err := os.Remove(filePath); err != nil {
			d.logger.Warn("Failed to remove file", "path", filePath, "error", err)
		}
	}

	return nil
}
