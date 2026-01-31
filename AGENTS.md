# AGENTS.md

Guidelines for agentic coding assistants working on the m3u8-download Go project.

## Build Commands

```bash
go run main.go                    # Run program
go build -o m3u8-download.exe     # Build binary
go mod tidy                       # Install/update dependencies
go fmt ./...                      # Format code
gofmt -l .                        # Check files needing format
go vet ./...                      # Static analysis
go test ./...                     # Run all tests
go test -v ./internal/downloader -run TestHTTPClient_Get  # Run specific test
go test ./... -count=1            # Force tests to run (disable cache)
go test -cover ./...              # Test with coverage
```

## Code Style Guidelines

### Formatting
- Use `gofmt` (standard Go formatter) with tabs, 120 char max line
- Always run `go fmt ./...` before committing

### Imports
- Group: stdlib → external → local packages, alphabetical order
- Use blank lines between groups

```go
import (
    "bytes"
    "fmt"
    "sync"

    "github.com/schollz/progressbar/v3"

    "m3u8-download/internal/decrypt"
    "m3u8-download/pkg/m3u8"
)
```

### Naming
- Packages: lowercase, single words (config, parser, downloader, decrypt)
- Exported: PascalCase (DownloadSegments, TSInfo, Playlist)
- Private: camelCase (getBaseURL, extractSegments, pKCS7UnPadding)
- Constants: UPPER_SNAKE_CASE
- Interfaces: -er suffix (Reader, Writer)

### Types & Structs
- Pointer receivers modify state, value receivers read-only
- Exported types require comments
- Fields: PascalCase exported, camelCase private
- Use struct embedding for composition

```go
type TSInfo struct {
    Name string
    Url  string
}

type Playlist struct {
    BaseURL     string
    Key         string
    IV          []byte
    Segments    []*TSInfo
    IsEncrypted bool
}
```

### Functions
- Exported functions need comments
- Small, focused, single responsibility
- Max 4-5 parameters, use structs for more

### Error Handling
- Always check and handle errors
- Use `fmt.Errorf` for wrapping with context
- Return errors, don't panic (except main())
- Use package-level error vars in pkg/m3u8/errors.go
- Custom error types for specific errors (HTTPError, RetryExhaustedError)

```go
func ParsePlaylist(content, m3u8URL string) (*m3u8.Playlist, error) {
    baseURL, err := getBaseURL(m3u8URL)
    if err != nil {
        return nil, fmt.Errorf("failed to get base URL: %w", err)
    }
}

// Use defined error types
if httpErr, ok := err.(*m3u8.HTTPError); ok {
    if httpErr.StatusCode >= 400 && httpErr.StatusCode < 500 {
        return nil, err
    }
}
```

### Concurrency
- Use `sync.WaitGroup` for goroutine lifecycle
- Use `sync/atomic` (Int64) for counters instead of mutex
- Buffered channels limit concurrency
- Always `defer` cleanup, close channels

```go
var wg sync.WaitGroup
var completed atomic.Int64
ch := make(chan struct{}, workers)
wg.Add(len(playlist.Segments))
for _, seg := range playlist.Segments {
    ch <- struct{}{}
    go func(s *m3u8.TSInfo) {
        defer wg.Done()
        defer func() { <-ch }()
        // download logic
        completed.Add(1)
    }(segment)
}
wg.Wait()
```

### File Organization
- Package → imports → constants → variables → types → functions
- `internal/` for private packages not importable outside module
- `pkg/` for shared packages potentially importable by others
- Use `filepath.Join()` for cross-platform paths

### Comments
- `//` single-line, `/* */` block
- Exported declarations must have comments
- Explain "why" not "what"
- Use English for code, English comments preferred

### Testing
- Files: `*_test.go`, Functions: `TestFunctionName`
- Use `t.Run()` for subtests, table-driven for multiple cases
- Keep independent and deterministic
- Integration tests in main_test.go

```go
func TestParsePlaylist(t *testing.T) {
    tests := []struct {
        name    string
        content string
        url     string
        wantErr bool
    }{ /* cases */ }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) { /* logic */ })
    }
}
```

### Dependencies
- Go 1.22+
- Key: `progressbar/v3`, `twinj/uuid`
- Use stdlib: `net/http`, `log/slog`, `crypto/aes`, `crypto/cipher`
- Update: `go get -u`, check vulns: `go list -json -m all`

### Best Practices
- `defer` for cleanup (file close, goroutine signaling)
- Avoid globals, pass dependencies as parameters
- Reuse HTTP client (see NewHTTPClient)
- Use io.Copy/DownloadStream for large files instead of ReadAll
- For decryption, reuse cipher instances (see Decryptor in decrypt/aes.go)
- Nil checks before pointer dereference

### Performance Optimizations
- Use streaming I/O for large files (io.CopyBuffer)
- Use sync/atomic for high-frequency counters
- Reuse expensive resources (AES cipher, HTTP client)
- Minimize lock contention by keeping critical sections small

### Project Structure
```
m3u8-download/
├── main.go                  # Entry point
├── go.mod/go.sum            # Dependencies
├── internal/
│   ├── config/              # CLI config, cache dir management
│   ├── decrypt/             # AES decryption
│   ├── downloader/          # Download logic, HTTP client, file merging
│   └── parser/              # M3U8 playlist parsing
├── pkg/
│   └── m3u8/                # Shared types and errors
└── cache/                   # Generated temp files (ignored by git)
```

## Notes

- CLI tool for M3U8 video stream downloads
- AES-128 encrypted stream support
- Configurable concurrent workers (default: 15)
- Retry logic with exponential backoff
- Temp files in `cache/{uuid}/`, output is merged `.ts`
- Clean up temp files after merging
- Uses slog for structured logging
