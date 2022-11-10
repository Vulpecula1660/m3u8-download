package service

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

// TsInfo 用於保存 ts 文件的下載地址和文件名
type TsInfo struct {
	Name string
	Url  string
}

// GetM3u8Bash 當 m3u8 中沒有完整連接時需要獲取 m3u8 文件前綴來拼接 ts 下載連接
func GetM3u8Bash(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	s := u.Scheme + "://" + u.Host + strings.ReplaceAll(filepath.Dir(u.EscapedPath()), "\\", "/")

	return s, nil
}

func GetM3u8Key(host, body string) (key string, err error) {
	lines := strings.Split(body, "\n")

	for _, line := range lines {
		if strings.Contains(line, "#EXT-X-KEY") {
			uriPos := strings.Index(line, "URI")
			quotationMarkPos := strings.LastIndex(line, "\"")
			keyUrl := strings.Split(line[uriPos:quotationMarkPos], "\"")[1]

			if !strings.Contains(line, "http") {
				keyUrl = fmt.Sprintf("%s/%s", host, keyUrl)
			}

			keyByte, err := HttpGet(keyUrl)
			if err != nil {
				return "", err
			}

			key = string(keyByte)
		}
	}

	return key, nil
}

func GetTsList(host, body, urlStr string) ([]*TsInfo, error) {
	lines := strings.Split(body, "\n")
	index := 0
	tsList := []*TsInfo{}

	for _, line := range lines {
		if !strings.HasPrefix(line, "#") && line != "" {
			index++
			var ts *TsInfo

			if strings.HasPrefix(line, "http") {
				ts = &TsInfo{
					Name: fmt.Sprintf("%06d.ts", index),
					Url:  line,
				}
			} else {
				ts = &TsInfo{
					Name: fmt.Sprintf("%06d.ts", index),
					Url:  fmt.Sprintf("%s/%s", host, line),
				}
			}

			tsList = append(tsList, ts)
		}
	}

	return tsList, nil
}

func aesDecrypt(crypted, key []byte, ivs ...[]byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	var iv []byte

	if len(ivs) == 0 {
		iv = key
	} else {
		iv = ivs[0]
	}

	blockMode := cipher.NewCBCDecrypter(block, iv[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = pKCS7UnPadding(origData)

	return origData, nil
}

func pKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
