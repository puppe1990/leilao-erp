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

func newPayablesHandler(t *testing.T) (*PayablesHandler, *store.SQLiteStore) {
	t.Helper()
	s, err := store.NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	h := NewPayablesHandler(setupTestRenderer(t), s, testSite(), cais.Config{}, setupTestInertia(t))
	return h, s
}

func seedOpenPayable(t *testing.T, s *store.SQLiteStore) (accountID, payableID int64) {
	t.Helper()
	accountID, err := s.InsertCashAccount("PIX principal", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}
	lotID, err := s.CreateLotPurchase(store.CreateLotInput{
		Name:        "Lote frete a pagar",
		PurchasedAt: "2026-07-21",
		ItemTitle:   "Caixa",
		ItemQty:     2,
		Costs: []store.CostInput{
			{Label: "Arremate", AmountCents: 1000, AlreadyPaid: false},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	pays, err := s.ListPayablesByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	if len(pays) != 1 {
		t.Fatalf("payables=%d", len(pays))
	}
	return accountID, pays[0].ID
}

func TestPayablesHandler_Index_OK(t *testing.T) {
	h, _ := newPayablesHandler(t)

	req := inertiaRequest(http.MethodGet, "/payables", nil)
	rr := httptest.NewRecorder()
	h.Index(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	assertInertiaComponent(t, rr, "Payables/Index")
	assertInertiaProp(t, rr, "payables")
	assertInertiaProp(t, rr, "cashAccounts")
}

func TestPayablesHandler_Settle_Redirects(t *testing.T) {
	h, s := newPayablesHandler(t)
	accountID, payableID := seedOpenPayable(t, s)

	form := url.Values{
		"cash_account_id": {fmt.Sprintf("%d", accountID)},
		"paid_at":         {"2026-07-25"},
	}
	req := inertiaRequest(http.MethodPost, fmt.Sprintf("/payables/%d/settle", payableID), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.Settle(rr, req, payableID)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303 body=%s", rr.Code, rr.Body.String())
	}
	if loc := rr.Header().Get("Location"); loc != "/payables" {
		t.Errorf("Location = %q, want /payables", loc)
	}

	pays, err := s.ListPayables()
	if err != nil {
		t.Fatal(err)
	}
	if len(pays) != 1 || pays[0].Status != "paid" {
		t.Fatalf("payable status=%v", pays)
	}
}
