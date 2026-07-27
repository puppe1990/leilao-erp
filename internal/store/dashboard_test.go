package store

import (
	"testing"
	"time"

	"github.com/puppe1990/leilao-erp/internal/domain"
)

func TestDashboardSummary(t *testing.T) {
	st := newTestStore(t)

	accountID, err := st.InsertCashAccount("PIX principal", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}

	const purchaseCost int64 = 2741
	lotID, err := st.CreateLotPurchase(CreateLotInput{
		Name:        "Monitores — leilão Jul/2026",
		PurchasedAt: "2026-07-20",
		ItemTitle:   "Monitor",
		ItemQty:     2,
		Costs: []CostInput{
			{Label: "Arremate", AmountCents: purchaseCost * 2, AlreadyPaid: true},
		},
		CashAccountID: accountID,
		PaidAt:        "2026-07-20T12:00:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Unpaid cost → open (and overdue) payable
	const unpaidFreight int64 = 500
	if err := st.AddPurchaseCost(lotID, CostInput{
		Label:       "Frete",
		AmountCents: unpaidFreight,
		AlreadyPaid: false,
	}, 0, ""); err != nil {
		t.Fatal(err)
	}
	// Force payable due_on into the past so it counts as overdue
	if _, err := st.db.Exec(
		`UPDATE payables SET due_on = '2020-01-01' WHERE lot_id = ? AND status = 'open'`,
		lotID,
	); err != nil {
		t.Fatal(err)
	}

	items, err := st.ListItemsByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("items=%d", len(items))
	}

	now := time.Now().UTC()
	soldAt := now.Format(time.RFC3339)
	const gross int64 = 15000
	unitCost0 := items[0].UnitCostCents

	_, err = st.CreateSale(CreateSaleInput{
		ItemID:        items[0].ID,
		SoldAt:        soldAt,
		Channel:       "direct",
		GrossCents:    gross,
		PaymentStatus: "received",
		CashAccountID: accountID,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Pending marketplace sale → open receivable (overdue due_on)
	const gross2, fee, shipping int64 = 18000, 3000, 2000
	wantNet2 := domain.SaleNet(gross2, fee, shipping)
	unitCost1 := items[1].UnitCostCents
	_, err = st.CreateSale(CreateSaleInput{
		ItemID:        items[1].ID,
		SoldAt:        soldAt,
		Channel:       "mercadolivre",
		GrossCents:    gross2,
		FeeCents:      fee,
		ShippingCents: shipping,
		PaymentStatus: "pending",
		DueOn:         "2020-01-15",
	})
	if err != nil {
		t.Fatal(err)
	}

	sum, err := st.DashboardSummary()
	if err != nil {
		t.Fatal(err)
	}

	// cash: -purchase*2 + sale net (15000)
	wantCash := -(purchaseCost * 2) + gross
	if sum.TotalCashCents != wantCash {
		t.Fatalf("TotalCashCents=%d want %d", sum.TotalCashCents, wantCash)
	}
	if len(sum.CashBalances) != 1 {
		t.Fatalf("CashBalances len=%d want 1", len(sum.CashBalances))
	}
	if sum.CashBalances[0].ID != accountID {
		t.Fatalf("CashBalances[0].ID=%d want %d", sum.CashBalances[0].ID, accountID)
	}
	if sum.CashBalances[0].Name != "PIX principal" {
		t.Fatalf("CashBalances[0].Name=%q", sum.CashBalances[0].Name)
	}
	if sum.CashBalances[0].Cents != wantCash {
		t.Fatalf("CashBalances[0].Cents=%d want %d", sum.CashBalances[0].Cents, wantCash)
	}

	if sum.OpenPayablesCents != unpaidFreight {
		t.Fatalf("OpenPayablesCents=%d want %d", sum.OpenPayablesCents, unpaidFreight)
	}
	if sum.OpenReceivablesCents != wantNet2 {
		t.Fatalf("OpenReceivablesCents=%d want %d", sum.OpenReceivablesCents, wantNet2)
	}

	wantProfit := domain.Margin(gross, unitCost0) + domain.Margin(wantNet2, unitCost1)
	if sum.MonthProfitCents != wantProfit {
		t.Fatalf("MonthProfitCents=%d want %d", sum.MonthProfitCents, wantProfit)
	}

	if sum.OverduePayables != 1 {
		t.Fatalf("OverduePayables=%d want 1", sum.OverduePayables)
	}
	if sum.OverdueReceivables != 1 {
		t.Fatalf("OverdueReceivables=%d want 1", sum.OverdueReceivables)
	}
	if sum.LotCount != 1 {
		t.Fatalf("LotCount=%d want 1", sum.LotCount)
	}
}

func TestCancelPendingSale(t *testing.T) {
	st := newTestStore(t)

	accountID, err := st.InsertCashAccount("PIX principal", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}

	const purchaseCost int64 = 2741
	lotID, err := st.CreateLotPurchase(CreateLotInput{
		Name:        "Monitores ML",
		PurchasedAt: "2026-07-20",
		ItemTitle:   "Monitor LG",
		ItemQty:     2,
		Costs: []CostInput{
			{Label: "Arremate", AmountCents: purchaseCost * 2, AlreadyPaid: true},
		},
		CashAccountID: accountID,
		PaidAt:        "2026-07-20T12:00:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}

	items, err := st.ListItemsByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	item := items[0]

	const gross, fee, shipping int64 = 18000, 3000, 2000
	saleID, err := st.CreateSale(CreateSaleInput{
		ItemID:        item.ID,
		SoldAt:        "2026-07-23T10:00:00Z",
		Channel:       "mercadolivre",
		GrossCents:    gross,
		FeeCents:      fee,
		ShippingCents: shipping,
		PaymentStatus: "pending",
		DueOn:         "2026-08-06",
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := st.CancelPendingSale(saleID); err != nil {
		t.Fatal(err)
	}

	sales, err := st.ListSales()
	if err != nil {
		t.Fatal(err)
	}
	if len(sales) != 1 {
		t.Fatalf("sales=%d", len(sales))
	}
	if sales[0].PaymentStatus != "cancelled" {
		t.Fatalf("payment_status=%s want cancelled", sales[0].PaymentStatus)
	}

	recs, err := st.ListReceivables()
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 {
		t.Fatalf("receivables=%d", len(recs))
	}
	if recs[0].Status != "cancelled" {
		t.Fatalf("receivable status=%s want cancelled", recs[0].Status)
	}

	items, err = st.ListItemsByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range items {
		if it.Status != "in_stock" {
			t.Fatalf("item %d status=%s want in_stock", it.ID, it.Status)
		}
	}

	lot, err := st.FindLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	if lot.Status != "open" {
		t.Fatalf("lot status=%s want open", lot.Status)
	}
}

func TestCancelReceivedSale_NotAllowed(t *testing.T) {
	st := newTestStore(t)

	accountID, err := st.InsertCashAccount("PIX principal", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}

	lotID, err := st.CreateLotPurchase(CreateLotInput{
		Name:        "Lote único",
		PurchasedAt: "2026-07-20",
		ItemTitle:   "Item",
		ItemQty:     1,
		Costs: []CostInput{
			{Label: "Arremate", AmountCents: 1000, AlreadyPaid: true},
		},
		CashAccountID: accountID,
		PaidAt:        "2026-07-20T12:00:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}

	items, err := st.ListItemsByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}

	saleID, err := st.CreateSale(CreateSaleInput{
		ItemID:        items[0].ID,
		SoldAt:        "2026-07-22T15:00:00Z",
		Channel:       "direct",
		GrossCents:    5000,
		PaymentStatus: "received",
		CashAccountID: accountID,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = st.CancelPendingSale(saleID)
	if err == nil {
		t.Fatal("expected error cancelling received sale")
	}
}
