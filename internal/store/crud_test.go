package store_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/puppe1990/leilao-erp/internal/store"
)

func testStore(t *testing.T) *store.SQLiteStore {
	t.Helper()
	st, err := store.NewSQLiteStore(filepath.Join(t.TempDir(), "crud.db"), "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func seedLot(t *testing.T, st *store.SQLiteStore, paid bool) (lotID, accountID int64) {
	t.Helper()
	accountID, err := st.InsertCashAccount("PIX", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}
	lotID, err = st.CreateLotPurchase(store.CreateLotInput{
		Name: "Lote CRUD", PurchasedAt: "2026-07-01", ItemTitle: "Item", ItemQty: 2,
		Costs:         []store.CostInput{{Label: "Arremate", AmountCents: 10000, AlreadyPaid: paid}},
		CashAccountID: accountID,
		PaidAt:        "2026-07-01T12:00:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}
	return lotID, accountID
}

func TestUpdateAndDeleteLot(t *testing.T) {
	st := testStore(t)
	lotID, _ := seedLot(t, st, true)

	if err := st.UpdateLot(lotID, store.UpdateLotInput{
		Name: "Lote Atualizado", PurchasedAt: "2026-07-02", Notes: "obs",
	}); err != nil {
		t.Fatal(err)
	}
	lot, err := st.FindLot(lotID)
	if err != nil || lot.Name != "Lote Atualizado" {
		t.Fatalf("lot=%+v err=%v", lot, err)
	}

	if err := st.DeleteLot(lotID); err != nil {
		t.Fatal(err)
	}
	if _, err := st.FindLot(lotID); !errors.Is(err, store.ErrNotFound) && err == nil {
		// FindLot may wrap — check not found
		if err == nil {
			t.Fatal("expected deleted")
		}
	}
}

func TestDeleteLot_RejectsSoldItems(t *testing.T) {
	st := testStore(t)
	lotID, accountID := seedLot(t, st, true)
	items, _ := st.ListItemsByLot(lotID)
	_, err := st.CreateSale(store.CreateSaleInput{
		ItemID: items[0].ID, SoldAt: "2026-07-10T10:00:00Z", Channel: "direct",
		GrossCents: 5000, PaymentStatus: "received", CashAccountID: accountID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := st.DeleteLot(lotID); !errors.Is(err, store.ErrCannotDelete) {
		t.Fatalf("err=%v", err)
	}
}

func TestCashAccountCRUD(t *testing.T) {
	st := testStore(t)
	id, err := st.InsertCashAccount("Banco", "bank", 1000)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.UpdateCashAccount(id, "Banco Inter", "bank", 2000); err != nil {
		t.Fatal(err)
	}
	a, err := st.FindCashAccount(id)
	if err != nil || a.Name != "Banco Inter" || a.OpeningBalanceCents != 2000 {
		t.Fatalf("%+v %v", a, err)
	}
	if err := st.DeleteCashAccount(id); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteCashAccount_WithEntries(t *testing.T) {
	st := testStore(t)
	_, accountID := seedLot(t, st, true) // creates cash out entry
	if err := st.DeleteCashAccount(accountID); !errors.Is(err, store.ErrCannotDelete) {
		t.Fatalf("err=%v", err)
	}
}

func TestUpdateSalePending(t *testing.T) {
	st := testStore(t)
	lotID, _ := seedLot(t, st, false)
	items, _ := st.ListItemsByLot(lotID)
	saleID, err := st.CreateSale(store.CreateSaleInput{
		ItemID: items[0].ID, SoldAt: "2026-07-10", Channel: "mercadolivre",
		GrossCents: 10000, FeeCents: 1000, ShippingCents: 0,
		PaymentStatus: "pending", DueOn: "2026-08-01",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := st.UpdateSale(saleID, store.UpdateSaleInput{
		SoldAt: "2026-07-11", Channel: "shopee", GrossCents: 12000, FeeCents: 500, ShippingCents: 0, DueOn: "2026-08-15",
	}); err != nil {
		t.Fatal(err)
	}
	sale, err := st.FindSaleByID(saleID)
	if err != nil || sale.NetCents != 11500 || sale.Channel != "shopee" {
		t.Fatalf("%+v %v", sale, err)
	}
}

func TestCancelPayableAndCreate(t *testing.T) {
	st := testStore(t)
	id, err := st.CreatePayable(store.CreatePayableInput{
		Description: "Frete avulso", AmountCents: 5000, DueOn: "2026-08-01",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := st.CancelPayable(id); err != nil {
		t.Fatal(err)
	}
	pays, _ := st.ListPayables()
	for _, p := range pays {
		if p.ID == id && p.Status != "cancelled" {
			t.Fatalf("status=%s", p.Status)
		}
	}
}

func TestUpdateItem(t *testing.T) {
	st := testStore(t)
	lotID, _ := seedLot(t, st, true)
	items, _ := st.ListItemsByLot(lotID)
	if err := st.UpdateItem(items[0].ID, "Monitor Dell", "SKU-1"); err != nil {
		t.Fatal(err)
	}
	it, err := st.FindItemByID(items[0].ID)
	if err != nil || it.Title != "Monitor Dell" {
		t.Fatalf("%+v %v", it, err)
	}
}
