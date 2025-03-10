# M3U8 下載工具

一個用 Go 語言開發的 M3U8 影片下載工具，支援 AES 加密的串流影片下載。

## 功能特點

- 支援 M3U8 格式影片下載
- 支援 AES 加密串流解密
- 使用 Goroutine 並行下載
- 自動合併分片
- 下載進度顯示
- 自動清理暫存檔案

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

## 使用方法

1. 開啟 `main.go` 檔案
2. 設定要下載的 M3U8 網址：
```go
func main() {
    urlStr := "你的 M3U8 網址"
    ...
}
```
3. 執行程式：
```bash
go run main.go
```

下載完成後會在當前目錄產生一個 .ts 檔案。

## 依賴套件

- github.com/go-resty/resty/v2 - HTTP 客戶端
- github.com/schollz/progressbar/v3 - 進度條顯示
- github.com/twinj/uuid - UUID 生成

## 授權條款

本專案採用 MIT 授權條款 - 詳見 [LICENSE](LICENSE) 檔案
