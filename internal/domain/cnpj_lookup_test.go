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
	return &CNPJLookupClient{
		client: &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: status,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		})},
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
