package domain

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

// roundTripFunc lets tests stub the BrasilAPI HTTP call without a live server.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func stubCNPJClient(status int, body string) *CNPJLookupClient {
	return stubCNPJRoutes(map[string]stubCNPJRoute{
		"": {status: status, body: body},
	})
}

type stubCNPJRoute struct {
	status int
	body   string
}

// stubCNPJRoutes dispatches by substring of the request URL (first match wins).
// The empty key is the fallback used by the original single-body stub.
func stubCNPJRoutes(routes map[string]stubCNPJRoute) *CNPJLookupClient {
	return &CNPJLookupClient{
		client: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			url := r.URL.String()
			var fallback *stubCNPJRoute
			for needle, route := range routes {
				if needle == "" {
					cp := route
					fallback = &cp
					continue
				}
				if strings.Contains(url, needle) {
					return jsonResponse(route.status, route.body), nil
				}
			}
			if fallback != nil {
				return jsonResponse(fallback.status, fallback.body), nil
			}
			return jsonResponse(http.StatusNotFound, `{}`), nil
		})},
	}
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func TestCNPJLookup_Ok(t *testing.T) {
	c := stubCNPJClient(http.StatusOK, `{
		"cnpj":"19131243000197","razao_social":"OPEN KNOWLEDGE BRASIL",
		"nome_fantasia":"REDE PELO CONHECIMENTO LIVRE",
		"logradouro":"PAULISTA","numero":"37","complemento":"ANDAR 4",
		"bairro":"BELA VISTA","municipio":"SAO PAULO","uf":"SP","cep":"01311902",
		"ddd_telefone_1":"11","telefone_1":"23851939","email":"contato@ok.org.br"
	}`)

	got, err := c.Lookup(context.Background(), "19.131.243/0001-97")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}

	want := CompanyFromCNPJ{
		CNPJ:         "19.131.243/0001-97",
		LegalName:    "OPEN KNOWLEDGE BRASIL",
		TradeName:    "REDE PELO CONHECIMENTO LIVRE",
		Street:       "PAULISTA",
		Number:       "37",
		Complement:   "ANDAR 4",
		Neighborhood: "BELA VISTA",
		City:         "SAO PAULO",
		State:        "SP",
		CEP:          "01.311-902",
		Phone:        "(11) 23851939",
		Email:        "contato@ok.org.br",
	}
	if got != want {
		t.Errorf("Lookup() = %+v\nwant %+v", got, want)
	}
}

func TestCNPJLookup_StreetKeepsNumberWhenMerged(t *testing.T) {
	c := stubCNPJClient(http.StatusOK, `{
		"cnpj":"19131243000197","razao_social":"OPEN KNOWLEDGE BRASIL",
		"logradouro":"PAULISTA 37","numero":"37"
	}`)

	got, err := c.Lookup(context.Background(), "19131243000197")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if got.Street != "PAULISTA 37" {
		t.Errorf("Street = %q, want PAULISTA 37", got.Street)
	}
}

func TestCNPJLookup_NameFallsBackToTradeName(t *testing.T) {
	c := stubCNPJClient(http.StatusOK, `{"cnpj":"19131243000197","razao_social":"","nome_fantasia":"FANTASIA LTDA"}`)

	got, err := c.Lookup(context.Background(), "19131243000197")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if got.LegalName != "" {
		t.Errorf("LegalName = %q, want empty (razao missing)", got.LegalName)
	}
	if got.TradeName != "FANTASIA LTDA" {
		t.Errorf("TradeName = %q, want FANTASIA LTDA", got.TradeName)
	}
}

func TestCNPJLookup_InvalidDigits(t *testing.T) {
	c := stubCNPJClient(http.StatusOK, `{}`)
	if _, err := c.Lookup(context.Background(), "123"); err == nil {
		t.Fatal("Lookup() with 3 digits should error")
	}
}

func TestCNPJLookup_NotFound(t *testing.T) {
	c := stubCNPJClient(http.StatusNotFound, `{"message":"CNPJ inválido"}`)

	_, err := c.Lookup(context.Background(), "19131243000197")
	if err == nil || !strings.Contains(err.Error(), "não encontrado") {
		t.Fatalf("Lookup() error = %v, want not-found message", err)
	}
}

func TestCNPJLookup_FillsStreetFromOpenCNPJWhenBrasilAPIOmitsIt(t *testing.T) {
	// Receita/BrasilAPI often ships MEI records with CEP+bairro and a blank logradouro.
	c := stubCNPJRoutes(map[string]stubCNPJRoute{
		"brasilapi.com.br": {
			status: http.StatusOK,
			body: `{
				"cnpj":"24490987000138","razao_social":"MATHEUS NUNES PUPPE 02399708024",
				"logradouro":"","numero":"","complemento":"","email":null,"ddd_telefone_1":"",
				"bairro":"BELENZINHO","municipio":"SAO PAULO","uf":"SP","cep":"03058000"
			}`,
		},
		"api.opencnpj.org": {
			status: http.StatusOK,
			body: `{
				"cnpj":"24490987000138","razao_social":"MATHEUS NUNES PUPPE 02399708024",
				"tipo_logradouro":"RUA","logradouro":"CONSELHEIRO COTEGIPE","numero":"219",
				"complemento":"APT 114B","bairro":"BELENZINHO","municipio":"SAO PAULO","uf":"SP",
				"cep":"03058000","email":"matheus.puppe90@hotmail.com",
				"telefones":[{"ddd":"51","numero":"93701099","is_fax":false}]
			}`,
		},
	})

	got, err := c.Lookup(context.Background(), "24.490.987/0001-38")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if got.Street != "RUA CONSELHEIRO COTEGIPE" {
		t.Errorf("Street = %q, want RUA CONSELHEIRO COTEGIPE", got.Street)
	}
	if got.Number != "219" {
		t.Errorf("Number = %q, want 219", got.Number)
	}
	if got.Complement != "APT 114B" {
		t.Errorf("Complement = %q, want APT 114B", got.Complement)
	}
	if got.Email != "matheus.puppe90@hotmail.com" {
		t.Errorf("Email = %q", got.Email)
	}
	if got.Phone != "(51) 93701099" {
		t.Errorf("Phone = %q, want (51) 93701099", got.Phone)
	}
	if got.Neighborhood != "BELENZINHO" || got.City != "SAO PAULO" || got.State != "SP" {
		t.Errorf("address remainder = %+v", got)
	}
}

func TestCNPJLookup_SkipsOpenCNPJWhenBrasilAPIHasStreet(t *testing.T) {
	c := stubCNPJRoutes(map[string]stubCNPJRoute{
		"brasilapi.com.br": {
			status: http.StatusOK,
			body: `{
				"cnpj":"19131243000197","razao_social":"OPEN KNOWLEDGE BRASIL",
				"logradouro":"PAULISTA","numero":"37","municipio":"SAO PAULO","uf":"SP"
			}`,
		},
		"api.opencnpj.org": {status: http.StatusInternalServerError, body: `{"error":"nope"}`},
	})

	got, err := c.Lookup(context.Background(), "19131243000197")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if got.Street != "PAULISTA" || got.Number != "37" {
		t.Errorf("Lookup() = %+v, want BrasilAPI street/number", got)
	}
}

func TestCNPJLookup_NetworkError(t *testing.T) {
	c := &CNPJLookupClient{client: &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return nil, context.DeadlineExceeded
	})}}

	_, err := c.Lookup(context.Background(), "19131243000197")
	if err == nil {
		t.Fatal("Lookup() should error on transport failure")
	}
	if !strings.Contains(err.Error(), "consultar CNPJ") {
		t.Errorf("Lookup() error = %v, want wrapped context", err)
	}
}
