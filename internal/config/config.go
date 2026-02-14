package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"m3u8-download/pkg/m3u8"
)

const (
	defaultWorkers   = 15
	defaultRetries   = 3
	defaultTimeout   = 30
	defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
)

// ParseMode indicates how CLI parsing should proceed in main.
type ParseMode int

const (
	// ParseModeRun indicates normal download execution should continue.
	ParseModeRun ParseMode = iota
	// ParseModeShowHelp indicates help was requested and already printed.
	ParseModeShowHelp
	// ParseModeShowVersion indicates version was requested.
	ParseModeShowVersion
)

// ParseFlags parses command-line arguments from os.Args for compatibility.
func ParseFlags() (*m3u8.DownloadConfig, error) {
	cfg, mode, err := ParseArgs(os.Args[1:], os.Stdout, os.Stderr)
	if err != nil {
		return nil, err
	}

	if mode == ParseModeShowHelp {
		return nil, flag.ErrHelp
	}
	if mode == ParseModeShowVersion {
		return nil, flag.ErrHelp
	}

	return cfg, nil
}

// ParseArgs parses CLI arguments and returns either config or help mode.
func ParseArgs(args []string, stdout, stderr io.Writer) (*m3u8.DownloadConfig, ParseMode, error) {
	if len(args) == 0 || args[0] == "help" {
		printHelp(stdout)
		return nil, ParseModeShowHelp, nil
	}

	var cfg m3u8.DownloadConfig
	var showVersion bool
	fs := newFlagSet(&cfg, &showVersion, stderr)

	err := fs.Parse(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printHelp(stdout)
			return nil, ParseModeShowHelp, nil
		}
		return nil, ParseModeRun, fmt.Errorf("%w；請使用 -h、--help 或 help 查看說明", err)
	}
	if showVersion {
		return nil, ParseModeShowVersion, nil
	}

	if cfg.URL == "" {
		return nil, ParseModeRun, fmt.Errorf("-url 參數為必填；請使用 -h、--help 或 help 查看說明")
	}

	if cfg.Workers <= 0 {
		cfg.Workers = defaultWorkers
	}

	if cfg.Retries <= 0 {
		cfg.Retries = defaultRetries
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultTimeout
	}

	if cfg.UserAgent == "" {
		cfg.UserAgent = defaultUserAgent
	}

	return &cfg, ParseModeRun, nil
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

func newFlagSet(cfg *m3u8.DownloadConfig, showVersion *bool, stderr io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet("m3u8-download", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {}

	fs.StringVar(&cfg.URL, "url", "", "M3U8 URL（必填）")
	fs.StringVar(&cfg.Output, "output", "", "輸出檔名（.ts）")
	fs.IntVar(&cfg.Workers, "workers", defaultWorkers, "並發下載數量")
	fs.IntVar(&cfg.Retries, "retries", defaultRetries, "重試次數")
	fs.IntVar(&cfg.Timeout, "timeout", defaultTimeout, "請求逾時秒數")
	fs.StringVar(&cfg.UserAgent, "user-agent", "", "自訂 User-Agent")
	fs.BoolVar(&cfg.Verbose, "verbose", false, "啟用詳細日誌")
	fs.StringVar(&cfg.ProxyURL, "proxy", "", "Proxy 網址")
	fs.StringVar(&cfg.Origin, "origin", "", "HTTP Origin header")
	fs.StringVar(&cfg.Referer, "referer", "", "HTTP Referer header")
	fs.BoolVar(showVersion, "version", false, "顯示版本資訊")

	return fs
}

func printHelp(stdout io.Writer) {
	_, _ = fmt.Fprintf(stdout, `M3U8 下載工具

用法：
  m3u8-download -url <M3U8_URL> [選項]
  m3u8-download help

選項：
  -url string
        M3U8 URL（必填）
  -output string
        輸出檔名（.ts），未提供時自動以 UUID 命名
  -workers int
        並發下載數量（預設 %d）
  -retries int
        重試次數（預設 %d）
  -timeout int
        請求逾時秒數（預設 %d）
  -user-agent string
        自訂 User-Agent
  -proxy string
        Proxy 網址
  -origin string
        HTTP Origin header
  -referer string
        HTTP Referer header
  -verbose
        啟用詳細日誌
  -version, --version
        顯示版本資訊
  -h, --help
        顯示說明

範例：
  m3u8-download -url "https://example.com/video.m3u8"
  m3u8-download -url "https://example.com/video.m3u8" -output "video.ts"
  m3u8-download --version
  m3u8-download help
`, defaultWorkers, defaultRetries, defaultTimeout)
}
