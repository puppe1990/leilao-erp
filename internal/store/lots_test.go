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

func TestAddPurchaseCost_RecalcInStockOnly(t *testing.T) {
	st := newTestStore(t)

	accountID, err := st.InsertCashAccount("PIX principal", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}

	// Create lot with 2 items total cost 1000 cents → 500 each
	lotID, err := st.CreateLotPurchase(CreateLotInput{
		Name:        "Lote frete extra",
		PurchasedAt: "2026-07-20",
		ItemTitle:   "Item",
		ItemQty:     2,
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
	if len(items) != 2 {
		t.Fatalf("items=%d", len(items))
	}
	for _, it := range items {
		if it.UnitCostCents != 500 {
			t.Fatalf("initial unit_cost=%d want 500", it.UnitCostCents)
		}
	}

	// Force item[0] status to sold
	if _, err := st.db.Exec(`UPDATE items SET status='sold' WHERE id=?`, items[0].ID); err != nil {
		t.Fatal(err)
	}

	// AddPurchaseCost +200 cents AlreadyPaid on same cash account
	err = st.AddPurchaseCost(lotID, CostInput{
		Label:       "Frete",
		AmountCents: 200,
		AlreadyPaid: true,
	}, accountID, "2026-07-21T10:00:00Z")
	if err != nil {
		t.Fatal(err)
	}

	items, err = st.ListItemsByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("items=%d", len(items))
	}

	// items ordered by id; [0] was sold
	if items[0].Status != "sold" {
		t.Fatalf("item[0] status=%s want sold", items[0].Status)
	}
	if items[0].UnitCostCents != 500 {
		t.Fatalf("sold item unit_cost=%d want 500 (frozen)", items[0].UnitCostCents)
	}
	if items[1].Status != "in_stock" {
		t.Fatalf("item[1] status=%s want in_stock", items[1].Status)
	}
	// remaining for 1 item = (1000+200) - 500 = 700
	if items[1].UnitCostCents != 700 {
		t.Fatalf("in_stock unit_cost=%d want 700", items[1].UnitCostCents)
	}

	// CashBalance decreases by 200 more (was -1000)
	bal, err := st.CashBalance(accountID)
	if err != nil {
		t.Fatal(err)
	}
	if bal != -1200 {
		t.Fatalf("balance=%d want -1200", bal)
	}

	// New payable paid exists
	pays, err := st.ListPayablesByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	if len(pays) != 2 {
		t.Fatalf("payables=%d want 2", len(pays))
	}
	var foundNew bool
	for _, p := range pays {
		if p.AmountCents == 200 {
			foundNew = true
			if p.Status != "paid" {
				t.Fatalf("new payable status=%s want paid", p.Status)
			}
			if p.PaidAt == nil || *p.PaidAt == "" {
				t.Fatalf("expected paid_at set on new payable")
			}
		}
	}
	if !foundNew {
		t.Fatal("expected payable with amount 200")
	}
}

func TestAddPurchaseCost_Unpaid(t *testing.T) {
	st := newTestStore(t)

	accountID, err := st.InsertCashAccount("PIX principal", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}

	lotID, err := st.CreateLotPurchase(CreateLotInput{
		Name:        "Lote frete a pagar",
		PurchasedAt: "2026-07-21",
		ItemTitle:   "Caixa",
		ItemQty:     2,
		Costs: []CostInput{
			{Label: "Arremate", AmountCents: 1000, AlreadyPaid: false},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	err = st.AddPurchaseCost(lotID, CostInput{
		Label:       "Frete",
		AmountCents: 200,
		AlreadyPaid: false,
	}, 0, "")
	if err != nil {
		t.Fatal(err)
	}

	items, err := st.ListItemsByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	// both still in_stock; remaining = 1200 across 2 → 600 each
	var sum int64
	for _, it := range items {
		if it.Status != "in_stock" {
			t.Fatalf("status=%s", it.Status)
		}
		sum += it.UnitCostCents
	}
	if sum != 1200 {
		t.Fatalf("cost sum=%d want 1200", sum)
	}
	for _, it := range items {
		if it.UnitCostCents != 600 {
			t.Fatalf("unit_cost=%d want 600", it.UnitCostCents)
		}
	}

	pays, err := st.ListPayablesByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	if len(pays) != 2 {
		t.Fatalf("payables=%d want 2", len(pays))
	}
	for _, p := range pays {
		if p.Status != "open" {
			t.Fatalf("status=%s want open", p.Status)
		}
		if p.PaidAt != nil {
			t.Fatalf("expected nil paid_at, got %v", *p.PaidAt)
		}
	}

	bal, err := st.CashBalance(accountID)
	if err != nil {
		t.Fatal(err)
	}
	if bal != 0 {
		t.Fatalf("balance=%d want 0 (no cash movement)", bal)
	}
}
