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

func newReceivablesHandler(t *testing.T) (*ReceivablesHandler, *store.SQLiteStore) {
	t.Helper()
	s, err := store.NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	h := NewReceivablesHandler(setupTestRenderer(t), s, testSite(), cais.Config{}, setupTestInertia(t))
	return h, s
}

func seedOpenReceivable(t *testing.T, s *store.SQLiteStore) (accountID, receivableID int64) {
	t.Helper()
	accountID, err := s.InsertCashAccount("PIX principal", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}
	lotID, err := s.CreateLotPurchase(store.CreateLotInput{
		Name:        "Monitores ML",
		PurchasedAt: "2026-07-20",
		ItemTitle:   "Monitor LG",
		ItemQty:     1,
		Costs: []store.CostInput{
			{Label: "Arremate", AmountCents: 2741, AlreadyPaid: true},
		},
		CashAccountID: accountID,
		PaidAt:        "2026-07-20T12:00:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}
	items, err := s.ListItemsByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) == 0 {
		t.Fatal("no items")
	}
	_, err = s.CreateSale(store.CreateSaleInput{
		ItemID:        items[0].ID,
		SoldAt:        "2026-07-22T12:00:00Z",
		Channel:       "mercadolivre",
		GrossCents:    15000,
		FeeCents:      1500,
		PaymentStatus: "pending",
		DueOn:         "2026-08-01",
	})
	if err != nil {
		t.Fatal(err)
	}
	recs, err := s.ListReceivables()
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 {
		t.Fatalf("receivables=%d", len(recs))
	}
	return accountID, recs[0].ID
}

func TestReceivablesHandler_Index_OK(t *testing.T) {
	h, _ := newReceivablesHandler(t)

	req := inertiaRequest(http.MethodGet, "/receivables", nil)
	rr := httptest.NewRecorder()
	h.Index(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	assertInertiaComponent(t, rr, "Receivables/Index")
	assertInertiaProp(t, rr, "receivables")
	assertInertiaProp(t, rr, "cashAccounts")
}

func TestReceivablesHandler_Settle_Redirects(t *testing.T) {
	h, s := newReceivablesHandler(t)
	accountID, receivableID := seedOpenReceivable(t, s)

	form := url.Values{
		"cash_account_id": {fmt.Sprintf("%d", accountID)},
		"received_at":     {"2026-07-26"},
	}
	req := inertiaRequest(http.MethodPost, fmt.Sprintf("/receivables/%d/settle", receivableID), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.Settle(rr, req, receivableID)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303 body=%s", rr.Code, rr.Body.String())
	}
	if loc := rr.Header().Get("Location"); loc != "/receivables" {
		t.Errorf("Location = %q, want /receivables", loc)
	}

	recs, err := s.ListReceivables()
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 || recs[0].Status != "received" {
		t.Fatalf("receivable status=%v", recs)
	}
}
