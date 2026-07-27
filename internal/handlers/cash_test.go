package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/leilao-erp/internal/store"
)

func newCashHandler(t *testing.T) (*CashHandler, *store.SQLiteStore) {
	t.Helper()
	s, err := store.NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	h := NewCashHandler(setupTestRenderer(t), s, testSite(), cais.Config{}, setupTestInertia(t))
	return h, s
}

func TestCashHandler_Index_OK(t *testing.T) {
	h, s := newCashHandler(t)
	if _, err := s.InsertCashAccount("PIX principal", "pix", 0); err != nil {
		t.Fatal(err)
	}

	req := inertiaRequest(http.MethodGet, "/cash", nil)
	rr := httptest.NewRecorder()
	h.Index(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	assertInertiaComponent(t, rr, "Cash/Index")
	assertInertiaProp(t, rr, "balances")
	assertInertiaProp(t, rr, "entries")
	assertInertiaProp(t, rr, "cashAccounts")
}

func TestCashHandler_CreateManual_Redirects(t *testing.T) {
	h, s := newCashHandler(t)
	accountID, err := s.InsertCashAccount("PIX principal", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{
		"account_id":  {fmt.Sprintf("%d", accountID)},
		"direction":   {"in"},
		"amount":      {"100,00"},
		"memo":        {"Ajuste de teste"},
		"occurred_at": {"2026-07-25"},
	}
	req := inertiaRequest(http.MethodPost, "/cash/entries", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.CreateManual(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303 body=%s", rr.Code, rr.Body.String())
	}
	if loc := rr.Header().Get("Location"); loc != "/cash" {
		t.Errorf("Location = %q, want /cash", loc)
	}

	entries, err := s.ListCashEntries(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	if entries[0].Category != "ajuste" {
		t.Errorf("category = %s, want ajuste", entries[0].Category)
	}
	if entries[0].AmountCents != 10000 {
		t.Errorf("amount = %d, want 10000", entries[0].AmountCents)
	}
	if entries[0].Direction != "in" {
		t.Errorf("direction = %s, want in", entries[0].Direction)
	}
}
