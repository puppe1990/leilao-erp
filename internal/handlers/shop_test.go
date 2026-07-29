package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"

	"github.com/puppe1990/leilao-erp/internal/store"
)

func newShopHandler(t *testing.T) (*ShopHandler, *store.SQLiteStore) {
	t.Helper()
	s, err := store.NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	h := NewShopHandler(s, testSite(), cais.Config{}, setupTestInertia(t))
	return h, s
}

func seedShopProduct(t *testing.T, s *store.SQLiteStore) (id int64, slug string) {
	t.Helper()
	acc, err := s.InsertCashAccount("PIX", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.CreateLotPurchase(store.CreateLotInput{
		Name: "Lote catálogo", PurchasedAt: "2026-01-01",
		ItemTitle: "Monitor Catálogo TDD", ItemQty: 1,
		CashAccountID: acc, PaidAt: "2026-01-01T12:00:00Z",
		Costs: []store.CostInput{{Label: "Arremate", AmountCents: 4000, AlreadyPaid: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	products, err := s.ListProducts()
	if err != nil || len(products) == 0 {
		t.Fatalf("products %v %v", products, err)
	}
	id = products[0].ID
	slug = products[0].Slug
	if slug == "" {
		t.Fatal("expected product slug")
	}
	if _, err := s.AddProductMedia(id, store.ProductMediaInput{
		Kind: "photo", URL: "/static/uploads/products/1/catalogo.jpg", SortOrder: 0,
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.UpdateProductShopVisible(id, true); err != nil {
		t.Fatal(err)
	}
	if err := s.SetWhatsAppPhone("11999990000"); err != nil {
		t.Fatal(err)
	}
	return id, slug
}

func TestShopHandler_Index_PublicOK(t *testing.T) {
	h, s := newShopHandler(t)
	_, slug := seedShopProduct(t, s)

	req := inertiaRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.Index(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	assertInertiaComponent(t, rr, "Shop/Index")
	assertInertiaProp(t, rr, "products")
	assertInertiaProp(t, rr, "whatsappSet")
	if !strings.Contains(rr.Body.String(), slug) {
		t.Fatalf("expected slug %q in index payload", slug)
	}
}

func TestShopHandler_Show_OK(t *testing.T) {
	h, s := newShopHandler(t)
	_, slug := seedShopProduct(t, s)

	req := inertiaRequest(http.MethodGet, "/produto/"+slug, nil)
	rr := httptest.NewRecorder()
	h.Show(rr, req, slug)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	assertInertiaComponent(t, rr, "Shop/Show")
	assertInertiaProp(t, rr, "product")
}

func TestShopHandler_Show_IDRedirectsToSlug(t *testing.T) {
	h, s := newShopHandler(t)
	id, slug := seedShopProduct(t, s)

	req := inertiaRequest(http.MethodGet, fmt.Sprintf("/produto/%d", id), nil)
	rr := httptest.NewRecorder()
	h.Show(rr, req, fmt.Sprintf("%d", id))

	if rr.Code != http.StatusMovedPermanently {
		t.Fatalf("status=%d want 301 body=%s", rr.Code, rr.Body.String())
	}
	loc := rr.Header().Get("Location")
	if loc != "/produto/"+slug {
		t.Fatalf("Location=%q want /produto/%s", loc, slug)
	}
}

func TestShopHandler_Show_IncludesVideos(t *testing.T) {
	h, s := newShopHandler(t)
	id, slug := seedShopProduct(t, s)
	if _, err := s.AddProductMedia(id, store.ProductMediaInput{
		Kind: "video", URL: "/static/uploads/products/1/demo.mp4", SortOrder: 1,
	}); err != nil {
		t.Fatal(err)
	}

	req := inertiaRequest(http.MethodGet, "/produto/"+slug, nil)
	rr := httptest.NewRecorder()
	h.Show(rr, req, slug)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "demo.mp4") {
		t.Fatalf("expected video URL in payload, body sample: %.200s", body)
	}
	if !strings.Contains(body, `"videos"`) {
		t.Fatal("expected videos prop")
	}
}

func TestShopHandler_Show_NoPhoto_NotFound(t *testing.T) {
	h, s := newShopHandler(t)
	id, err := s.EnsureProductByName("Sem foto shop", "principal", nil)
	if err != nil {
		t.Fatal(err)
	}
	p, err := s.FindProduct(id)
	if err != nil {
		t.Fatal(err)
	}

	req := inertiaRequest(http.MethodGet, "/produto/"+p.Slug, nil)
	rr := httptest.NewRecorder()
	h.Show(rr, req, p.Slug)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status=%d want 404", rr.Code)
	}
}

func TestShopHandler_Show_NotShopVisible_NotFound(t *testing.T) {
	h, s := newShopHandler(t)
	id, slug := seedShopProduct(t, s)
	if err := s.UpdateProductShopVisible(id, false); err != nil {
		t.Fatal(err)
	}

	req := inertiaRequest(http.MethodGet, "/produto/"+slug, nil)
	rr := httptest.NewRecorder()
	h.Show(rr, req, slug)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status=%d want 404", rr.Code)
	}
}
