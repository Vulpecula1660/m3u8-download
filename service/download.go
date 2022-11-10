package service

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	progressbar "github.com/schollz/progressbar/v3"
)

var client = getClient()

func getClient() *resty.Client {
	return resty.New().
		SetTimeout(15 * time.Second).
		SetRetryCount(5).
		SetRetryWaitTime(3 * time.Second).
		SetRetryMaxWaitTime(150 * time.Second)
}

func HttpGet(url string) ([]byte, error) {
	resp, err := client.R().Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("get %s fail status code: %d", url, resp.StatusCode())
	}

	return resp.Body(), nil
}

func DownloadTs(url string, filepath string, key string, bar *progressbar.ProgressBar, wg *sync.WaitGroup, ch chan struct{}) {
	defer func() {
		err := bar.Add(1)
		if err != nil {
			log.Fatal(err)
		}

		<-ch
		wg.Done()
	}()

	body, err := HttpGet(url)
	if err != nil {
		log.Fatal(fmt.Errorf("HttpGet error:%s", err))
	}

	if key != "" {
		body, err = aesDecrypt(body, []byte(key))
		if err != nil {
			log.Fatal(fmt.Errorf("aesDecrypt error:%s", err))
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

	file, err := os.Create(filepath)
	if err != nil {
		log.Fatal(fmt.Errorf("openFile error:%s", err))
	}

	defer file.Close()

	writer := bufio.NewWriter(file)

	_, err = writer.Write(body)
	if err != nil {
		log.Fatal(fmt.Errorf("write error:%s", err))
	}
}
