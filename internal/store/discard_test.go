package store

import (
	"errors"
	"testing"
)

// seedDiscardLot creates one lot with qty units of a monitor title.
func seedDiscardLot(t *testing.T, st *SQLiteStore, title string, qty int) int64 {
	t.Helper()
	accID, err := st.InsertCashAccount("Caixa", "cash", 0)
	if err != nil {
		t.Fatal(err)
	}
	lotID, err := st.CreateLotPurchase(CreateLotInput{
		Name:          "Lote baixa",
		PurchasedAt:   "2026-01-01",
		ItemTitle:     title,
		ItemQty:       qty,
		CashAccountID: accID,
		PaidAt:        "2026-01-01T12:00:00Z",
		Costs:         []CostInput{{Label: "Arremate", AmountCents: 30000, AlreadyPaid: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return lotID
}

func TestDiscardItem_LeavesStockButKeepsCost(t *testing.T) {
	st := newTestStore(t)
	title := "Monitor Dell P2016t 19,5\""
	lotID := seedDiscardLot(t, st, title, 3)

	items, err := st.ListItemsByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("want 3 items, got %d", len(items))
	}
	victim := items[0]
	costBefore := victim.UnitCostCents

	if err := st.DiscardItem(victim.ID); err != nil {
		t.Fatal(err)
	}

	stock, err := st.ListItemsInStock()
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range stock {
		if it.ID == victim.ID {
			t.Fatalf("discarded item %d still listed in stock", victim.ID)
		}
	}
	if len(stock) != 2 {
		t.Fatalf("want 2 remaining in stock, got %d", len(stock))
	}

	// Row survives with its cost — history intact.
	kept, err := st.FindItemByID(victim.ID)
	if err != nil {
		t.Fatalf("discarded item should still exist: %v", err)
	}
	if kept.UnitCostCents != costBefore {
		t.Fatalf("unit cost changed: want %d, got %d", costBefore, kept.UnitCostCents)
	}
}

func TestDiscardItem_HidesFromStockGroups(t *testing.T) {
	st := newTestStore(t)
	title := "Monitor Dell P2016t 19,5\""
	lotID := seedDiscardLot(t, st, title, 1)

	items, err := st.ListItemsByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.DiscardItem(items[0].ID); err != nil {
		t.Fatal(err)
	}

	groups, err := st.ListStockProductGroups()
	if err != nil {
		t.Fatal(err)
	}
	for _, g := range groups {
		if g.Name == title {
			t.Fatalf("discarded product %q should not appear in stock groups", title)
		}
	}
}

func TestDiscardItem_RejectsSoldUnit(t *testing.T) {
	st := newTestStore(t)
	title := "Monitor Dell P1913Sb 19\""
	lotID := seedDiscardLot(t, st, title, 1)
	accID, err := st.InsertCashAccount("Caixa 2", "cash", 0)
	if err != nil {
		t.Fatal(err)
	}
	items, err := st.ListItemsByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.CreateSale(CreateSaleInput{
		ItemID:        items[0].ID,
		SoldAt:        "2026-01-05T10:00:00Z",
		Channel:       "direct",
		GrossCents:    20000,
		PaymentStatus: "received",
		CashAccountID: accID,
	}); err != nil {
		t.Fatal(err)
	}

	if err := st.DiscardItem(items[0].ID); !errors.Is(err, ErrCannotUpdate) {
		t.Fatalf("want ErrCannotUpdate for sold unit, got %v", err)
	}
}

func TestDiscardItem_RejectsDoubleDiscard(t *testing.T) {
	st := newTestStore(t)
	lotID := seedDiscardLot(t, st, "Monitor LG W1943C 19\"", 1)
	items, err := st.ListItemsByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.DiscardItem(items[0].ID); err != nil {
		t.Fatal(err)
	}
	if err := st.DiscardItem(items[0].ID); !errors.Is(err, ErrCannotUpdate) {
		t.Fatalf("want ErrCannotUpdate on second discard, got %v", err)
	}
}

func TestUndiscardItem_RestoresToStock(t *testing.T) {
	st := newTestStore(t)
	title := "Monitor Samsung 743B 17\" (com base)"
	lotID := seedDiscardLot(t, st, title, 1)
	items, err := st.ListItemsByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.DiscardItem(items[0].ID); err != nil {
		t.Fatal(err)
	}
	if err := st.UndiscardItem(items[0].ID); err != nil {
		t.Fatal(err)
	}

	stock, err := st.ListItemsInStock()
	if err != nil {
		t.Fatal(err)
	}
	if len(stock) != 1 || stock[0].ID != items[0].ID {
		t.Fatalf("want item %d back in stock, got %+v", items[0].ID, stock)
	}
}
