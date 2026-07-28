package store

import (
	"testing"
)

func TestListStockProductGroups_AglutinatesSameTitle(t *testing.T) {
	st := newTestStore(t)

	accID, err := st.InsertCashAccount("Caixa", "cash", 0)
	if err != nil {
		t.Fatal(err)
	}

	// Two lots, same monitor title → one product group with qty 5
	_, err = st.CreateLotPurchase(CreateLotInput{
		Name:          "Lote A",
		PurchasedAt:   "2026-01-01",
		ItemTitle:     "Monitor Dell P1914Sf 19\" (sem base)",
		ItemQty:       3,
		CashAccountID: accID,
		PaidAt:        "2026-01-01T12:00:00Z",
		Costs:         []CostInput{{Label: "Arremate", AmountCents: 30000, AlreadyPaid: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = st.CreateLotPurchase(CreateLotInput{
		Name:          "Lote B",
		PurchasedAt:   "2026-01-02",
		ItemTitle:     "Monitor Dell P1914Sf 19\" (sem base)",
		ItemQty:       2,
		CashAccountID: accID,
		PaidAt:        "2026-01-02T12:00:00Z",
		Costs:         []CostInput{{Label: "Arremate", AmountCents: 20000, AlreadyPaid: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	// Different product
	_, err = st.CreateLotPurchase(CreateLotInput{
		Name:          "Lote C",
		PurchasedAt:   "2026-01-03",
		ItemTitle:     "Monitor Samsung 733NW 17\" (sem base)",
		ItemQty:       1,
		CashAccountID: accID,
		PaidAt:        "2026-01-03T12:00:00Z",
		Costs:         []CostInput{{Label: "Arremate", AmountCents: 5000, AlreadyPaid: true}},
	})
	if err != nil {
		t.Fatal(err)
	}

	groups, err := st.ListStockProductGroups()
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 2 {
		t.Fatalf("want 2 product groups, got %d: %+v", len(groups), groups)
	}

	byName := map[string]int{}
	for _, g := range groups {
		byName[g.Name] = g.QtyInStock
		if g.ID == 0 {
			t.Errorf("product %q missing catalog id", g.Name)
		}
		if g.SampleItemID == 0 {
			t.Errorf("product %q missing sample item", g.Name)
		}
	}
	if byName["Monitor Dell P1914Sf 19\" (sem base)"] != 5 {
		t.Errorf("Dell qty want 5 got %d", byName["Monitor Dell P1914Sf 19\" (sem base)"])
	}
	if byName["Monitor Samsung 733NW 17\" (sem base)"] != 1 {
		t.Errorf("Samsung qty want 1 got %d", byName["Monitor Samsung 733NW 17\" (sem base)"])
	}

	products, err := st.ListProducts()
	if err != nil {
		t.Fatal(err)
	}
	if len(products) < 2 {
		t.Fatalf("catalog should have ≥2 products, got %d", len(products))
	}
}

func TestUpdateProductSaleHint_SyncsUnits(t *testing.T) {
	st := newTestStore(t)
	accID, err := st.InsertCashAccount("Caixa", "cash", 0)
	if err != nil {
		t.Fatal(err)
	}
	_, err = st.CreateLotPurchase(CreateLotInput{
		Name:          "Lote",
		PurchasedAt:   "2026-01-01",
		ItemTitle:     "Monitor X",
		ItemQty:       2,
		CashAccountID: accID,
		PaidAt:        "2026-01-01T12:00:00Z",
		Costs:         []CostInput{{Label: "Arremate", AmountCents: 10000, AlreadyPaid: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	groups, err := st.ListStockProductGroups()
	if err != nil || len(groups) != 1 {
		t.Fatalf("groups: %v %v", groups, err)
	}
	hint := int64(19900)
	if err := st.UpdateProductSaleHint(groups[0].ID, &hint); err != nil {
		t.Fatal(err)
	}
	units, err := st.ListItemsInStock()
	if err != nil {
		t.Fatal(err)
	}
	for _, u := range units {
		if u.SalePriceHintCents == nil || *u.SalePriceHintCents != 19900 {
			t.Fatalf("unit %d sale hint want 19900 got %v", u.ID, u.SalePriceHintCents)
		}
	}
}
