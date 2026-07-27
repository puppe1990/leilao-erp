package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"

	"github.com/puppe1990/leilao-erp/internal/store"
)

func TestDashboardHandler_InertiaComponent(t *testing.T) {
	h := NewDashboardHandler(setupTestRenderer(t), setupTestStore(t), testSite(), cais.Config{}, setupTestInertia(t))

	req := inertiaRequest(http.MethodGet, "/dashboard", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	assertInertiaComponent(t, rr, "Dashboard")
}

func TestDashboardHandler_FinancePropsEmpty(t *testing.T) {
	h := NewDashboardHandler(setupTestRenderer(t), setupTestStore(t), testSite(), cais.Config{Env: "test"}, setupTestInertia(t))

	req := inertiaRequest(http.MethodGet, "/dashboard", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}

	assertInertiaProp(t, rr, "balances")
	assertInertiaProp(t, rr, "totalCashFormatted")
	assertInertiaProp(t, rr, "openPayablesFormatted")
	assertInertiaProp(t, rr, "openReceivablesFormatted")
	assertInertiaProp(t, rr, "monthProfitFormatted")
	assertInertiaProp(t, rr, "overduePayables")
	assertInertiaProp(t, rr, "overdueReceivables")
	assertInertiaProp(t, rr, "lotCount")

	ctaLot := assertInertiaProp(t, rr, "ctaLot")
	if ctaLot != true {
		t.Errorf("ctaLot = %v, want true when no lots", ctaLot)
	}

	totalCash := assertInertiaProp(t, rr, "totalCashFormatted")
	if totalCash != "R$ 0,00" {
		t.Errorf("totalCashFormatted = %v, want R$ 0,00", totalCash)
	}
}

func TestDashboardHandler_FinancePropsWithData(t *testing.T) {
	s := setupTestStore(t).(*store.SQLiteStore)

	accountID, err := s.InsertCashAccount("PIX principal", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.CreateLotPurchase(store.CreateLotInput{
		Name:        "Lote teste",
		PurchasedAt: "2026-07-20",
		ItemTitle:   "Item",
		ItemQty:     2,
		Costs: []store.CostInput{
			{Label: "Arremate", AmountCents: 10000, AlreadyPaid: true},
		},
		CashAccountID: accountID,
		PaidAt:        "2026-07-20T12:00:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}

	h := NewDashboardHandler(setupTestRenderer(t), s, testSite(), cais.Config{}, setupTestInertia(t))

	req := inertiaRequest(http.MethodGet, "/dashboard", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}

	ctaLot := assertInertiaProp(t, rr, "ctaLot")
	if ctaLot != false {
		t.Errorf("ctaLot = %v, want false when lots exist", ctaLot)
	}

	lotCount := assertInertiaProp(t, rr, "lotCount")
	// JSON numbers decode as float64
	if n, ok := lotCount.(float64); !ok || n != 1 {
		t.Errorf("lotCount = %v, want 1", lotCount)
	}

	balances := assertInertiaProp(t, rr, "balances")
	balSlice, ok := balances.([]any)
	if !ok || len(balSlice) != 1 {
		t.Fatalf("balances = %v, want 1 account", balances)
	}
	bal0, ok := balSlice[0].(map[string]any)
	if !ok {
		t.Fatalf("balance[0] type %T", balSlice[0])
	}
	if bal0["name"] != "PIX principal" {
		t.Errorf("balance name = %v", bal0["name"])
	}
	if bal0["formatted"] != "-R$ 100,00" {
		t.Errorf("balance formatted = %v, want -R$ 100,00", bal0["formatted"])
	}

	totalCash := assertInertiaProp(t, rr, "totalCashFormatted")
	if totalCash != "-R$ 100,00" {
		t.Errorf("totalCashFormatted = %v, want -R$ 100,00", totalCash)
	}
}
