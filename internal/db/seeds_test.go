package db

import (
	"testing"

	"github.com/puppe1990/leilao-erp/internal/store"
)

func newTestStore(t *testing.T) *store.SQLiteStore {
	t.Helper()
	s, err := store.NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestRunSeeds_idempotentMonitorsLot(t *testing.T) {
	s := newTestStore(t)

	if err := RunSeeds(s); err != nil {
		t.Fatalf("first seed: %v", err)
	}
	if err := RunSeeds(s); err != nil {
		t.Fatalf("second seed: %v", err)
	}

	accounts, err := s.ListCashAccounts()
	if err != nil {
		t.Fatal(err)
	}
	var pixCount int
	for _, a := range accounts {
		if a.Name == seedPIXAccountName {
			pixCount++
		}
	}
	if pixCount != 1 {
		t.Fatalf("PIX principal count = %d, want 1", pixCount)
	}

	lots, err := s.ListLots()
	if err != nil {
		t.Fatal(err)
	}
	var monitorLots []store.LotListItem
	for _, lot := range lots {
		if lot.Name == seedMonitorsLotName {
			monitorLots = append(monitorLots, lot)
		}
	}
	if len(monitorLots) != 1 {
		t.Fatalf("monitors lots = %d, want 1 after re-seed", len(monitorLots))
	}

	items, err := s.ListItemsByLot(monitorLots[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 22 {
		t.Fatalf("items = %d, want 22 after re-seed", len(items))
	}

	var sum int64
	for _, it := range items {
		sum += it.UnitCostCents
		if it.Title != "Monitor" {
			t.Errorf("title = %q, want Monitor", it.Title)
		}
		if it.Status != "in_stock" {
			t.Errorf("status = %q, want in_stock", it.Status)
		}
	}
	if sum != 60300 {
		t.Fatalf("unit cost sum = %d, want 60300", sum)
	}

	bal, err := s.CashBalance(accounts[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	// Seeded twice must not double the cash outflow.
	if bal != -60300 {
		t.Fatalf("cash balance = %d, want -60300 (idempotent paid arremate)", bal)
	}
}

func TestRunSeeds_createsPIXWhenMissing(t *testing.T) {
	s := newTestStore(t)

	if err := RunSeeds(s); err != nil {
		t.Fatal(err)
	}

	accounts, err := s.ListCashAccounts()
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) == 0 {
		t.Fatal("expected PIX principal account")
	}
	if accounts[0].Name != seedPIXAccountName {
		t.Errorf("name = %q, want %q", accounts[0].Name, seedPIXAccountName)
	}
	if accounts[0].Kind != "pix" {
		t.Errorf("kind = %q, want pix", accounts[0].Kind)
	}
}
