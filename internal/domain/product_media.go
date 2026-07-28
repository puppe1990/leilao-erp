package domain

import (
	"fmt"
	"net/url"
	"strings"
)

// IsProductMediaKind reports whether kind is photo or video.
func IsProductMediaKind(kind string) bool {
	k := strings.ToLower(strings.TrimSpace(kind))
	return k == "photo" || k == "video"
}

// NormalizeProductMediaURL validates and trims a media reference.
// Accepts absolute https URLs or app-relative paths starting with /static/.
func NormalizeProductMediaURL(kind, raw string) (string, error) {
	if !IsProductMediaKind(kind) {
		return "", fmt.Errorf("kind must be photo or video")
	}
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", fmt.Errorf("url required")
	}
	// App-local static path (uploads)
	if strings.HasPrefix(s, "/static/") {
		if strings.Contains(s, "..") {
			return "", fmt.Errorf("invalid path")
		}
		return s, nil
	}
	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid url")
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", fmt.Errorf("url scheme must be http or https")
	}
	return u.String(), nil
}
