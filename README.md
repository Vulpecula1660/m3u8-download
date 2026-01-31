# M3U8 下載工具

一個用 Go 語言開發的 M3U8 影片下載工具，支援 AES 加密的串流影片下載。

## 功能特點

- 支援 M3U8 格式影片下載
- 支援 AES 加密串流解密
- 使用 Goroutine 並行下載
- 自動合併分片
- 下載進度顯示
- 自動清理暫存檔案
- 智能重試機制（指數退避）
- 結構化日誌輸出
- 自訂 HTTP 請求選項

## 安裝需求

- Go 1.22 或以上版本

## 安裝步驟

1. 克隆此專案：
```bash
git clone https://github.com/yourusername/m3u8-download.git
```

2. 進入專案目錄：
```bash
cd m3u8-download
```

3. 安裝依賴：
```bash
go mod tidy
```

4. 建置程式：
```bash
go build -o m3u8-download.exe
```

## 使用方法

### 命令列參數

```bash
./m3u8-download -url <M3U8_URL> [選項]
```

### 可用選項

| 參數 | 說明 | 預設值 |
|------|------|--------|
| `-url` | M3U8 網址 (必填) | - |
| `-output` | 輸出檔名 | 自動生成 |
| `-workers` | 並發下載數量 | 15 |
| `-retries` | 重試次數 | 3 |
| `-timeout` | 請求逾時時間 (秒) | 30 |
| `-user-agent` | 自訂 User-Agent | 預設 UA |
| `-proxy` | Proxy 網址 | - |
| `-verbose` | 啟用詳細日誌 | false |

### 使用範例

#### 基本下載
```bash
./m3u8-download -url "https://example.com/video.m3u8"
```

#### 指定輸出檔名
```bash
./m3u8-download -url "https://example.com/video.m3u8" -output "output.ts"
```

#### 自訂並發數和重試次數
```bash
./m3u8-download -url "https://example.com/video.m3u8" -workers 20 -retries 5
```

#### 啟用詳細日誌
```bash
./m3u8-download -url "https://example.com/video.m3u8" -verbose
```

#### 使用 Proxy
```bash
./m3u8-download -url "https://example.com/video.m3u8" -proxy "http://127.0.0.1:7890"
```

## 專案架構

```
m3u8-download/
├── cmd/
│   └── m3u8-download/     # CLI 入口點
├── internal/
│   ├── config/            # 組態管理
│   ├── decrypt/           # AES 解密
│   ├── downloader/        # 下載邏輯
│   └── parser/            # M3U8 解析
├── pkg/
│   └── m3u8/              # 公共類型和錯誤
├── service/               # 向後相容層
├── test/                  # 測試資源
└── main.go                # 主程式
```

## 開發指南

### 執行測試

```bash
# 執行所有測試
go test ./...

# 執行特定套件測試
go test -v ./internal/parser

# 執行特定測試函數
go test -v ./internal/decrypt -run TestDecrypt

# 測試覆蓋率
go test -cover ./...
```

### 程式碼格式化

```bash
go fmt ./...
```

### 靜態分析

```bash
go vet ./...
```

## 依賴套件

- `github.com/go-resty/resty/v2` - HTTP 客戶端
- `github.com/schollz/progressbar/v3` - 進度條顯示
- `github.com/twinj/uuid` - UUID 生成

## 改進項目

### 已完成的改進
- ✅ 程式碼架構重構
- ✅ 改善錯誤處理
- ✅ 新增 CLI 參數
- ✅ 新增單元測試
- ✅ 進階下載控制（重試機制）
- ✅ 增強日誌輸出
- ✅ 自訂 HTTP 請求選項

### 未來計劃
- 進階下載控制（續傳功能）
- 整合測試
- 程式庫化

## 授權條款

本專案採用 MIT 授權條款 - 詳見 [LICENSE](LICENSE) 檔案
