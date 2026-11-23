package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"

	"github.com/puppe1990/leilao-erp/internal/store"
)

func setupFindsHandler(t *testing.T) (*FindsHandler, *store.SQLiteStore) {
	t.Helper()
	st, err := store.NewSQLiteStore(":memory:", "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	h := NewFindsHandler(setupTestRenderer(t), st, testSite(), cais.Config{Env: "development"}, setupTestInertia(t))
	return h, st
}

func TestFinds_IndexEmpty(t *testing.T) {
	h, _ := setupFindsHandler(t)
	req := inertiaRequest(http.MethodGet, "/finds", nil)
	rr := httptest.NewRecorder()
	h.Index(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	assertInertiaComponent(t, rr, "Finds/Index")
	rows := assertInertiaProp(t, rr, "finds")
	list, ok := rows.([]any)
	if !ok || len(list) != 0 {
		t.Fatalf("finds=%v", rows)
	}
}

func TestFinds_CreateUpdateDelete(t *testing.T) {
	h, st := setupFindsHandler(t)

	form := url.Values{}
	form.Set("name", "Dell P2016t 19,5")
	form.Set("url", "https://www.olx.com.br/vi/1555022453")
	form.Set("source", "olx")
	form.Set("buy", "175")
	form.Set("sell", "250")
	req := inertiaRequest(http.MethodPost, "/finds", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withUserSession(t, st, req)
	rr := httptest.NewRecorder()
	h.Create(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("create status=%d body=%s", rr.Code, rr.Body.String())
	}
	if loc := rr.Header().Get("Location"); loc != "/finds" {
		t.Fatalf("Location=%q", loc)
	}

	list, err := st.ListFinds("")
	if err != nil || len(list) != 1 {
		t.Fatalf("%+v %v", list, err)
	}
	id := list[0].ID
	if list[0].BuyCents != 17500 || list[0].SellCents != 25000 {
		t.Fatalf("cents %+v", list[0])
	}

	form2 := url.Values{}
	form2.Set("name", "Dell P2016t")
	form2.Set("url", "https://www.facebook.com/marketplace/item/1645626870456007/")
	form2.Set("source", "facebook")
	form2.Set("buy", "150")
	form2.Set("sell", "220")
	req2 := inertiaRequest(http.MethodPost, "/finds/1", strings.NewReader(form2.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2 = withUserSession(t, st, req2)
	rr2 := httptest.NewRecorder()
	h.Update(rr2, req2, id)
	if rr2.Code != http.StatusSeeOther {
		t.Fatalf("update status=%d body=%s", rr2.Code, rr2.Body.String())
	}
	found, _ := st.GetFind(id)
	if found.Name != "Dell P2016t" || found.Source != store.FindSourceFacebook || found.BuyCents != 15000 {
		t.Fatalf("%+v", found)
	}

	req3 := inertiaRequest(http.MethodPost, "/finds/1/delete", nil)
	req3 = withUserSession(t, st, req3)
	rr3 := httptest.NewRecorder()
	h.Destroy(rr3, req3, id)
	if rr3.Code != http.StatusSeeOther {
		t.Fatalf("delete status=%d", rr3.Code)
	}
	if _, err := st.GetFind(id); err == nil {
		t.Fatal("expected deleted")
	}
}

func TestFinds_CreateRejectsBadURL(t *testing.T) {
	h, st := setupFindsHandler(t)
	form := url.Values{}
	form.Set("name", "HP")
	form.Set("url", "olx.com.br/vi/1")
	form.Set("source", "olx")
	form.Set("buy", "60")
	form.Set("sell", "120")
	req := inertiaRequest(http.MethodPost, "/finds", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withUserSession(t, st, req)
	rr := httptest.NewRecorder()
	h.Create(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d want re-render", rr.Code)
	}
	assertInertiaErrors(t, rr, "form")
}

func TestFinds_IndexFiltersSourceAndShowsMargin(t *testing.T) {
	h, st := setupFindsHandler(t)
	_, err := st.CreateFind(store.FindInput{
		Name: "Kwara DeWalt", URL: "https://kwara.com.br/bens/1", Source: store.FindSourceKwara, BuyCents: 26975, SellCents: 40000,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = st.CreateFind(store.FindInput{
		Name: "OLX iPhone 8", URL: "https://www.olx.com.br/vi/2", Source: store.FindSourceOLX, BuyCents: 30000, SellCents: 30000,
	})
	if err != nil {
		t.Fatal(err)
	}

	req := inertiaRequest(http.MethodGet, "/finds?source=kwara", nil)
	rr := httptest.NewRecorder()
	h.Index(rr, req)
	rows := assertInertiaProp(t, rr, "finds").([]any)
	if len(rows) != 1 {
		t.Fatalf("filtered=%d", len(rows))
	}
	row := rows[0].(map[string]any)
	if row["name"] != "Kwara DeWalt" {
		t.Fatalf("%v", row)
	}
	if row["buy"] != "R$ 269,75" || row["sell"] != "R$ 400,00" {
		t.Fatalf("money %v", row)
	}
	if row["margin"] != "R$ 130,25" {
		t.Fatalf("margin %v", row)
	}
	if row["url"] != "https://kwara.com.br/bens/1" {
		t.Fatalf("url %v", row)
	}
}
