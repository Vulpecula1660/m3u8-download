package parser

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"m3u8-download/pkg/m3u8"
)

func ParsePlaylist(content, m3u8URL string) (*m3u8.Playlist, error) {
	baseURL, err := getBaseURL(m3u8URL)
	if err != nil {
		return nil, fmt.Errorf("failed to get base URL: %w", err)
	}

	playlist := &m3u8.Playlist{
		BaseURL: baseURL,
	}

	lines := strings.Split(content, "\n")
	key, iv := extractEncryptionKey(baseURL, lines)
	playlist.Key = key
	playlist.IV = iv
	playlist.IsEncrypted = key != ""

	segments, err := extractSegments(baseURL, lines)
	if err != nil {
		return nil, err
	}

	if len(segments) == 0 {
		return nil, m3u8.ErrNoTSFiles
	}

	playlist.Segments = segments

	return playlist, nil
}

func getBaseURL(m3u8URL string) (string, error) {
	u, err := url.Parse(m3u8URL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	s := u.Scheme + "://" + u.Host + strings.ReplaceAll(filepath.Dir(u.EscapedPath()), "\\", "/")
	return s, nil
}

func extractEncryptionKey(baseURL string, lines []string) (string, []byte) {
	var key string
	var iv []byte

	for _, line := range lines {
		if strings.Contains(line, "#EXT-X-KEY") {
			uriPos := strings.Index(line, "URI")
			quotationMarkPos := strings.LastIndex(line, "\"")
			if uriPos == -1 || quotationMarkPos == -1 {
				continue
			}

			keyUrl := strings.Split(line[uriPos:quotationMarkPos], "\"")[1]

			if !strings.Contains(line, "http") {
				keyUrl = fmt.Sprintf("%s/%s", baseURL, keyUrl)
			}

			key = keyUrl

			if ivPos := strings.Index(line, "IV="); ivPos != -1 {
				ivStart := strings.Index(line[ivPos:], "0x")
				if ivStart != -1 {
					ivStr := line[ivPos+ivStart+2 : ivPos+ivStart+34]
					if len(ivStr) == 32 {
						iv = parseIV(ivStr)
					}
				}
			}

			break
		}
	}

	return key, iv
}

func parseIV(hex string) []byte {
	iv := make([]byte, 16)
	for i := 0; i < 32; i += 2 {
		b, err := fmt.Sscanf(hex[i:i+2], "%02x", &iv[i/2])
		if err != nil || b != 1 {
			return nil
		}
	}
	return iv
}

func extractSegments(baseURL string, lines []string) ([]*m3u8.TSInfo, error) {
	var segments []*m3u8.TSInfo
	index := 0

	for _, line := range lines {
		if !strings.HasPrefix(line, "#") && line != "" {
			index++
			ts := &m3u8.TSInfo{
				Name: fmt.Sprintf("%06d.ts", index),
			}

			if strings.HasPrefix(line, "http") {
				ts.Url = line
			} else {
				ts.Url = fmt.Sprintf("%s/%s", baseURL, line)
			}

			segments = append(segments, ts)
		}
	}

	return segments, nil
}
