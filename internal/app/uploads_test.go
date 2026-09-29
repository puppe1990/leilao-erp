package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveUploadsDir_PrefersEnv(t *testing.T) {
	got := ResolveUploadsDir("/var/lib/leilao-erp/uploads", "/opt/app/data/app.db", "/opt/app/web/static")
	if got != "/var/lib/leilao-erp/uploads" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveUploadsDir_UsesDBSiblingWhenPresent(t *testing.T) {
	root := t.TempDir()
	dbPath := filepath.Join(root, "app.db")
	uploads := filepath.Join(root, "uploads")
	if err := os.Mkdir(uploads, 0o755); err != nil {
		t.Fatal(err)
	}
	got := ResolveUploadsDir("", dbPath, filepath.Join(root, "static"))
	if got != uploads {
		t.Fatalf("got %q want %q", got, uploads)
	}
}

func TestResolveUploadsDir_FallsBackToStaticUploads(t *testing.T) {
	staticDir := filepath.Join(t.TempDir(), "static")
	got := ResolveUploadsDir("", ":memory:", staticDir)
	want := filepath.Join(staticDir, "uploads")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
