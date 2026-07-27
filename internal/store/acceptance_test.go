package store_test

import (
	"path/filepath"
	"testing"

	"github.com/puppe1990/leilao-erp/internal/db"
	"github.com/puppe1990/leilao-erp/internal/store"
)

func TestAcceptance_SeedSellSettle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "accept.db")
	st, err := store.NewSQLiteStore(path, "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })

	if err := db.RunSeeds(st); err != nil {
		t.Fatal(err)
	}
	if err := db.RunSeeds(st); err != nil {
		t.Fatal(err)
	}

	sum, err := st.DashboardSummary()
	if err != nil {
		t.Fatal(err)
	}
	if sum.LotCount != 1 {
		t.Fatalf("lots=%d", sum.LotCount)
	}
	if sum.TotalCashCents != -60300 {
		t.Fatalf("cash after seed=%d", sum.TotalCashCents)
	}

	items, err := st.ListItemsInStock()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 22 {
		t.Fatalf("stock=%d", len(items))
	}
	accs, err := st.ListCashAccounts()
	if err != nil || len(accs) == 0 {
		t.Fatalf("accounts %v %v", accs, err)
	}

	// Direct sale PIX
	_, err = st.CreateSale(store.CreateSaleInput{
		ItemID: items[0].ID, SoldAt: "2026-07-27T12:00:00Z", Channel: "direct",
		GrossCents: 15000, PaymentStatus: "received", CashAccountID: accs[0].ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Marketplace pending then settle
	items, _ = st.ListItemsInStock()
	_, err = st.CreateSale(store.CreateSaleInput{
		ItemID: items[0].ID, SoldAt: "2026-07-27T13:00:00Z", Channel: "mercadolivre",
		GrossCents: 18000, FeeCents: 3000, ShippingCents: 2000,
		PaymentStatus: "pending", DueOn: "2026-08-10",
	})
	if err != nil {
		t.Fatal(err)
	}
	recs, err := st.ListReceivables()
	if err != nil {
		t.Fatal(err)
	}
	var openID int64
	for _, r := range recs {
		if r.Status == "open" {
			openID = r.ID
			break
		}
	}
	if openID == 0 {
		t.Fatal("no open receivable")
	}
	if err := st.SettleReceivable(openID, accs[0].ID, "2026-07-27T14:00:00Z"); err != nil {
		t.Fatal(err)
	}

	sum, err = st.DashboardSummary()
	if err != nil {
		t.Fatal(err)
	}
	// -60300 + 15000 + 13000 = -32300
	if sum.TotalCashCents != -32300 {
		t.Fatalf("cash=%d want -32300", sum.TotalCashCents)
	}
	if sum.OpenReceivablesCents != 0 {
		t.Fatalf("open AR=%d", sum.OpenReceivablesCents)
	}
	stock, _ := st.ListItemsInStock()
	if len(stock) != 20 {
		t.Fatalf("stock left=%d want 20", len(stock))
	}
}
