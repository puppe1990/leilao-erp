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
	if got := fmt.Sprint(assertInertiaProp(t, rr, "shopShowQty")); got != "true" {
		t.Fatalf("shopShowQty=%s want true", got)
	}
	if !strings.Contains(rr.Body.String(), slug) {
		t.Fatalf("expected slug %q in index payload", slug)
	}
	if got := assertInertiaProp(t, rr, "companyName"); got != "Pix Eletro" {
		t.Fatalf("companyName=%v, want Pix Eletro", got)
	}
	og, ok := assertInertiaProp(t, rr, "og").(map[string]any)
	if !ok {
		t.Fatalf("og prop type %T", assertInertiaProp(t, rr, "og"))
	}
	img, _ := og["image"].(string)
	if !strings.Contains(img, "og-pixeletro-2") {
		t.Fatalf("og image %q should bust cache with og-pixeletro-2", img)
	}
	title, _ := og["title"].(string)
	if !strings.Contains(title, "Eletro e tech") {
		t.Fatalf("og title %q should say Eletro e tech", title)
	}
	if strings.Contains(title, "Catálogo de monitores") {
		t.Fatalf("og title still monitor-only: %q", title)
	}
	desc, _ := og["description"].(string)
	if !strings.Contains(desc, "Fotos reais") {
		t.Fatalf("og description %q should mention real photos", desc)
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
	if got := fmt.Sprint(assertInertiaProp(t, rr, "shopShowQty")); got != "true" {
		t.Fatalf("shopShowQty=%s want true", got)
	}
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

func TestShopHandler_Show_NoPhoto_VisibleWithStock_OK(t *testing.T) {
	h, s := newShopHandler(t)
	acc, err := s.InsertCashAccount("PIX", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateLotPurchase(store.CreateLotInput{
		Name: "Lote sem foto", PurchasedAt: "2026-01-01",
		ItemTitle: "Monitor Sem Foto Detalhe", ItemQty: 1,
		CashAccountID: acc, PaidAt: "2026-01-01T12:00:00Z",
		Costs: []store.CostInput{{Label: "Arremate", AmountCents: 4000, AlreadyPaid: true}},
	}); err != nil {
		t.Fatal(err)
	}
	products, err := s.ListProducts()
	if err != nil || len(products) == 0 {
		t.Fatalf("products %v %v", products, err)
	}
	p := products[0]
	if err := s.UpdateProductShopVisible(p.ID, true); err != nil {
		t.Fatal(err)
	}

	req := inertiaRequest(http.MethodGet, "/produto/"+p.Slug, nil)
	rr := httptest.NewRecorder()
	h.Show(rr, req, p.Slug)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d want 200 body=%s", rr.Code, rr.Body.String())
	}
	assertInertiaComponent(t, rr, "Shop/Show")
}

func seedLeilaoShopLot(t *testing.T, s *store.SQLiteStore) {
	t.Helper()
	acc, err := s.InsertCashAccount("PIX", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}
	lotID, err := s.CreateLotPurchase(store.CreateLotInput{
		Name: "Monitores — leilão Jul/2026", PurchasedAt: "2026-07-01",
		ItemTitle: "Monitor Lote A", ItemQty: 2,
		CashAccountID: acc, PaidAt: "2026-07-01T12:00:00Z",
		Costs: []store.CostInput{{Label: "Arremate", AmountCents: 20000, AlreadyPaid: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	items, err := s.ListItemsByLot(lotID)
	if err != nil || len(items) < 2 {
		t.Fatalf("items %v %v", items, err)
	}
	hint := int64(12900)
	if err := s.UpdateItem(items[1].ID, store.UpdateItemInput{Title: "Monitor Lote B", SalePriceHintCents: &hint}); err != nil {
		t.Fatal(err)
	}
	products, err := s.ListProducts()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range products {
		if _, err := s.AddProductMedia(p.ID, store.ProductMediaInput{
			Kind: "photo", URL: "/static/uploads/products/1/lot.jpg",
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := s.AddProductMedia(p.ID, store.ProductMediaInput{
			Kind: "video", URL: "/static/uploads/products/1/lot.mp4",
		}); err != nil {
			t.Fatal(err)
		}
		_ = s.UpdateProductShopVisible(p.ID, true)
	}
	if err := s.SetWhatsAppPhone("11999990000"); err != nil {
		t.Fatal(err)
	}
}

func TestShopHandler_Index_IncludesLotOffer(t *testing.T) {
	h, s := newShopHandler(t)
	seedLeilaoShopLot(t, s)

	req := inertiaRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.Index(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	offer := assertInertiaProp(t, rr, "lotOffer")
	m, ok := offer.(map[string]any)
	if !ok || m == nil {
		t.Fatalf("lotOffer not object: %T %v", offer, offer)
	}
	if fmt.Sprint(m["qty"]) == "0" {
		t.Fatalf("lotOffer qty empty: %v", m)
	}
	if !strings.Contains(rr.Body.String(), "lot.jpg") {
		t.Fatal("expected lot photo in index payload")
	}
	if !strings.Contains(rr.Body.String(), "/produto/"+store.ShopLotSlug) {
		t.Fatal("expected lot product href in index payload")
	}
	prods := assertInertiaProp(t, rr, "products")
	list, ok := prods.([]any)
	if !ok || len(list) < 1 {
		t.Fatalf("products %T %v", prods, prods)
	}
	first, _ := list[0].(map[string]any)
	if fmt.Sprint(first["slug"]) != store.ShopLotSlug {
		t.Fatalf("first product slug=%v want %s", first["slug"], store.ShopLotSlug)
	}
	if fmt.Sprint(first["isLot"]) != "true" {
		t.Fatalf("first product isLot=%v", first["isLot"])
	}
}

func TestShopHandler_Show_LotSlug_OK(t *testing.T) {
	h, s := newShopHandler(t)
	seedLeilaoShopLot(t, s)

	req := inertiaRequest(http.MethodGet, "/produto/"+store.ShopLotSlug, nil)
	rr := httptest.NewRecorder()
	h.Show(rr, req, store.ShopLotSlug)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	assertInertiaComponent(t, rr, "Shop/Lot")
	if got := fmt.Sprint(assertInertiaProp(t, rr, "shopShowQty")); got != "true" {
		t.Fatalf("shopShowQty=%s want true", got)
	}
	offer := assertInertiaProp(t, rr, "lotOffer")
	m, ok := offer.(map[string]any)
	if !ok {
		t.Fatalf("lotOffer %T", offer)
	}
	if fmt.Sprint(m["href"]) != "/produto/"+store.ShopLotSlug {
		t.Fatalf("href=%v", m["href"])
	}
	photos, _ := m["photos"].([]any)
	videos, _ := m["videos"].([]any)
	if len(photos) < 1 {
		t.Fatalf("want photos, got %v", m["photos"])
	}
	if len(videos) < 1 {
		t.Fatalf("want videos, got %v", m["videos"])
	}
}

func TestShopHandler_Index_shopShowQtyOff(t *testing.T) {
	h, s := newShopHandler(t)
	seedShopProduct(t, s)
	if err := s.SetShopShowQty(false); err != nil {
		t.Fatal(err)
	}
	req := inertiaRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.Index(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if got := fmt.Sprint(assertInertiaProp(t, rr, "shopShowQty")); got != "false" {
		t.Fatalf("shopShowQty=%s want false", got)
	}
}

func TestShopHandler_Index_OmitsLotWhenDisabled(t *testing.T) {
	h, s := newShopHandler(t)
	seedLeilaoShopLot(t, s)
	if err := s.SetShopLotEnabled(false); err != nil {
		t.Fatal(err)
	}

	req := inertiaRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.Index(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	offer := assertInertiaProp(t, rr, "lotOffer")
	if offer != nil {
		t.Fatalf("lotOffer=%v want nil when disabled", offer)
	}
	prods := assertInertiaProp(t, rr, "products")
	list, ok := prods.([]any)
	if !ok {
		t.Fatalf("products %T", prods)
	}
	for _, row := range list {
		m, _ := row.(map[string]any)
		if fmt.Sprint(m["isLot"]) == "true" || fmt.Sprint(m["slug"]) == store.ShopLotSlug {
			t.Fatalf("catalog still has lot card: %v", m)
		}
	}
}

func TestShopHandler_Show_LotSlug_NotFoundWhenDisabled(t *testing.T) {
	h, s := newShopHandler(t)
	seedLeilaoShopLot(t, s)
	if err := s.SetShopLotEnabled(false); err != nil {
		t.Fatal(err)
	}
	req := inertiaRequest(http.MethodGet, "/produto/"+store.ShopLotSlug, nil)
	rr := httptest.NewRecorder()
	h.Show(rr, req, store.ShopLotSlug)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status=%d want 404 body=%s", rr.Code, rr.Body.String())
	}
}

func TestShopHandler_Lot_NotFoundWithoutOffer(t *testing.T) {
	h, _ := newShopHandler(t)
	req := inertiaRequest(http.MethodGet, "/produto/"+store.ShopLotSlug, nil)
	rr := httptest.NewRecorder()
	h.Show(rr, req, store.ShopLotSlug)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status=%d want 404", rr.Code)
	}
}
