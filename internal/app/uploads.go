package app

import (
	"os"
	"path/filepath"
	"strings"
)

// ResolveUploadsDir picks the persistent product-media directory.
// UPLOADS_DIR wins; otherwise dirname(DB_PATH)/uploads when that folder
// already exists (Lightsail/Cleat layout); otherwise web/static/uploads.
func ResolveUploadsDir(uploadsEnv, dbPath, staticDir string) string {
	if s := strings.TrimSpace(uploadsEnv); s != "" {
		return s
	}
	if dbPath != "" && dbPath != ":memory:" {
		sibling := filepath.Join(filepath.Dir(dbPath), "uploads")
		if st, err := os.Stat(sibling); err == nil && st.IsDir() {
			return sibling
		}
	}
	if staticDir != "" {
		return filepath.Join(staticDir, "uploads")
	}
	return ""
}
