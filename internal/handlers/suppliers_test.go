package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/puppe1990/cais/pkg/cais"

	"github.com/puppe1990/leilao-erp/internal/store"
)

func setupSuppliersHandler(t *testing.T) (*SuppliersHandler, *store.SQLiteStore) {
	t.Helper()
	st, err := store.NewSQLiteStore(":memory:", "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	h := NewSuppliersHandler(setupTestRenderer(t), st, testSite(), cais.Config{Env: "development"}, setupTestInertia(t))
	return h, st
}

func TestSuppliers_IndexEmpty(t *testing.T) {
	h, _ := setupSuppliersHandler(t)
	req := inertiaRequest(http.MethodGet, "/suppliers", nil)
	rr := httptest.NewRecorder()
	h.Index(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	assertInertiaComponent(t, rr, "Suppliers/Index")
	rows := assertInertiaProp(t, rr, "suppliers")
	list, ok := rows.([]any)
	if !ok || len(list) != 0 {
		t.Fatalf("suppliers=%v", rows)
	}
}

func TestSuppliers_CreateUpdateDelete(t *testing.T) {
	h, st := setupSuppliersHandler(t)

	form := url.Values{}
	form.Set("name", "PMCELL São Paulo")
	form.Set("whatsapp", "(11) 91130-4693")
	form.Set("notes", "power bank, carregador 25W")
	form.Set("min_order", "300")
	req := inertiaRequest(http.MethodPost, "/suppliers", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withUserSession(t, st, req)
	rr := httptest.NewRecorder()
	h.Create(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("create status=%d body=%s", rr.Code, rr.Body.String())
	}
	if loc := rr.Header().Get("Location"); loc != "/suppliers" {
		t.Fatalf("Location=%q", loc)
	}

	list, err := st.ListSuppliers()
	if err != nil || len(list) != 1 {
		t.Fatalf("%+v %v", list, err)
	}
	id := list[0].ID
	if list[0].ChatID != "5511911304693@c.us" || list[0].MinOrderCents != 30000 {
		t.Fatalf("created %+v", list[0])
	}

	form2 := url.Values{}
	form2.Set("name", "PMCELL")
	form2.Set("whatsapp", "5511911304693")
	form2.Set("notes", "25 de Março")
	form2.Set("min_order", "300")
	req2 := inertiaRequest(http.MethodPost, "/suppliers/1", strings.NewReader(form2.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2 = withUserSession(t, st, req2)
	rr2 := httptest.NewRecorder()
	h.Update(rr2, req2, id)
	if rr2.Code != http.StatusSeeOther {
		t.Fatalf("update status=%d body=%s", rr2.Code, rr2.Body.String())
	}
	found, _ := st.GetSupplier(id)
	if found.Name != "PMCELL" || found.Notes != "25 de Março" {
		t.Fatalf("%+v", found)
	}

	req3 := inertiaRequest(http.MethodPost, "/suppliers/1/delete", nil)
	req3 = withUserSession(t, st, req3)
	rr3 := httptest.NewRecorder()
	h.Destroy(rr3, req3, id)
	if rr3.Code != http.StatusSeeOther {
		t.Fatalf("delete status=%d", rr3.Code)
	}
	if _, err := st.GetSupplier(id); err == nil {
		t.Fatal("expected deleted")
	}
}

func TestSuppliers_CreateRejectsBadWhatsApp(t *testing.T) {
	h, st := setupSuppliersHandler(t)
	form := url.Values{}
	form.Set("name", "Loja")
	form.Set("whatsapp", "abc")
	form.Set("notes", "")
	form.Set("min_order", "0")
	req := inertiaRequest(http.MethodPost, "/suppliers", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withUserSession(t, st, req)
	rr := httptest.NewRecorder()
	h.Create(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d want re-render", rr.Code)
	}
	assertInertiaErrors(t, rr, "whatsapp")
}

func TestSuppliers_ShowConversation(t *testing.T) {
	h, st := setupSuppliersHandler(t)
	id, err := st.CreateSupplier(store.SupplierInput{
		Name: "Awei 25 de Março", WhatsApp: "11972711435", Notes: "fone TWS",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = st.SaveSupplierMessage(id, store.SupplierMessageInput{
		WAMessageID: "true_5511972711435@c.us_AAA",
		FromMe:      true,
		Body:        "Vocês têm tabela de atacado e pedido mínimo?",
		SentAt:      time.Date(2026, 11, 24, 14, 4, 29, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}

	req := inertiaRequest(http.MethodGet, "/suppliers/1", nil)
	rr := httptest.NewRecorder()
	h.Show(rr, req, id)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	assertInertiaComponent(t, rr, "Suppliers/Show")
	sup := assertInertiaProp(t, rr, "supplier").(map[string]any)
	if sup["name"] != "Awei 25 de Março" {
		t.Fatalf("supplier %v", sup)
	}
	if sup["minOrder"] != "R$ 0,00" {
		t.Fatalf("minOrder %v", sup)
	}
	msgs := assertInertiaProp(t, rr, "messages").([]any)
	if len(msgs) != 1 {
		t.Fatalf("messages=%d", len(msgs))
	}
	msg := msgs[0].(map[string]any)
	if msg["fromMe"] != true {
		t.Fatalf("fromMe %v", msg)
	}
	if msg["body"] != "Vocês têm tabela de atacado e pedido mínimo?" {
		t.Fatalf("body %v", msg)
	}
}

func TestSuppliers_IndexShowsLastMessage(t *testing.T) {
	h, st := setupSuppliersHandler(t)
	id, err := st.CreateSupplier(store.SupplierInput{Name: "Atacado 25", WhatsApp: "11950849676"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = st.SaveSupplierMessage(id, store.SupplierMessageInput{
		WAMessageID: "true_x_1",
		FromMe:      true,
		Body:        "pedido mínimo?",
		SentAt:      time.Date(2026, 11, 24, 14, 4, 7, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}

	req := inertiaRequest(http.MethodGet, "/suppliers", nil)
	rr := httptest.NewRecorder()
	h.Index(rr, req)
	rows := assertInertiaProp(t, rr, "suppliers").([]any)
	if len(rows) != 1 {
		t.Fatalf("rows=%d", len(rows))
	}
	row := rows[0].(map[string]any)
	if row["name"] != "Atacado 25" {
		t.Fatalf("%v", row)
	}
	if row["lastBody"] != "pedido mínimo?" {
		t.Fatalf("lastBody %v", row)
	}
	if row["messageCount"] != float64(1) {
		t.Fatalf("messageCount %v", row)
	}
}
