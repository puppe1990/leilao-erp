package store

import (
	"testing"
)

func seedProduct(t *testing.T, st *SQLiteStore) int64 {
	t.Helper()
	accID, err := st.InsertCashAccount("Caixa", "cash", 0)
	if err != nil {
		t.Fatal(err)
	}
	_, err = st.CreateLotPurchase(CreateLotInput{
		Name:          "Lote midia",
		PurchasedAt:   "2026-01-01",
		ItemTitle:     "Monitor Teste Midia 22\"",
		ItemQty:       1,
		CashAccountID: accID,
		PaidAt:        "2026-01-01T12:00:00Z",
		Costs:         []CostInput{{Label: "Arremate", AmountCents: 10000, AlreadyPaid: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	products, err := st.ListProducts()
	if err != nil || len(products) == 0 {
		t.Fatalf("products: %v %v", products, err)
	}
	return products[0].ID
}

func TestAddListDeleteProductMedia_PhotoAndVideo(t *testing.T) {
	st := newTestStore(t)
	productID := seedProduct(t, st)

	// empty list
	list, err := st.ListProductMedia(productID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("want empty media, got %d", len(list))
	}

	// photo by path (local upload)
	photoID, err := st.AddProductMedia(productID, ProductMediaInput{
		Kind:      "photo",
		URL:       "/static/uploads/products/1/foto.jpg",
		SortOrder: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if photoID <= 0 {
		t.Fatal("photo id expected")
	}

	// video by external URL
	videoID, err := st.AddProductMedia(productID, ProductMediaInput{
		Kind:      "video",
		URL:       "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		SortOrder: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	list, err = st.ListProductMedia(productID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("want 2 media, got %d", len(list))
	}
	if list[0].Kind != "photo" || list[0].URL == "" {
		t.Errorf("first media photo: %+v", list[0])
	}
	if list[1].Kind != "video" {
		t.Errorf("second media video: %+v", list[1])
	}
	if list[0].ProductID != productID || list[1].ProductID != productID {
		t.Error("product_id mismatch")
	}

	// reject bad kind
	if _, err := st.AddProductMedia(productID, ProductMediaInput{Kind: "audio", URL: "x"}); err == nil {
		t.Fatal("expected error for kind audio")
	}
	// reject empty url
	if _, err := st.AddProductMedia(productID, ProductMediaInput{Kind: "photo", URL: "  "}); err == nil {
		t.Fatal("expected error for empty url")
	}
	// reject missing product
	if _, err := st.AddProductMedia(999999, ProductMediaInput{Kind: "photo", URL: "/x.jpg"}); err == nil {
		t.Fatal("expected error for missing product")
	}

	// delete photo
	if err := st.DeleteProductMedia(photoID); err != nil {
		t.Fatal(err)
	}
	list, err = st.ListProductMedia(productID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != videoID {
		t.Fatalf("after delete want only video %d, got %+v", videoID, list)
	}

	// delete missing
	if err := st.DeleteProductMedia(photoID); err == nil {
		t.Fatal("expected not found on second delete")
	}
}

func TestListProducts_IncludesMediaCounts(t *testing.T) {
	st := newTestStore(t)
	productID := seedProduct(t, st)
	_, _ = st.AddProductMedia(productID, ProductMediaInput{Kind: "photo", URL: "/a.jpg"})
	_, _ = st.AddProductMedia(productID, ProductMediaInput{Kind: "photo", URL: "/b.jpg"})
	_, _ = st.AddProductMedia(productID, ProductMediaInput{Kind: "video", URL: "https://youtu.be/x"})

	products, err := st.ListProducts()
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, p := range products {
		if p.ID == productID {
			found = true
			if p.PhotoCount != 2 {
				t.Errorf("PhotoCount=%d want 2", p.PhotoCount)
			}
			if p.VideoCount != 1 {
				t.Errorf("VideoCount=%d want 1", p.VideoCount)
			}
		}
	}
	if !found {
		t.Fatal("product not listed")
	}
}
