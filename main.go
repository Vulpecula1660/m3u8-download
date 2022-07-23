package main

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	progressbar "github.com/schollz/progressbar/v3"
	"github.com/twinj/uuid"
)

// TsInfo 用于保存 ts 文件的下载地址和文件名
type TsInfo struct {
	Name string
	Url  string
}

func main() {
	urlStr := "https://videos-fms.jwpsrv.com/0_62dc60ac_0x367586ff1a48295a012a9daf8ee879cfb1c5ac09/content/conversions/LOPLPiDX/videos/IFBsp7yL-24721151.mp4.m3u8"

	body, err := HttpGet(urlStr)
	if err != nil {
		fmt.Println("HttpGet error:", err)
		log.Fatal(err)
	}

	bash, err := GetM3u8Bash(urlStr)
	if err != nil {
		fmt.Println("GetM3u8Bash error:", err)
		log.Fatal(err)
	}

	key, err := GetM3u8Key(bash, body)
	if err != nil {
		fmt.Println("GetM3u8Key error:", err)
		log.Fatal(err)
	}

	tsList, err := GetTsList(body, urlStr)
	if err != nil {
		fmt.Println("GetTsList error:", err)
		log.Fatal(err)
	}

	id := uuid.NewV4().String()

	err = InitCachePath(id)
	if err != nil {
		fmt.Println("InitCachePath error:", err)
		log.Fatal(err)
	}

	WorkPath, err := os.Getwd()
	if err != nil {
		fmt.Println("Getwd error:", err)
		log.Fatal(err)
	}

	tsPath := WorkPath + "/cache/" + id
	bar := progressbar.Default(int64(len(tsList)))
	var wg sync.WaitGroup

	wg.Add(len(tsList))
	for _, v := range tsList {
		FilePath := tsPath + "/" + v.Name
		go DownloadTs(v.Url, FilePath, key, bar, &wg)
	}

	wg.Wait()

	err = MergeFile(tsPath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("下載完成")
}

func HttpGet(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func GetM3u8Bash(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	host := u.Scheme + "://" + u.Host + strings.ReplaceAll(filepath.Dir(u.EscapedPath()), "\\", "/")
	return host, nil
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
			key, err = HttpGet(keyUrl)
			if err != nil {
				return "", err
			}
		}
	}
	return key, nil
}

func GetTsList(body, urlStr string) ([]TsInfo, error) {
	bash, err := GetM3u8Bash(urlStr)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(body, "\n")
	index := 0
	tsList := []TsInfo{}
	for _, line := range lines {
		if !strings.HasPrefix(line, "#") && line != "" {
			index++
			var ts TsInfo

			if strings.HasPrefix(line, "http") {
				ts = TsInfo{
					Name: fmt.Sprintf("%06d.ts", index),
					Url:  line,
				}
			} else {
				ts = TsInfo{
					Name: fmt.Sprintf("%06d.ts", index),
					Url:  fmt.Sprintf("%s/%s", bash, line),
				}
			}
			tsList = append(tsList, ts)
		}
	}
	return tsList, nil
}

func InitCachePath(id string) error {
	WorkPath, _ := os.Getwd()
	err := os.MkdirAll(WorkPath+"/cache/"+id, os.ModePerm)
	if err != nil {
		return err
	}
	// os.Remove(dl.Filename)
	return nil
}

func DownloadTs(url string, filepath string, key string, bar *progressbar.ProgressBar, wg *sync.WaitGroup) {
	defer wg.Done()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(url)
		fmt.Println("HttpGet error:", err)
		log.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("HttpGet error:", err)
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("HttpGet error:", err)
		log.Fatal(err)
	}
	if key != "" {
		body, err = AesDecrypt(body, []byte(key))
		if err != nil {
			fmt.Println("AesDecrypt error:", err)
			log.Fatal(err)
		}
	}
	// https://en.wikipedia.org/wiki/MPEG_transport_stream
	// Some TS files do not start with SyncByte 0x47, they can not be played after merging,
	// Need to remove the bytes before the SyncByte 0x47(71).
	syncByte := uint8(71) // 0x47
	bLen := len(body)
	for j := 0; j < bLen; j++ {
		if body[j] == syncByte {
			body = body[j:]
			break
		}
	}
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0o666)
	if err != nil {
		fmt.Println("OpenFile error:", err)
		log.Fatal(err)
	}
	defer file.Close()
	_, err = file.Write(body)
	if err != nil {
		fmt.Println("Write error:", err)
		log.Fatal(err)
	}
	bar.Add(1)
}

func AesDecrypt(crypted, key []byte, ivs ...[]byte) ([]byte, error) {
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
	origData = PKCS7UnPadding(origData)
	return origData, nil
}

func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func MergeFile(tsPath string) error {
	// TODO:
	outFile, err := os.OpenFile("test2",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer outFile.Close()
	tsFileList, _ := ioutil.ReadDir(tsPath)
	for _, f := range tsFileList {
		tsFilePath := tsPath + "/" + f.Name()
		tsFileContent, err := ioutil.ReadFile(tsFilePath)
		if err != nil {
			return err
		}
		if _, err := outFile.Write(tsFileContent); err != nil {
			return err
		}
		if err = os.Remove(tsFilePath); err != nil {
			return err
		}
	}
	// 删除ts缓存目录
	err = os.RemoveAll(tsPath)
	if err != nil {
		return err
	}
	// 删除缓存目录
	WorkPath, _ := os.Getwd()
	err = os.RemoveAll(WorkPath + "/cache")
	if err != nil {
		return err
	}

	return nil
}
