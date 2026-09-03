package domain

import "strings"

// CatalogThumbURL rewrites a local product photo URL to its .thumb.jpg sibling
// (generated next to the original for faster catalog loads). Falls back to the
// original when the path is not a local JPEG.
func CatalogThumbURL(photoURL string) string {
	u := strings.TrimSpace(photoURL)
	if u == "" {
		return ""
	}
	lower := strings.ToLower(u)
	if strings.Contains(lower, ".thumb.") {
		return u
	}
	// only rewrite app-local uploads
	if !strings.HasPrefix(u, "/static/uploads/") {
		return u
	}
	switch {
	case strings.HasSuffix(lower, ".jpg"):
		return u[:len(u)-4] + ".thumb.jpg"
	case strings.HasSuffix(lower, ".jpeg"):
		return u[:len(u)-5] + ".thumb.jpg"
	default:
		return u
	}
}
