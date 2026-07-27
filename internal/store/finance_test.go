package store

import (
	"testing"

	"github.com/puppe1990/leilao-erp/internal/domain"
)

func TestSettlePayable(t *testing.T) {
	st := newTestStore(t)

	accountID, err := st.InsertCashAccount("PIX principal", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}

	const amount int64 = 1000
	lotID, err := st.CreateLotPurchase(CreateLotInput{
		Name:        "Lote frete a pagar",
		PurchasedAt: "2026-07-21",
		ItemTitle:   "Caixa",
		ItemQty:     2,
		Costs: []CostInput{
			{Label: "Arremate", AmountCents: amount, AlreadyPaid: false},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	pays, err := st.ListPayablesByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	if len(pays) != 1 {
		t.Fatalf("payables=%d", len(pays))
	}
	if pays[0].Status != "open" {
		t.Fatalf("status=%s want open", pays[0].Status)
	}
	payableID := pays[0].ID

	balBefore, err := st.CashBalance(accountID)
	if err != nil {
		t.Fatal(err)
	}
	if balBefore != 0 {
		t.Fatalf("balance before=%d want 0", balBefore)
	}

	paidAt := "2026-07-25T14:00:00Z"
	if err := st.SettlePayable(payableID, accountID, paidAt); err != nil {
		t.Fatal(err)
	}

	pays, err = st.ListPayablesByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	if pays[0].Status != "paid" {
		t.Fatalf("status=%s want paid", pays[0].Status)
	}
	if pays[0].PaidAt == nil || *pays[0].PaidAt != paidAt {
		t.Fatalf("paid_at=%v want %s", pays[0].PaidAt, paidAt)
	}

	bal, err := st.CashBalance(accountID)
	if err != nil {
		t.Fatal(err)
	}
	if bal != -amount {
		t.Fatalf("balance=%d want %d (cash out)", bal, -amount)
	}

	if err := st.SettlePayable(payableID, accountID, paidAt); err == nil {
		t.Fatal("expected error on second settle")
	}
}

func TestSettleReceivable(t *testing.T) {
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
	if len(items) < 1 {
		t.Fatal("expected items")
	}
	item := items[0]

	const gross, fee, shipping int64 = 18000, 3000, 2000
	wantNet := domain.SaleNet(gross, fee, shipping) // 13000

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

	recs, err := st.ListReceivables()
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 {
		t.Fatalf("receivables=%d", len(recs))
	}
	if recs[0].Status != "open" {
		t.Fatalf("receivable status=%s want open", recs[0].Status)
	}
	recID := recs[0].ID

	balBefore, err := st.CashBalance(accountID)
	if err != nil {
		t.Fatal(err)
	}

	receivedAt := "2026-08-06T15:00:00Z"
	if err := st.SettleReceivable(recID, accountID, receivedAt); err != nil {
		t.Fatal(err)
	}

	recs, err = st.ListReceivables()
	if err != nil {
		t.Fatal(err)
	}
	if recs[0].Status != "received" {
		t.Fatalf("status=%s want received", recs[0].Status)
	}
	if recs[0].ReceivedAt == nil || *recs[0].ReceivedAt != receivedAt {
		t.Fatalf("received_at=%v want %s", recs[0].ReceivedAt, receivedAt)
	}

	balAfter, err := st.CashBalance(accountID)
	if err != nil {
		t.Fatal(err)
	}
	if balAfter != balBefore+wantNet {
		t.Fatalf("balance=%d want %d (cash in net)", balAfter, balBefore+wantNet)
	}

	sales, err := st.ListSales()
	if err != nil {
		t.Fatal(err)
	}
	if len(sales) != 1 {
		t.Fatalf("sales=%d", len(sales))
	}
	if sales[0].ID != saleID {
		t.Fatalf("sale id=%d want %d", sales[0].ID, saleID)
	}
	if sales[0].PaymentStatus != "received" {
		t.Fatalf("payment_status=%s want received", sales[0].PaymentStatus)
	}

	if err := st.SettleReceivable(recID, accountID, receivedAt); err == nil {
		t.Fatal("expected error on second settle")
	}
}
