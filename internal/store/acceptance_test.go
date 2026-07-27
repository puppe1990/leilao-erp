package store_test

import (
	"path/filepath"
	"testing"

	"github.com/puppe1990/leilao-erp/internal/db"
	"github.com/puppe1990/leilao-erp/internal/models"
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
	// Arremate 60300 + Uber 1435 + Uber 1452 + Lalamove 5476 = 68663
	const seedCash int64 = -68663
	if sum.TotalCashCents != seedCash {
		t.Fatalf("cash after seed=%d want %d", sum.TotalCashCents, seedCash)
	}

	items, err := st.ListItemsInStock()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 22 {
		t.Fatalf("stock=%d want 22", len(items))
	}
	monitors := stockByTitle(items, "Monitor")
	if len(monitors) != 22 {
		t.Fatalf("monitors in stock=%d want 22", len(monitors))
	}
	accs, err := st.ListCashAccounts()
	if err != nil || len(accs) == 0 {
		t.Fatalf("accounts %v %v", accs, err)
	}

	// Direct sale PIX (monitor only)
	_, err = st.CreateSale(store.CreateSaleInput{
		ItemID: monitors[0].ID, SoldAt: "2026-07-27T12:00:00Z", Channel: "direct",
		GrossCents: 15000, PaymentStatus: "received", CashAccountID: accs[0].ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Marketplace pending then settle (monitor only)
	items, _ = st.ListItemsInStock()
	monitors = stockByTitle(items, "Monitor")
	_, err = st.CreateSale(store.CreateSaleInput{
		ItemID: monitors[0].ID, SoldAt: "2026-07-27T13:00:00Z", Channel: "mercadolivre",
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
	// -68663 + 15000 + 13000 = -40663
	if sum.TotalCashCents != seedCash+15000+13000 {
		t.Fatalf("cash=%d want %d", sum.TotalCashCents, seedCash+15000+13000)
	}
	if sum.OpenReceivablesCents != 0 {
		t.Fatalf("open AR=%d", sum.OpenReceivablesCents)
	}
	stock, _ := st.ListItemsInStock()
	if len(stockByTitle(stock, "Monitor")) != 20 {
		t.Fatalf("monitors left=%d want 20 (total stock=%d)", len(stockByTitle(stock, "Monitor")), len(stock))
	}
}

func stockByTitle(items []models.Item, title string) []models.Item {
	var out []models.Item
	for _, it := range items {
		if it.Title == title {
			out = append(out, it)
		}
	}
	return out
}
