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
		"category":    {"ajuste"},
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

func TestCashHandler_CreateManual_Despesa(t *testing.T) {
	h, s := newCashHandler(t)
	accountID, err := s.InsertCashAccount("PIX principal", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{
		"account_id":  {fmt.Sprintf("%d", accountID)},
		"direction":   {"out"},
		"amount":      {"41,55"},
		"category":    {"despesa"},
		"memo":        {"Cabo HDMI VGA"},
		"occurred_at": {"2026-07-28"},
	}
	req := inertiaRequest(http.MethodPost, "/cash/entries", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.CreateManual(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}

	entries, err := s.ListCashEntries(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Category != "despesa" || entries[0].AmountCents != 4155 {
		t.Fatalf("entry=%+v", entries)
	}
}

func TestCashHandler_UpdateAndDeleteEntry(t *testing.T) {
	h, s := newCashHandler(t)
	accountID, err := s.InsertCashAccount("PIX principal", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}
	id, err := s.InsertManualCashEntry(accountID, "out", 1000, "2026-07-28T12:00:00Z", "despesa", "x")
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{
		"account_id":  {fmt.Sprintf("%d", accountID)},
		"direction":   {"out"},
		"amount":      {"50,00"},
		"category":    {"frete"},
		"memo":        {"atualizado"},
		"occurred_at": {"2026-07-29"},
	}
	req := inertiaRequest(http.MethodPost, fmt.Sprintf("/cash/entries/%d", id), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UpdateEntry(rr, req, id)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("update status=%d body=%s", rr.Code, rr.Body.String())
	}

	e, err := s.FindCashEntry(id)
	if err != nil {
		t.Fatal(err)
	}
	if e.AmountCents != 5000 || e.Category != "frete" || e.Memo == nil || *e.Memo != "atualizado" {
		t.Fatalf("updated=%+v", e)
	}

	reqDel := inertiaRequest(http.MethodPost, fmt.Sprintf("/cash/entries/%d/delete", id), nil)
	rrDel := httptest.NewRecorder()
	h.DestroyEntry(rrDel, reqDel, id)
	if rrDel.Code != http.StatusSeeOther {
		t.Fatalf("delete status=%d", rrDel.Code)
	}
	if _, err := s.FindCashEntry(id); err == nil {
		t.Fatal("expected not found after delete")
	}
}
