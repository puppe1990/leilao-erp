package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/session"

	"github.com/puppe1990/leilao-erp/internal/store"
)

func setupConfigHandler(t *testing.T) (*ConfigHandler, *store.SQLiteStore) {
	t.Helper()
	st, err := store.NewSQLiteStore(filepath.Join(t.TempDir(), "cfg.db"), "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	h := NewConfigHandler(setupTestRenderer(t), st, testSite(), cais.Config{Env: "development"}, setupTestInertia(t))
	return h, st
}

func withUserSession(t *testing.T, st *store.SQLiteStore, req *http.Request) *http.Request {
	t.Helper()
	user, err := st.FindUserByEmail("admin@leilao.local")
	if err != nil {
		t.Fatal(err)
	}
	// Sign in via sessions store so UserID middleware state is set
	rr := httptest.NewRecorder()
	if err := session.SignIn(rr, st.Sessions(), req, user.ID, session.CookieOptions{}); err != nil {
		t.Fatal(err)
	}
	// Copy session cookie to request and load session into context
	for _, c := range rr.Result().Cookies() {
		req.AddCookie(c)
	}
	// LoadSession-style: use WithUserID if available
	return session.WithUserID(req, user.ID)
}

func TestConfig_Index_requiresAuth(t *testing.T) {
	h, _ := setupConfigHandler(t)
	req := inertiaRequest(http.MethodGet, "/config", nil)
	rr := httptest.NewRecorder()
	h.Index(rr, req)
	if rr.Code != http.StatusSeeOther && rr.Code != http.StatusFound {
		// unauthenticated may redirect
		if rr.Code == http.StatusOK {
			t.Fatal("expected redirect without session")
		}
	}
}

func TestConfig_UpdateCompany(t *testing.T) {
	h, st := setupConfigHandler(t)
	form := url.Values{}
	form.Set("company_name", "Puppe Leilões")
	req := inertiaRequest(http.MethodPost, "/config/company", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withUserSession(t, st, req)

	rr := httptest.NewRecorder()
	h.UpdateCompany(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	name, err := st.CompanyName()
	if err != nil || name != "Puppe Leilões" {
		t.Fatalf("company=%q err=%v", name, err)
	}
}

func TestConfig_UpdatePassword(t *testing.T) {
	h, st := setupConfigHandler(t)
	form := url.Values{}
	form.Set("current_password", "change-me-now")
	form.Set("new_password", "senha-nova-123")
	form.Set("new_password_confirmation", "senha-nova-123")
	req := inertiaRequest(http.MethodPost, "/config/password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withUserSession(t, st, req)

	rr := httptest.NewRecorder()
	h.UpdatePassword(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}

	user, err := st.FindUserByEmail("admin@leilao.local")
	if err != nil {
		t.Fatal(err)
	}
	if !session.VerifyPassword(user.PasswordHash, "senha-nova-123") {
		t.Fatal("password not updated")
	}
}

func TestConfig_UpdatePassword_wrongCurrent(t *testing.T) {
	h, st := setupConfigHandler(t)
	form := url.Values{}
	form.Set("current_password", "errada")
	form.Set("new_password", "senha-nova-123")
	form.Set("new_password_confirmation", "senha-nova-123")
	req := inertiaRequest(http.MethodPost, "/config/password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withUserSession(t, st, req)

	rr := httptest.NewRecorder()
	h.UpdatePassword(rr, req)
	assertInertiaErrors(t, rr, "current_password")
}
