package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"

	"github.com/puppe1990/leilao-erp/internal/store"
)

func setupProductsHandler(t *testing.T) (*ProductsHandler, *store.SQLiteStore, string) {
	t.Helper()
	st, err := store.NewSQLiteStore(filepath.Join(t.TempDir(), "prod.db"), "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	staticDir := t.TempDir()
	h := NewProductsHandler(
		setupTestRenderer(t), st, testSite(), cais.Config{Env: "development"}, setupTestInertia(t),
	).WithStaticDir(staticDir)
	return h, st, staticDir
}

func seedProductID(t *testing.T, st *store.SQLiteStore) int64 {
	t.Helper()
	acc, err := st.InsertCashAccount("PIX", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}
	_, err = st.CreateLotPurchase(store.CreateLotInput{
		Name: "L", PurchasedAt: "2026-01-01", ItemTitle: "Monitor Midia", ItemQty: 1,
		CashAccountID: acc, PaidAt: "2026-01-01T12:00:00Z",
		Costs: []store.CostInput{{Label: "A", AmountCents: 1000, AlreadyPaid: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	ps, err := st.ListProducts()
	if err != nil || len(ps) == 0 {
		t.Fatal(err)
	}
	return ps[0].ID
}

func TestProducts_AddMedia_VideoURL(t *testing.T) {
	h, st, _ := setupProductsHandler(t)
	pid := seedProductID(t, st)

	form := url.Values{}
	form.Set("kind", "video")
	form.Set("url", "https://www.youtube.com/watch?v=abc")
	req := inertiaRequest(http.MethodPost, "/products/"+strconv.FormatInt(pid, 10)+"/media", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withUserSession(t, st, req)

	rr := httptest.NewRecorder()
	h.AddMedia(rr, req, pid)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	list, err := st.ListProductMedia(pid)
	if err != nil || len(list) != 1 || list[0].Kind != "video" {
		t.Fatalf("media=%+v err=%v", list, err)
	}
}

func TestProducts_AddMedia_PhotoURL_AndDelete(t *testing.T) {
	h, st, _ := setupProductsHandler(t)
	pid := seedProductID(t, st)

	form := url.Values{}
	form.Set("kind", "photo")
	form.Set("url", "/static/uploads/products/1/x.jpg")
	req := inertiaRequest(http.MethodPost, "/products/"+strconv.FormatInt(pid, 10)+"/media", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withUserSession(t, st, req)
	rr := httptest.NewRecorder()
	h.AddMedia(rr, req, pid)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("add status=%d body=%s", rr.Code, rr.Body.String())
	}
	list, _ := st.ListProductMedia(pid)
	if len(list) != 1 {
		t.Fatal(list)
	}
	mid := list[0].ID

	req2 := inertiaRequest(http.MethodPost, "/products/"+strconv.FormatInt(pid, 10)+"/media/"+strconv.FormatInt(mid, 10)+"/delete", nil)
	req2.SetPathValue("mediaId", strconv.FormatInt(mid, 10))
	req2 = withUserSession(t, st, req2)
	rr2 := httptest.NewRecorder()
	h.DestroyMedia(rr2, req2, pid)
	if rr2.Code != http.StatusSeeOther {
		t.Fatalf("del status=%d body=%s", rr2.Code, rr2.Body.String())
	}
	list, _ = st.ListProductMedia(pid)
	if len(list) != 0 {
		t.Fatalf("want empty after delete, got %d", len(list))
	}
}

func TestProducts_AddMedia_RejectsBadKind(t *testing.T) {
	h, st, _ := setupProductsHandler(t)
	pid := seedProductID(t, st)
	form := url.Values{}
	form.Set("kind", "audio")
	form.Set("url", "https://example.com/a.mp3")
	req := inertiaRequest(http.MethodPost, "/products/"+strconv.FormatInt(pid, 10)+"/media", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withUserSession(t, st, req)
	rr := httptest.NewRecorder()
	h.AddMedia(rr, req, pid)
	list, _ := st.ListProductMedia(pid)
	if len(list) != 0 {
		t.Fatal("should not insert audio")
	}
}
