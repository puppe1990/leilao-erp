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

func newLotsHandler(t *testing.T) (*LotsHandler, *store.SQLiteStore) {
	t.Helper()
	s, err := store.NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	h := NewLotsHandler(setupTestRenderer(t), s, testSite(), cais.Config{}, setupTestInertia(t))
	return h, s
}

func TestLotsHandler_Index_Inertia(t *testing.T) {
	h, _ := newLotsHandler(t)

	req := inertiaRequest(http.MethodGet, "/lots", nil)
	rr := httptest.NewRecorder()
	h.Index(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	assertInertiaComponent(t, rr, "Lots/Index")
	assertInertiaProp(t, rr, "lots")
}

func TestLotsHandler_Create_Valid_Redirects(t *testing.T) {
	h, s := newLotsHandler(t)

	accountID, err := s.InsertCashAccount("PIX principal", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{
		"name":            {"Monitores — leilão Jul/2026"},
		"purchased_at":    {"2026-07-20"},
		"item_title":      {"Monitor"},
		"item_qty":        {"22"},
		"cost_label":      {"Arremate"},
		"cost_amount":     {"603,00"},
		"already_paid":    {"true"},
		"cash_account_id": {fmt.Sprintf("%d", accountID)},
	}
	req := inertiaRequest(http.MethodPost, "/lots", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303 body=%s", rr.Code, rr.Body.String())
	}
	loc := rr.Header().Get("Location")
	if !strings.HasPrefix(loc, "/lots/") {
		t.Errorf("Location = %q, want /lots/{id}", loc)
	}

	lots, err := s.ListLots()
	if err != nil {
		t.Fatal(err)
	}
	if len(lots) != 1 {
		t.Fatalf("lots = %d, want 1", len(lots))
	}
	if lots[0].ItemCount != 22 {
		t.Errorf("item count = %d, want 22", lots[0].ItemCount)
	}
	if lots[0].TotalCostCents != 60300 {
		t.Errorf("total cost = %d, want 60300", lots[0].TotalCostCents)
	}
}

func TestLotsHandler_Create_QtyZero_Validation(t *testing.T) {
	h, s := newLotsHandler(t)

	accountID, err := s.InsertCashAccount("PIX principal", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{
		"name":            {"Lote inválido"},
		"purchased_at":    {"2026-07-20"},
		"item_title":      {"Item"},
		"item_qty":        {"0"},
		"cost_label":      {"Arremate"},
		"cost_amount":     {"100,00"},
		"already_paid":    {"false"},
		"cash_account_id": {fmt.Sprintf("%d", accountID)},
	}
	req := inertiaRequest(http.MethodPost, "/lots", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	// gonertia Render always writes 200; validation is via props.errors (422 semantics).
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 (inertia validation) body=%s", rr.Code, rr.Body.String())
	}
	assertInertiaComponent(t, rr, "Lots/New")
	assertInertiaErrors(t, rr, "item_qty")

	lots, err := s.ListLots()
	if err != nil {
		t.Fatal(err)
	}
	if len(lots) != 0 {
		t.Errorf("expected no lots created, got %d", len(lots))
	}
}

func TestLotsHandler_Show_Inertia(t *testing.T) {
	h, s := newLotsHandler(t)

	accountID, err := s.InsertCashAccount("PIX principal", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}
	lotID, err := s.CreateLotPurchase(store.CreateLotInput{
		Name:        "Lote show",
		PurchasedAt: "2026-07-21",
		ItemTitle:   "Caixa",
		ItemQty:     2,
		Costs: []store.CostInput{
			{Label: "Arremate", AmountCents: 1000, AlreadyPaid: false},
		},
		CashAccountID: accountID,
	})
	if err != nil {
		t.Fatal(err)
	}

	req := inertiaRequest(http.MethodGet, fmt.Sprintf("/lots/%d", lotID), nil)
	rr := httptest.NewRecorder()
	h.Show(rr, req, lotID)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
	assertInertiaComponent(t, rr, "Lots/Show")
	assertInertiaProp(t, rr, "lot")
	assertInertiaProp(t, rr, "items")
	assertInertiaProp(t, rr, "costs")
}
