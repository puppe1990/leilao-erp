package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"
	inertia "github.com/romsar/gonertia/v3"

	"github.com/puppe1990/leilao-erp/internal/store"
)

// TestSignupRoutesNotRegistered ensures public signup is disabled.
// Go ServeMux treats "GET /" as a catch-all, so GET /signup falls through to the
// public catalog rather than 404 — we assert it never renders Signup, and that POST has no handler.
func TestSignupRoutesNotRegistered(t *testing.T) {
	s, err := store.NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })

	i, err := inertia.New(`<!DOCTYPE html><html><body>{{ .inertia }}</body></html>`)
	if err != nil {
		t.Fatal(err)
	}

	a, err := New(cais.Config{Env: "test"}, Deps{
		Renderer: cais.NewRendererStub(i18n.DefaultCatalog()),
		Store:    s,
		Site:     meta.Site{AppName: "leilao-erp", AppURL: "http://localhost"},
		Catalog:  i18n.DefaultCatalog(),
		Inertia:  i,
	})
	if err != nil {
		t.Fatal(err)
	}
	h := a.Handler()

	// GET /signup must not render a Signup page (public signup removed).
	getReq := httptest.NewRequest(http.MethodGet, "/signup", nil)
	getReq.Header.Set("X-Inertia", "true")
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)
	if getRR.Code == http.StatusOK {
		var payload map[string]any
		if err := json.Unmarshal(getRR.Body.Bytes(), &payload); err != nil {
			t.Fatalf("GET /signup body not inertia json: %v body=%s", err, getRR.Body.String())
		}
		if payload["component"] == "Signup" {
			t.Fatal("GET /signup rendered Signup; public signup must stay disabled")
		}
	}

	// POST /signup must not be a registered success endpoint.
	postReq := httptest.NewRequest(http.MethodPost, "/signup", nil)
	postRR := httptest.NewRecorder()
	h.ServeHTTP(postRR, postReq)
	switch postRR.Code {
	case http.StatusOK, http.StatusCreated, http.StatusSeeOther, http.StatusFound:
		t.Errorf("POST /signup status = %d, want non-success (signup must not be registered)", postRR.Code)
	}
}
