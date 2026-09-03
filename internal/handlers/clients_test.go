package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"

	"github.com/puppe1990/leilao-erp/internal/domain"
	"github.com/puppe1990/leilao-erp/internal/store"
)

func setupClientsHandler(t *testing.T) (*ClientsHandler, *store.SQLiteStore) {
	t.Helper()
	st, err := store.NewSQLiteStore(filepath.Join(t.TempDir(), "clients.db"), "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	h := NewClientsHandler(setupTestRenderer(t), st, testSite(), cais.Config{Env: "development"}, setupTestInertia(t))
	return h, st
}

// fakeCNPJLookuper stubs the external CNPJ API for handler tests.
type fakeCNPJLookuper struct {
	company domain.CompanyFromCNPJ
	err     error
}

func (f fakeCNPJLookuper) Lookup(*http.Request) (domain.CompanyFromCNPJ, error) {
	return f.company, f.err
}

func TestClients_CNPJLookup(t *testing.T) {
	h, _ := setupClientsHandler(t)
	h.cnpj = fakeCNPJLookuper{company: domain.CompanyFromCNPJ{
		CNPJ: "19.131.243/0001-97", LegalName: "OPEN KNOWLEDGE BRASIL", TradeName: "REDE PELO CONHECIMENTO LIVRE",
		Street: "PAULISTA", Number: "37", City: "SAO PAULO", State: "SP", CEP: "01.311-902",
		Phone: "(11) 23851939", Email: "contato@ok.org.br",
	}}

	req := httptest.NewRequest(http.MethodGet, "/clients/cnpj?cnpj=19.131.243/0001-97", nil)
	rr := httptest.NewRecorder()
	h.CNPJLookup(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
	}
	var got map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["cnpj"] != "19.131.243/0001-97" || got["name"] != "OPEN KNOWLEDGE BRASIL" || got["state"] != "SP" {
		t.Fatalf("payload = %v", got)
	}
}

func TestClients_CNPJLookup_Invalid(t *testing.T) {
	h, _ := setupClientsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/clients/cnpj?cnpj=123", nil)
	rr := httptest.NewRecorder()
	h.CNPJLookup(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "CNPJ inválido") {
		t.Fatalf("body = %s", rr.Body.String())
	}
}

func TestClients_CNPJLookup_NotFound(t *testing.T) {
	h, _ := setupClientsHandler(t)
	h.cnpj = fakeCNPJLookuper{err: domain.ErrCNPJNotFound}

	req := httptest.NewRequest(http.MethodGet, "/clients/cnpj?cnpj=19131243000197", nil)
	rr := httptest.NewRecorder()
	h.CNPJLookup(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rr.Code)
	}
}

func TestClients_CNPJLookup_UpstreamError(t *testing.T) {
	h, _ := setupClientsHandler(t)
	h.cnpj = fakeCNPJLookuper{err: errors.New("boom")}

	req := httptest.NewRequest(http.MethodGet, "/clients/cnpj?cnpj=19131243000197", nil)
	rr := httptest.NewRecorder()
	h.CNPJLookup(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Falha ao consultar") {
		t.Fatalf("body = %s", rr.Body.String())
	}
}

func TestClients_CreateUpdateDelete(t *testing.T) {
	h, st := setupClientsHandler(t)

	form := url.Values{}
	form.Set("name", "João Comprador")
	form.Set("phone", "11987654321")
	form.Set("email", "joao@ex.com")
	req := inertiaRequest(http.MethodPost, "/clients", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withUserSession(t, st, req)
	rr := httptest.NewRecorder()
	h.Create(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("create status=%d body=%s", rr.Code, rr.Body.String())
	}

	list, err := st.ListClients()
	if err != nil || len(list) != 1 {
		t.Fatalf("%+v %v", list, err)
	}
	id := list[0].ID

	form2 := url.Values{}
	form2.Set("name", "João C.")
	form2.Set("phone", "11000000000")
	req2 := inertiaRequest(http.MethodPost, "/clients/1", strings.NewReader(form2.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2 = withUserSession(t, st, req2)
	rr2 := httptest.NewRecorder()
	h.Update(rr2, req2, id)
	if rr2.Code != http.StatusSeeOther {
		t.Fatalf("update status=%d", rr2.Code)
	}
	found, _ := st.FindClient(id)
	if found.Name != "João C." {
		t.Fatalf("%+v", found)
	}

	req3 := inertiaRequest(http.MethodPost, "/clients/1/delete", nil)
	req3 = withUserSession(t, st, req3)
	rr3 := httptest.NewRecorder()
	h.Destroy(rr3, req3, id)
	if rr3.Code != http.StatusSeeOther {
		t.Fatalf("delete status=%d", rr3.Code)
	}
	if n, _ := st.ListClients(); len(n) != 0 {
		t.Fatal(n)
	}
}

func TestClients_CreateRequiresName(t *testing.T) {
	h, st := setupClientsHandler(t)
	form := url.Values{}
	form.Set("name", "  ")
	req := inertiaRequest(http.MethodPost, "/clients", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withUserSession(t, st, req)
	rr := httptest.NewRecorder()
	h.Create(rr, req)
	list, _ := st.ListClients()
	if len(list) != 0 {
		t.Fatal("should not create without name")
	}
}
