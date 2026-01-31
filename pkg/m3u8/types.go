package m3u8

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

type DownloadConfig struct {
	URL          string
	Output       string
	Workers      int
	Retries      int
	Timeout      int
	UserAgent    string
	Verbose      bool
	ProxyURL     string
	Origin       string
	Referer      string
	CustomHeader map[string]string
}

type DownloadStats struct {
	Total     int
	Completed int
	Failed    int
	StartTime int64
	EndTime   int64
}
