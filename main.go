package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"m3u8-download/service"

	progressbar "github.com/schollz/progressbar/v3"
	"github.com/twinj/uuid"
)

func main() {
	urlStr := ""

	bodyByte, err := service.HttpGet(urlStr)
	if err != nil {
		log.Fatal(fmt.Errorf("HttpGet error:%s", err))
	}

	body := string(bodyByte)

	host, err := service.GetM3u8Bash(urlStr)
	if err != nil {
		log.Fatal(fmt.Errorf("GetM3u8Bash error:%s", err))
	}

	key, err := service.GetM3u8Key(host, body)
	if err != nil {
		log.Fatal(fmt.Errorf("GetM3u8Key error:%s", err))
	}

	tsList, err := service.GetTsList(host, body, urlStr)
	if err != nil {
		log.Fatal(fmt.Errorf("GetTsList error:%s", err))
	}

	id := uuid.NewV4().String()

	err = service.InitCachePath(id)
	if err != nil {
		log.Fatal(fmt.Errorf("InitCachePath error:%s", err))
	}

	WorkPath, err := os.Getwd()
	if err != nil {
		log.Fatal(fmt.Errorf(" Getwd error:%s", err))
	}

	tsPath := filepath.Join(WorkPath, "cache", id)
	bar := progressbar.Default(int64(len(tsList)))
	var wg sync.WaitGroup
	ch := make(chan struct{}, 15)

	wg.Add(len(tsList))
	for _, v := range tsList {
		ch <- struct{}{}
		FilePath := tsPath + "/" + v.Name
		go service.DownloadTs(v.Url, FilePath, key, bar, &wg, ch)
	}

	wg.Wait()

	err = service.MergeFile(tsPath, id+".ts")
	if err != nil {
		log.Fatal(fmt.Errorf(" MergeFile error:%s", err))
	}

	fmt.Println("下載完成")
}
