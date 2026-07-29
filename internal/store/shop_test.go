package store

import (
	"testing"
)

func TestListProductsWithPhotos_OnlyWithPhotoAndStock(t *testing.T) {
	st := newTestStore(t)

	// product A: photo + stock (via lot)
	acc, err := st.InsertCashAccount("PIX", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}
	_, err = st.CreateLotPurchase(CreateLotInput{
		Name: "Lote shop", PurchasedAt: "2026-01-01",
		ItemTitle: "Monitor Shop A", ItemQty: 1,
		CashAccountID: acc, PaidAt: "2026-01-01T12:00:00Z",
		Costs: []CostInput{{Label: "Arremate", AmountCents: 5000, AlreadyPaid: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	products, err := st.ListProducts()
	if err != nil || len(products) == 0 {
		t.Fatalf("list: %v %v", products, err)
	}
	withStock := products[0].ID
	if _, err := st.AddProductMedia(withStock, ProductMediaInput{
		Kind: "photo", URL: "/static/uploads/products/1/a.jpg", SortOrder: 0,
	}); err != nil {
		t.Fatal(err)
	}
	if err := st.UpdateProductShopVisible(withStock, true); err != nil {
		t.Fatal(err)
	}

	// product B: no photo, has stock (another lot item title = new product)
	_, err = st.CreateLotPurchase(CreateLotInput{
		Name: "Lote B", PurchasedAt: "2026-01-02",
		ItemTitle: "Monitor Shop B", ItemQty: 1,
		CashAccountID: acc, PaidAt: "2026-01-02T12:00:00Z",
		Costs: []CostInput{{Label: "Arremate", AmountCents: 3000, AlreadyPaid: true}},
	})
	if err != nil {
		t.Fatal(err)
	}

	// product C: photo but no stock
	noStock, err := st.EnsureProductByName("Monitor Shop C no stock", "principal", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.AddProductMedia(noStock, ProductMediaInput{
		Kind: "photo", URL: "/static/uploads/products/c/c.jpg", SortOrder: 0,
	}); err != nil {
		t.Fatal(err)
	}

	list, err := st.ListProductsWithPhotos()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("want 1 catalog item with photo+stock, got %d %+v", len(list), list)
	}
	if list[0].ID != withStock {
		t.Fatalf("id=%d want %d", list[0].ID, withStock)
	}
	if list[0].FirstPhotoURL == "" {
		t.Fatal("expected FirstPhotoURL")
	}
	if list[0].QtyInStock < 1 {
		t.Fatalf("qty=%d", list[0].QtyInStock)
	}
	if !list[0].ShopVisible {
		t.Fatal("expected ShopVisible")
	}

	// Hide from shop even with photo+stock
	if err := st.UpdateProductShopVisible(withStock, false); err != nil {
		t.Fatal(err)
	}
	list, err = st.ListProductsWithPhotos()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("want 0 when shop_visible=0, got %d", len(list))
	}
}

func TestWhatsAppPhoneSetting(t *testing.T) {
	st := newTestStore(t)
	phone, err := st.WhatsAppPhone()
	if err != nil {
		t.Fatal(err)
	}
	if phone != "" {
		t.Fatalf("default empty, got %q", phone)
	}
	if err := st.SetWhatsAppPhone("11 98765-4321"); err != nil {
		t.Fatal(err)
	}
	phone, err = st.WhatsAppPhone()
	if err != nil {
		t.Fatal(err)
	}
	if phone != "11 98765-4321" {
		t.Fatalf("got %q", phone)
	}
}
