# M3U8 下載工具

一個用 Go 語言開發的 M3U8 影片下載工具，支援 AES 加密的串流影片下載。

## 功能特點

- 支援 M3U8 格式影片下載
- 支援 AES-128 加密串流解密
- 可配置並發下載（預設：15 個 worker）
- 智能重試機制（指數退避）
- 下載進度顯示
- 自動合併分片檔案
- 自動清理暫存檔案
- 結構化日誌輸出
- 可自訂 HTTP 請求選項
- 支援 `-version` / `--version` 查詢版本資訊

## 安裝需求

- 直接使用 Release 二進位：不需要 Go 環境
- 從原始碼建置：Go 1.22 或以上版本

## 從 Releases 下載（建議）

1. 前往 [Releases](../../releases) 頁面下載對應平台壓縮包。
2. 解壓後即可執行 `m3u8-download`（Windows 為 `m3u8-download.exe`）。
3. 可搭配 `checksums.txt` 驗證下載完整性。

目前提供：
- `linux/amd64`
- `linux/arm64`
- `windows/amd64`
- `darwin/arm64`

## 從原始碼建置

```bash
# 安裝依賴
go mod tidy

# 建置程式
go build -o m3u8-download
```

## 使用方法

### 命令列參數

```bash
./m3u8-download -url <M3U8_URL> [選項]
./m3u8-download -h
./m3u8-download --help
./m3u8-download --version
./m3u8-download help
```

未提供任何參數時，程式會直接顯示 help 說明並結束。

### 可用選項

| 參數 | 說明 | 預設值 |
|------|------|--------|
| `-url` | M3U8 網址 (必填) | - |
| `-output` | 輸出檔名 (.ts) | 自動生成 UUID |
| `-workers` | 並發下載數量 | 15 |
| `-retries` | 重試次數 | 3 |
| `-timeout` | 請求逾時時間 (秒) | 30 |
| `-user-agent` | 自訂 User-Agent | 預設瀏覽器 UA |
| `-proxy` | Proxy 網址 | - |
| `-origin` | HTTP Origin header | - |
| `-referer` | HTTP Referer header | - |
| `-verbose` | 啟用詳細日誌 | false |
| `-version`, `--version` | 顯示版本資訊 | - |
| `-h`, `--help` | 顯示 help 說明 | - |

### 使用範例

#### 基本下載
```bash
./m3u8-download -url "https://example.com/video.m3u8"
```

#### 指定輸出檔名
```bash
./m3u8-download -url "https://example.com/video.m3u8" -output "video.ts"
```

#### 自訂並發數和重試次數
```bash
./m3u8-download -url "https://example.com/video.m3u8" -workers 20 -retries 5
```

#### 啟用詳細日誌
```bash
./m3u8-download -url "https://example.com/video.m3u8" -verbose
```

#### 顯示 help
```bash
./m3u8-download help
```

#### 顯示版本
```bash
./m3u8-download --version
```

#### 使用 Proxy
```bash
./m3u8-download -url "https://example.com/video.m3u8" -proxy "http://127.0.0.1:7890"
```

#### 指定 HTTP Headers
```bash
./m3u8-download -url "https://example.com/video.m3u8" -origin "https://example.com" -referer "https://example.com/video"
```

## 專案架構

```
m3u8-download/
├── .github/workflows/
│   └── release.yml           # Tag 觸發自動發布 GitHub Releases
├── .goreleaser.yml           # GoReleaser 發版設定
├── main.go                  # 程式入口點
├── go.mod/go.sum            # 依賴管理
├── internal/
│   ├── config/              # CLI 參數解析、快取目錄管理
│   ├── decrypt/             # AES-128 解密實作
│   ├── downloader/          # 下載邏輯、HTTP 客戶端、檔案合併
│   └── parser/              # M3U8 播放清單解析
├── pkg/
│   └── m3u8/                # 共享類型和錯誤定義
└── cache/                   # 生成的暫存檔案 (被 git 忽略)
```

## 開發指南

### 執行測試

```bash
# 執行所有測試
go test ./...

# 執行特定套件測試
go test -v ./internal/parser

# 執行特定測試函數
go test -v ./internal/downloader -run TestHTTPClient_Get

# 測試覆蓋率
go test -cover ./...

# 強制重新執行測試（不使用快取）
go test ./... -count=1
```

### 程式碼格式化

```bash
go fmt ./...
```

### 靜態分析

```bash
go vet ./...
```

### 發版（GitHub Releases）

推送語義化版本標籤後，GitHub Actions 會自動建立 Release 並上傳多平台二進位：

```bash
git tag v1.0.0
git push origin v1.0.0
```

## 依賴套件

- `github.com/schollz/progressbar/v3` - 進度條顯示
- `github.com/twinj/uuid` - UUID 生成

標準函式庫：
- `net/http` - HTTP 客戶端
- `log/slog` - 結構化日誌
- `crypto/aes`, `crypto/cipher` - AES 加密解密

## 授權條款

本專案採用 MIT 授權條款 - 詳見 [LICENSE](LICENSE) 檔案
