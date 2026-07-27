package store

import (
	"testing"
)

func TestCreateLotPurchase_MonitorsAlreadyPaid(t *testing.T) {
	st := newTestStore(t)

	accountID, err := st.InsertCashAccount("PIX principal", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}

	input := CreateLotInput{
		Name:        "Monitores — leilão Jul/2026",
		PurchasedAt: "2026-07-20",
		ItemTitle:   "Monitor",
		ItemQty:     22,
		Costs: []CostInput{
			{Label: "Arremate", AmountCents: 60300, AlreadyPaid: true},
		},
		CashAccountID: accountID,
		PaidAt:        "2026-07-20T12:00:00Z",
	}

	lotID, err := st.CreateLotPurchase(input)
	if err != nil {
		t.Fatal(err)
	}
	if lotID == 0 {
		t.Fatal("lotID = 0")
	}

	items, err := st.ListItemsByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 22 {
		t.Fatalf("items=%d", len(items))
	}

	var sum int64
	var high, low int
	for _, it := range items {
		sum += it.UnitCostCents
		if it.Status != "in_stock" {
			t.Fatalf("status=%s", it.Status)
		}
		if it.Title != "Monitor" {
			t.Fatalf("title=%q", it.Title)
		}
		switch it.UnitCostCents {
		case 2741:
			high++
		case 2740:
			low++
		default:
			t.Fatalf("unexpected unit cost %d", it.UnitCostCents)
		}
	}
	if sum != 60300 {
		t.Fatalf("cost sum=%d", sum)
	}
	if high != 20 || low != 2 {
		t.Fatalf("high=%d low=%d want 20/2", high, low)
	}

	pays, err := st.ListPayablesByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	if len(pays) != 1 || pays[0].Status != "paid" || pays[0].AmountCents != 60300 {
		t.Fatalf("payable %+v", pays)
	}
	if pays[0].PaidAt == nil || *pays[0].PaidAt == "" {
		t.Fatalf("expected paid_at set, got %+v", pays[0])
	}

	bal, err := st.CashBalance(accountID)
	if err != nil || bal != -60300 {
		t.Fatalf("balance=%d err=%v", bal, err)
	}
}

func TestCreateLotPurchase_Unpaid(t *testing.T) {
	st := newTestStore(t)

	accountID, err := st.InsertCashAccount("PIX principal", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}

	input := CreateLotInput{
		Name:        "Lote frete a pagar",
		PurchasedAt: "2026-07-21",
		ItemTitle:   "Caixa",
		ItemQty:     2,
		Costs: []CostInput{
			{Label: "Arremate", AmountCents: 1000, AlreadyPaid: false},
		},
	}

	lotID, err := st.CreateLotPurchase(input)
	if err != nil {
		t.Fatal(err)
	}

	items, err := st.ListItemsByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("items=%d", len(items))
	}
	var sum int64
	for _, it := range items {
		sum += it.UnitCostCents
		if it.Status != "in_stock" {
			t.Fatalf("status=%s", it.Status)
		}
	}
	if sum != 1000 {
		t.Fatalf("cost sum=%d", sum)
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
	if pays[0].AmountCents != 1000 {
		t.Fatalf("amount=%d", pays[0].AmountCents)
	}
	if pays[0].PaidAt != nil {
		t.Fatalf("expected nil paid_at, got %v", *pays[0].PaidAt)
	}

	bal, err := st.CashBalance(accountID)
	if err != nil {
		t.Fatal(err)
	}
	if bal != 0 {
		t.Fatalf("balance=%d want 0 (no cash movement)", bal)
	}
}

func TestCreateLotPurchase_Validation(t *testing.T) {
	st := newTestStore(t)

	_, err := st.CreateLotPurchase(CreateLotInput{
		Name: "x", PurchasedAt: "2026-01-01", ItemTitle: "y", ItemQty: 0,
		Costs: []CostInput{{Label: "a", AmountCents: 100}},
	})
	if err == nil {
		t.Fatal("expected error for qty=0")
	}

	_, err = st.CreateLotPurchase(CreateLotInput{
		Name: "x", PurchasedAt: "2026-01-01", ItemTitle: "y", ItemQty: 1,
		Costs: nil,
	})
	if err == nil {
		t.Fatal("expected error for empty costs")
	}

	_, err = st.CreateLotPurchase(CreateLotInput{
		Name: "x", PurchasedAt: "2026-01-01", ItemTitle: "y", ItemQty: 1,
		Costs: []CostInput{{Label: "a", AmountCents: 0}},
	})
	if err == nil {
		t.Fatal("expected error for zero total")
	}

	_, err = st.CreateLotPurchase(CreateLotInput{
		Name: "x", PurchasedAt: "2026-01-01", ItemTitle: "y", ItemQty: 1,
		Costs: []CostInput{{Label: "a", AmountCents: 100, AlreadyPaid: true}},
		// CashAccountID missing
		PaidAt: "2026-01-01T00:00:00Z",
	})
	if err == nil {
		t.Fatal("expected error when paid cost lacks cash account")
	}
}
