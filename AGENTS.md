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
go test -v ./service -run TestName  # Run specific test
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
    "crypto/aes"
    "fmt"
    "net/url"

    "github.com/go-resty/resty/v2"

    "m3u8-download/service"
)
```

### Naming
- Packages: lowercase, single words
- Exported: PascalCase (GetM3u8Bash, TsInfo)
- Private: camelCase (aesDecrypt, pKCS7UnPadding)
- Constants: UPPER_SNAKE_CASE
- Interfaces: -er suffix (Reader, Writer)

### Types & Structs
- Pointer receivers modify, value receivers read
- Exported types require comments
- Fields: PascalCase exported, camelCase private

```go
// TsInfo stores download URL and filename
type TsInfo struct {
    Name string
    Url  string
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
- `log.Fatal()` only in `main()`

```go
func GetM3u8Bash(urlStr string) (string, error) {
    u, err := url.Parse(urlStr)
    if err != nil {
        return "", err
    }
}
```

### Concurrency
- Use `sync.WaitGroup` for goroutine lifecycle
- Buffered channels limit concurrency
- Always `defer` cleanup, close channels

```go
var wg sync.WaitGroup
ch := make(chan struct{}, 15)
wg.Add(len(tsList))
for _, v := range tsList {
    ch <- struct{}{}
    go service.DownloadTs(v.Url, filePath, key, bar, &wg, ch)
}
wg.Wait()
```

### File Organization
- Package → imports → constants → variables → types → functions → main
- Group related functionality in service packages

### Comments
- `//` single-line, `/* */` block
- Exported declarations must have comments
- Explain "why" not "what"
- Use Chinese for consistency with existing codebase

### Testing
- Files: `*_test.go`, Functions: `TestFunctionName`
- Use `t.Run()` for subtests, table-driven for multiple cases
- Keep independent and deterministic

```go
func TestGetM3u8Bash(t *testing.T) {
    tests := []struct {
        name    string
        url     string
        want    string
        wantErr bool
    }{ /* cases */ }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) { /* logic */ })
    }
}
```

### Dependencies
- Go 1.22+
- Key: `resty/v2` (HTTP), `progressbar/v3`, `twinj/uuid`
- Update: `go get -u`, check vulns: `go list -json -m all`

### Best Practices
- `defer` for cleanup (file close, goroutine signaling)
- Avoid globals, pass dependencies as parameters
- Reuse HTTP client once (see `getClient()` in download.go)
- `filepath.Join()` for cross-platform paths
- Nil checks before pointer dereference

### Project Structure
```
m3u8-download/
├── main.go           # Entry point
├── go.mod/go.sum     # Dependencies
├── service/          # Business logic
│   ├── download.go   # Download/decryption
│   ├── file.go       # File operations
│   └── m3u8.go       # M3U8 parsing
└── cache/            # Generated temp files
```

## Notes

- CLI tool for M3U8 video stream downloads
- AES-encrypted stream support
- 15 concurrent goroutines for downloading
- Temp files in `cache/{uuid}/`, output is merged `.ts`
- Clean up temp files after merging
- Chinese comments/messages - maintain consistency
