package store

import (
	"testing"
)

func seedLeilaoAndPessoalLots(t *testing.T, st *SQLiteStore) {
	t.Helper()
	acc, err := st.InsertCashAccount("PIX", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}
	lotID, err := st.CreateLotPurchase(CreateLotInput{
		Name: "Monitores — leilão Jul/2026", PurchasedAt: "2026-07-01",
		ItemTitle: "Monitor A", ItemQty: 3,
		CashAccountID: acc, PaidAt: "2026-07-01T12:00:00Z",
		Costs: []CostInput{{Label: "Arremate", AmountCents: 60000, AlreadyPaid: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	items, err := st.ListItemsByLot(lotID)
	if err != nil || len(items) != 3 {
		t.Fatalf("items: %d %v", len(items), err)
	}
	hintA, hintB, hintC := int64(10000), int64(20000), int64(30000)
	if err := st.UpdateItem(items[0].ID, UpdateItemInput{Title: "Monitor A", SalePriceHintCents: &hintA}); err != nil {
		t.Fatal(err)
	}
	if err := st.UpdateItem(items[1].ID, UpdateItemInput{Title: "Monitor B", SalePriceHintCents: &hintB}); err != nil {
		t.Fatal(err)
	}
	if err := st.UpdateItem(items[2].ID, UpdateItemInput{Title: "Monitor C", SalePriceHintCents: &hintC}); err != nil {
		t.Fatal(err)
	}
	if _, err := st.CreateSale(CreateSaleInput{
		ItemID: items[2].ID, SoldAt: "2026-08-01", Channel: "direct",
		GrossCents: 30000, PaymentStatus: "received", CashAccountID: acc,
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := st.CreateLotPurchase(CreateLotInput{
		Name: "Pessoal — não leilão", PurchasedAt: "2026-08-01",
		ItemTitle: "Monitor LG 29", ItemQty: 1,
		CashAccountID: acc, PaidAt: "2026-08-01T12:00:00Z",
		Costs: []CostInput{{Label: "Próprio", AmountCents: 50000, AlreadyPaid: true}},
	}); err != nil {
		t.Fatal(err)
	}

	products, err := st.ListProducts()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range products {
		switch p.Name {
		case "Monitor A":
			if _, err := st.AddProductMedia(p.ID, ProductMediaInput{
				Kind: "photo", URL: "/static/uploads/products/1/a.jpg",
			}); err != nil {
				t.Fatal(err)
			}
			if _, err := st.AddProductMedia(p.ID, ProductMediaInput{
				Kind: "video", URL: "/static/uploads/products/1/a.mp4",
			}); err != nil {
				t.Fatal(err)
			}
			_ = st.UpdateProductShopVisible(p.ID, true)
		case "Monitor B":
			if _, err := st.AddProductMedia(p.ID, ProductMediaInput{
				Kind: "photo", URL: "/static/uploads/products/2/b.jpg",
			}); err != nil {
				t.Fatal(err)
			}
			_ = st.UpdateProductShopVisible(p.ID, true)
		case "Monitor C":
			if _, err := st.AddProductMedia(p.ID, ProductMediaInput{
				Kind: "photo", URL: "/static/uploads/products/3/c.jpg",
			}); err != nil {
				t.Fatal(err)
			}
		case "Monitor LG 29":
			if _, err := st.AddProductMedia(p.ID, ProductMediaInput{
				Kind: "photo", URL: "/static/uploads/products/22/lg.jpg",
			}); err != nil {
				t.Fatal(err)
			}
			_ = st.UpdateProductShopVisible(p.ID, true)
		}
	}
}

func TestShopLotOffer_LeilaoRemainingExcludesPessoalAndSold(t *testing.T) {
	st := newTestStore(t)
	seedLeilaoAndPessoalLots(t, st)

	offer, err := st.ShopLotOffer()
	if err != nil {
		t.Fatal(err)
	}
	if offer == nil {
		t.Fatal("expected lot offer")
	}
	if offer.Qty != 2 {
		t.Fatalf("qty=%d want 2 (A+B in stock; C sold; LG pessoal excluded)", offer.Qty)
	}
	if offer.PriceCents != ShopLotPriceCents {
		t.Fatalf("price=%d want %d", offer.PriceCents, ShopLotPriceCents)
	}
	if offer.RetailCents != 30000 {
		t.Fatalf("retail=%d want 30000 (10000+20000)", offer.RetailCents)
	}
	if len(offer.Items) != 2 {
		t.Fatalf("items=%d want 2 %+v", len(offer.Items), offer.Items)
	}
	if len(offer.Photos) != 2 {
		t.Fatalf("photos=%d want 2 (A+B; not sold C, not pessoal LG) %+v", len(offer.Photos), offer.Photos)
	}
	if len(offer.Videos) != 1 {
		t.Fatalf("videos=%d want 1 %+v", len(offer.Videos), offer.Videos)
	}
	for _, m := range append(append([]ShopLotMedia{}, offer.Photos...), offer.Videos...) {
		if m.ProductName == "Monitor LG 29" || m.ProductName == "Monitor C" {
			t.Fatalf("lot media leaked %q", m.ProductName)
		}
		if m.URL == "" {
			t.Fatal("media url empty")
		}
	}
}

func TestShopLotOffer_ExcludesHiddenCatalogProductsKeepsPrice(t *testing.T) {
	st := newTestStore(t)
	acc, err := st.InsertCashAccount("PIX", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}
	lotID, err := st.CreateLotPurchase(CreateLotInput{
		Name: "Monitores — leilão Jul/2026", PurchasedAt: "2026-07-01",
		ItemTitle: "Monitor Keep A", ItemQty: 3,
		CashAccountID: acc, PaidAt: "2026-07-01T12:00:00Z",
		Costs: []CostInput{{Label: "Arremate", AmountCents: 60000, AlreadyPaid: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	items, err := st.ListItemsByLot(lotID)
	if err != nil || len(items) != 3 {
		t.Fatalf("items: %d %v", len(items), err)
	}
	hintA, hintB, hintHide := int64(10000), int64(20000), int64(27900)
	if err := st.UpdateItem(items[0].ID, UpdateItemInput{Title: "Monitor Keep A", SalePriceHintCents: &hintA}); err != nil {
		t.Fatal(err)
	}
	if err := st.UpdateItem(items[1].ID, UpdateItemInput{Title: "Monitor Keep B", SalePriceHintCents: &hintB}); err != nil {
		t.Fatal(err)
	}
	if err := st.UpdateItem(items[2].ID, UpdateItemInput{Title: "Monitor Hide Me", SalePriceHintCents: &hintHide}); err != nil {
		t.Fatal(err)
	}
	products, err := st.ListProducts()
	if err != nil {
		t.Fatal(err)
	}
	var hideID int64
	for _, p := range products {
		if _, err := st.AddProductMedia(p.ID, ProductMediaInput{
			Kind: "photo", URL: "/static/uploads/products/1/" + p.Slug + ".jpg",
		}); err != nil {
			t.Fatal(err)
		}
		if err := st.UpdateProductShopVisible(p.ID, true); err != nil {
			t.Fatal(err)
		}
		if p.Name == "Monitor Hide Me" {
			hideID = p.ID
		}
	}
	if hideID == 0 {
		t.Fatal("Monitor Hide Me not found")
	}
	if err := st.UpdateProductShopVisible(hideID, false); err != nil {
		t.Fatal(err)
	}

	offer, err := st.ShopLotOffer()
	if err != nil {
		t.Fatal(err)
	}
	if offer == nil {
		t.Fatal("expected lot of remaining visible units")
	}
	if offer.Qty != 2 {
		t.Fatalf("qty=%d want 2 (hidden unit excluded)", offer.Qty)
	}
	if offer.PriceCents != ShopLotPriceCents {
		t.Fatalf("price=%d want %d", offer.PriceCents, ShopLotPriceCents)
	}
	for _, it := range offer.Items {
		if it.Name == "Monitor Hide Me" {
			t.Fatal("hidden product leaked into lot items")
		}
	}
	for _, m := range offer.Photos {
		if m.ProductName == "Monitor Hide Me" {
			t.Fatal("hidden product leaked into lot photos")
		}
	}
}

func TestShopLotEnabled_DefaultTrue(t *testing.T) {
	st := newTestStore(t)
	on, err := st.ShopLotEnabled()
	if err != nil {
		t.Fatal(err)
	}
	if !on {
		t.Fatal("want shop lot enabled when setting is missing")
	}
}

func TestShopLotOffer_NilWhenDisabled(t *testing.T) {
	st := newTestStore(t)
	seedLeilaoAndPessoalLots(t, st)
	if err := st.SetShopLotEnabled(false); err != nil {
		t.Fatal(err)
	}
	on, err := st.ShopLotEnabled()
	if err != nil || on {
		t.Fatalf("enabled=%v err=%v", on, err)
	}
	offer, err := st.ShopLotOffer()
	if err != nil {
		t.Fatal(err)
	}
	if offer != nil {
		t.Fatalf("want nil offer when shop lot is off, got %+v", offer)
	}
}

func TestShopLotOffer_ReturnsAfterReenable(t *testing.T) {
	st := newTestStore(t)
	seedLeilaoAndPessoalLots(t, st)
	if err := st.SetShopLotEnabled(false); err != nil {
		t.Fatal(err)
	}
	if err := st.SetShopLotEnabled(true); err != nil {
		t.Fatal(err)
	}
	offer, err := st.ShopLotOffer()
	if err != nil {
		t.Fatal(err)
	}
	if offer == nil || offer.Qty != 2 {
		t.Fatalf("want lot offer after re-enable, got %+v", offer)
	}
}

func TestShopLotOffer_NilWhenFewerThanTwo(t *testing.T) {
	st := newTestStore(t)
	acc, err := st.InsertCashAccount("PIX", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.CreateLotPurchase(CreateLotInput{
		Name: "Monitores — leilão Jul/2026", PurchasedAt: "2026-07-01",
		ItemTitle: "Monitor Único", ItemQty: 1,
		CashAccountID: acc, PaidAt: "2026-07-01T12:00:00Z",
		Costs: []CostInput{{Label: "Arremate", AmountCents: 10000, AlreadyPaid: true}},
	}); err != nil {
		t.Fatal(err)
	}
	offer, err := st.ShopLotOffer()
	if err != nil {
		t.Fatal(err)
	}
	if offer != nil {
		t.Fatalf("want nil offer with 1 remaining, got %+v", offer)
	}
}

func TestListProductsWithPhotos_IncludesVisibleWithoutPhoto(t *testing.T) {
	st := newTestStore(t)
	acc, err := st.InsertCashAccount("PIX", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.CreateLotPurchase(CreateLotInput{
		Name: "Lote sem foto", PurchasedAt: "2026-01-01",
		ItemTitle: "Monitor Sem Foto Shop", ItemQty: 1,
		CashAccountID: acc, PaidAt: "2026-01-01T12:00:00Z",
		Costs: []CostInput{{Label: "Arremate", AmountCents: 4000, AlreadyPaid: true}},
	}); err != nil {
		t.Fatal(err)
	}
	products, err := st.ListProducts()
	if err != nil || len(products) == 0 {
		t.Fatalf("products %v %v", products, err)
	}
	id := products[0].ID
	if err := st.UpdateProductShopVisible(id, true); err != nil {
		t.Fatal(err)
	}
	list, err := st.ListProductsWithPhotos()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("want 1 visible in-stock product without photo, got %d", len(list))
	}
	if list[0].ID != id {
		t.Fatalf("id=%d want %d", list[0].ID, id)
	}
	if list[0].PhotoCount != 0 {
		t.Fatalf("photoCount=%d want 0", list[0].PhotoCount)
	}
}
