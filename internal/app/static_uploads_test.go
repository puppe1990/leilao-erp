package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"
	inertia "github.com/romsar/gonertia/v3"

	"github.com/puppe1990/leilao-erp/internal/store"
)

func TestStaticUploadsServedFromUploadsDir(t *testing.T) {
	s, err := store.NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	i, err := inertia.New(`<!DOCTYPE html><html><body>{{ .inertia }}</body></html>`)
	if err != nil {
		t.Fatal(err)
	}

	staticDir := t.TempDir()
	uploadsDir := t.TempDir()
	productDir := filepath.Join(uploadsDir, "products", "12")
	if err := os.MkdirAll(productDir, 0o755); err != nil {
		t.Fatal(err)
	}
	photo := filepath.Join(productDir, "philips-front-r2.jpg")
	if err := os.WriteFile(photo, []byte("jpeg-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}

	a, err := New(cais.Config{Env: "test"}, Deps{
		Renderer:   cais.NewRendererStub(i18n.DefaultCatalog()),
		Store:      s,
		StaticDir:  staticDir,
		UploadsDir: uploadsDir,
		Site:       meta.Site{AppName: "leilao-erp", AppURL: "http://localhost"},
		Catalog:    i18n.DefaultCatalog(),
		Inertia:    i,
	})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/static/uploads/products/12/philips-front-r2.jpg", nil)
	rr := httptest.NewRecorder()
	a.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if rr.Body.String() != "jpeg-bytes" {
		t.Fatalf("body=%q", rr.Body.String())
	}
}
