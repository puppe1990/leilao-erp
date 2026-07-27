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

func newSalesHandler(t *testing.T) (*SalesHandler, *store.SQLiteStore) {
	t.Helper()
	s, err := store.NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	h := NewSalesHandler(setupTestRenderer(t), s, testSite(), cais.Config{}, setupTestInertia(t))
	return h, s
}

func seedLotWithItems(t *testing.T, s *store.SQLiteStore, qty int) (accountID int64, itemIDs []int64) {
	t.Helper()
	accountID, err := s.InsertCashAccount("PIX principal", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}
	lotID, err := s.CreateLotPurchase(store.CreateLotInput{
		Name:        "Monitores — teste",
		PurchasedAt: "2026-07-20",
		ItemTitle:   "Monitor",
		ItemQty:     qty,
		Costs: []store.CostInput{
			{Label: "Arremate", AmountCents: 60300, AlreadyPaid: true},
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
	itemIDs = make([]int64, len(items))
	for i, it := range items {
		itemIDs[i] = it.ID
	}
	return accountID, itemIDs
}

func TestSalesHandler_Index_Inertia(t *testing.T) {
	h, _ := newSalesHandler(t)

	req := inertiaRequest(http.MethodGet, "/sales", nil)
	rr := httptest.NewRecorder()
	h.Index(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	assertInertiaComponent(t, rr, "Sales/Index")
	assertInertiaProp(t, rr, "sales")
}

func TestSalesHandler_Create_Direct_Redirects(t *testing.T) {
	h, s := newSalesHandler(t)
	accountID, itemIDs := seedLotWithItems(t, s, 2)

	form := url.Values{
		"item_id":         {fmt.Sprintf("%d", itemIDs[0])},
		"channel":         {"direct"},
		"gross":           {"150,00"},
		"fee":             {"0"},
		"shipping":        {"0"},
		"payment_status":  {"received"},
		"cash_account_id": {fmt.Sprintf("%d", accountID)},
		"sold_at":         {"2026-07-22"},
	}
	req := inertiaRequest(http.MethodPost, "/sales", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303 body=%s", rr.Code, rr.Body.String())
	}
	if loc := rr.Header().Get("Location"); loc != "/sales" {
		t.Errorf("Location = %q, want /sales", loc)
	}

	sales, err := s.ListSales()
	if err != nil {
		t.Fatal(err)
	}
	if len(sales) != 1 {
		t.Fatalf("sales = %d, want 1", len(sales))
	}
	if sales[0].Channel != "direct" {
		t.Errorf("channel = %s, want direct", sales[0].Channel)
	}
	if sales[0].PaymentStatus != "received" {
		t.Errorf("payment_status = %s, want received", sales[0].PaymentStatus)
	}
	if sales[0].GrossCents != 15000 {
		t.Errorf("gross = %d, want 15000", sales[0].GrossCents)
	}

	items, err := s.ListItemsInStock()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Errorf("in_stock items = %d, want 1", len(items))
	}
}

func TestSalesHandler_Create_SoldItem_Validation(t *testing.T) {
	h, s := newSalesHandler(t)
	accountID, itemIDs := seedLotWithItems(t, s, 1)

	_, err := s.CreateSale(store.CreateSaleInput{
		ItemID:        itemIDs[0],
		SoldAt:        "2026-07-22T12:00:00Z",
		Channel:       "direct",
		GrossCents:    10000,
		PaymentStatus: "received",
		CashAccountID: accountID,
	})
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{
		"item_id":         {fmt.Sprintf("%d", itemIDs[0])},
		"channel":         {"direct"},
		"gross":           {"150,00"},
		"payment_status":  {"received"},
		"cash_account_id": {fmt.Sprintf("%d", accountID)},
	}
	req := inertiaRequest(http.MethodPost, "/sales", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 (inertia validation) body=%s", rr.Code, rr.Body.String())
	}
	assertInertiaComponent(t, rr, "Sales/New")
	assertInertiaErrors(t, rr, "item_id")

	sales, err := s.ListSales()
	if err != nil {
		t.Fatal(err)
	}
	if len(sales) != 1 {
		t.Errorf("expected only original sale, got %d", len(sales))
	}
}

func TestSalesHandler_Cancel_Pending_Redirects(t *testing.T) {
	h, s := newSalesHandler(t)
	_, itemIDs := seedLotWithItems(t, s, 1)

	saleID, err := s.CreateSale(store.CreateSaleInput{
		ItemID:        itemIDs[0],
		SoldAt:        "2026-07-22T12:00:00Z",
		Channel:       "mercadolivre",
		GrossCents:    20000,
		FeeCents:      2000,
		ShippingCents: 0,
		PaymentStatus: "pending",
		DueOn:         "2026-08-01",
	})
	if err != nil {
		t.Fatal(err)
	}

	req := inertiaRequest(http.MethodPost, fmt.Sprintf("/sales/%d/cancel", saleID), nil)
	rr := httptest.NewRecorder()
	h.Cancel(rr, req, saleID)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303 body=%s", rr.Code, rr.Body.String())
	}
	if loc := rr.Header().Get("Location"); loc != "/sales" {
		t.Errorf("Location = %q, want /sales", loc)
	}

	sales, err := s.ListSales()
	if err != nil {
		t.Fatal(err)
	}
	if len(sales) != 1 {
		t.Fatalf("sales = %d, want 1", len(sales))
	}
	if sales[0].PaymentStatus != "cancelled" {
		t.Errorf("payment_status = %s, want cancelled", sales[0].PaymentStatus)
	}
}
