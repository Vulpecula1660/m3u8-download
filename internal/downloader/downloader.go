package downloader

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"

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
	var decryptor *decrypt.Decryptor

	var completed atomic.Int64
	var failed atomic.Int64

	if playlist.IsEncrypted && playlist.Key != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			keyData, keyDataErr = d.httpClient.Get(playlist.Key)
			if keyDataErr == nil {
				decryptor, keyDataErr = decrypt.NewDecryptor(keyData, nil)
			}
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

			err := d.downloadSegment(seg.Url, filePath, playlist.IsEncrypted, decryptor, keyDataLoaded, &mu)
			if err != nil {
				d.logger.Error("Failed to download segment", "index", idx, "url", seg.Url, "error", err)
				failed.Add(1)
				return
			}

			completed.Add(1)
		}(i, segment)
	}

	wg.Wait()

	if keyDataErr != nil {
		return stats, fmt.Errorf("failed to download encryption key: %w", keyDataErr)
	}

	stats.Completed = int(completed.Load())
	stats.Failed = int(failed.Load())

	return stats, nil
}

func (d *Downloader) downloadSegment(url, filePath string, isEncrypted bool, decryptor *decrypt.Decryptor, keyLoaded bool, mu *sync.Mutex) error {
	if !isEncrypted {
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer file.Close()

		buf := new(bytes.Buffer)
		if err := d.httpClient.DownloadStream(url, buf); err != nil {
			return err
		}

		data := decrypt.RemoveSyncBytePrefix(buf.Bytes())
		_, err = file.Write(data)
		return err
	}

	mu.Lock()
	ready := keyLoaded
	mu.Unlock()

	if !ready {
		return fmt.Errorf("encryption key not loaded")
	}

	data, err := d.httpClient.Get(url)
	if err != nil {
		return err
	}

	mu.Lock()
	decrypted, err := decryptor.Decrypt(data)
	mu.Unlock()

	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	data = decrypt.RemoveSyncBytePrefix(decrypted)

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
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

	buf := make([]byte, 32*1024)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := fmt.Sprintf("%s/%s", cacheDir, entry.Name())
		inFile, err := os.Open(filePath)
		if err != nil {
			d.logger.Warn("Failed to open file", "path", filePath, "error", err)
			continue
		}

		if _, err := io.CopyBuffer(outFile, inFile, buf); err != nil {
			inFile.Close()
			return fmt.Errorf("failed to copy file: %w", err)
		}

		inFile.Close()

		if err := os.Remove(filePath); err != nil {
			d.logger.Warn("Failed to remove file", "path", filePath, "error", err)
		}
	}

	return nil
}
