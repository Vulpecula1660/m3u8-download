package parser

import (
	"testing"
)

func TestParsePlaylist(t *testing.T) {
	tests := []struct {
		name    string
		content string
		url     string
		wantErr bool
	}{
		{
			name: "simple playlist",
			content: `#EXTM3U
#EXT-X-VERSION:3
#EXTINF:10.0,
segment1.ts
#EXTINF:10.0,
segment2.ts
#EXT-X-ENDLIST`,
			url:     "http://example.com/video.m3u8",
			wantErr: false,
		},
		{
			name:    "empty playlist",
			content: "#EXTM3U",
			url:     "http://example.com/video.m3u8",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			playlist, err := ParsePlaylist(tt.content, tt.url)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if playlist == nil {
				t.Error("playlist is nil")
				return
			}

			if playlist.BaseURL == "" {
				t.Error("BaseURL is empty")
			}

			if len(playlist.Segments) == 0 {
				t.Error("no segments found")
			}
		})
	}
}

func TestExtractEncryptionKey(t *testing.T) {
	tests := []struct {
		name    string
		content string
		baseURL string
		wantKey bool
	}{
		{
			name: "no encryption",
			content: `#EXTM3U
#EXTINF:10.0,
segment1.ts`,
			baseURL: "http://example.com",
			wantKey: false,
		},
		{
			name: "with encryption",
			content: `#EXTM3U
#EXT-X-KEY:METHOD=AES-128,URI="key.key"
#EXTINF:10.0,
segment1.ts`,
			baseURL: "http://example.com",
			wantKey: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := splitLines(tt.content)
			key, _ := extractEncryptionKey(tt.baseURL, lines)

			if tt.wantKey && key == "" {
				t.Error("expected key but got empty")
			}

			if !tt.wantKey && key != "" {
				t.Error("expected no key but got one")
			}
		})
	}
}

func TestExtractSegments(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		baseURL   string
		wantCount int
	}{
		{
			name: "two segments",
			content: `#EXTM3U
#EXTINF:10.0,
segment1.ts
#EXTINF:10.0,
segment2.ts`,
			baseURL:   "http://example.com",
			wantCount: 2,
		},
		{
			name: "absolute URLs",
			content: `#EXTM3U
#EXTINF:10.0,
http://cdn.example.com/segment1.ts`,
			baseURL:   "http://example.com",
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := splitLines(tt.content)
			segments, err := extractSegments(tt.baseURL, lines)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(segments) != tt.wantCount {
				t.Errorf("got %d segments, want %d", len(segments), tt.wantCount)
			}

			for i, seg := range segments {
				if seg.Name == "" {
					t.Errorf("segment %d has empty name", i)
				}
				if seg.Url == "" {
					t.Errorf("segment %d has empty URL", i)
				}
			}
		})
	}
}

func splitLines(content string) []string {
	lines := make([]string, 0)
	start := 0
	for i, c := range content {
		if c == '\n' {
			lines = append(lines, content[start:i])
			start = i + 1
		}
	}
	if start < len(content) {
		lines = append(lines, content[start:])
	}
	return lines
}

func TestGetBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name:    "simple URL",
			url:     "http://example.com/video.m3u8",
			want:    "http://example.com/",
			wantErr: false,
		},
		{
			name:    "URL with path",
			url:     "http://example.com/path/to/video.m3u8",
			want:    "http://example.com/path/to",
			wantErr: false,
		},
		{
			name:    "invalid URL",
			url:     "://invalid",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getBaseURL(tt.url)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
