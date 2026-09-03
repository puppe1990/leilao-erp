package domain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// CompanyFromCNPJ is the data the client form absorbs after a CNPJ lookup.
// Keep it in domain (no DB) so it is testable without the store.
type CompanyFromCNPJ struct {
	CNPJ         string
	LegalName    string
	TradeName    string
	Street       string
	Number       string
	Complement   string
	Neighborhood string
	City         string
	State        string
	CEP          string
	Phone        string
	Email        string
}

// cnpjLookupDoer is satisfied by *http.Client; tests inject a fake.
type cnpjLookupDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type cnpjResponse struct {
	CNPJ         string `json:"cnpj"`
	RazaoSocial  string `json:"razao_social"`
	NomeFantasia string `json:"nome_fantasia"`
	Logradouro   string `json:"logradouro"`
	Numero       string `json:"numero"`
	Complemento  string `json:"complemento"`
	Bairro       string `json:"bairro"`
	Municipio    string `json:"municipio"`
	UF           string `json:"uf"`
	CEP          string `json:"cep"`
	Email        string `json:"email"`
	DDDTelefone1 string `json:"ddd_telefone_1"`
	Telefone1    string `json:"telefone_1"`
}

// ErrCNPJNotFound is returned when the Receita has no record for the CNPJ.
var ErrCNPJNotFound = errors.New("CNPJ não encontrado")

// CNPJLookupClient performs CNPJ lookups against BrasilAPI.
type CNPJLookupClient struct {
	client cnpjLookupDoer
}

// NewCNPJLookupClient builds a client with a short timeout (AGENTS: keep requests short).
func NewCNPJLookupClient() *CNPJLookupClient {
	return &CNPJLookupClient{client: &http.Client{Timeout: 5 * time.Second}}
}

// Lookup fetches company data from the free BrasilAPI CNPJ endpoint.
// Only digits from cnpj are used; BrazilAPI tolerates raw digits.
func (h *CNPJLookupClient) Lookup(ctx context.Context, cnpj string) (CompanyFromCNPJ, error) {
	digits := DigitsOnly(cnpj)
	if len(digits) != 14 {
		return CompanyFromCNPJ{}, fmt.Errorf("CNPJ %q deve ter 14 dígitos", cnpj)
	}
	url := "https://brasilapi.com.br/api/cnpj/v1/" + digits
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return CompanyFromCNPJ{}, fmt.Errorf("montar requisição CNPJ: %w", err)
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return CompanyFromCNPJ{}, fmt.Errorf("consultar CNPJ %s: %w", digits, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNotFound {
		return CompanyFromCNPJ{}, ErrCNPJNotFound
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return CompanyFromCNPJ{}, fmt.Errorf("BrasilAPI respondeu %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	var raw cnpjResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return CompanyFromCNPJ{}, fmt.Errorf("decodificar resposta da BrasilAPI: %w", err)
	}
	return normalizeCNPJResult(raw, digits), nil
}

func normalizeCNPJResult(raw cnpjResponse, digits string) CompanyFromCNPJ {
	cep := strings.TrimSpace(raw.CEP)
	if len(cep) == 8 {
		cep = cep[0:2] + "." + cep[2:5] + "-" + cep[5:8]
	}
	phone := ""
	if ddd := strings.TrimSpace(raw.DDDTelefone1); ddd != "" {
		phone = "(" + ddd + ") " + strings.TrimSpace(raw.Telefone1)
	} else {
		phone = strings.TrimSpace(raw.Telefone1)
	}
	return CompanyFromCNPJ{
		CNPJ:         FormatCNPJOrRaw(digits),
		LegalName:    strings.TrimSpace(raw.RazaoSocial),
		TradeName:    strings.TrimSpace(raw.NomeFantasia),
		Street:       strings.TrimSpace(raw.Logradouro),
		Number:       strings.TrimSpace(raw.Numero),
		Complement:   strings.TrimSpace(raw.Complemento),
		Neighborhood: strings.TrimSpace(raw.Bairro),
		City:         strings.TrimSpace(raw.Municipio),
		State:        strings.TrimSpace(raw.UF),
		CEP:          cep,
		Phone:        phone,
		Email:        strings.TrimSpace(raw.Email),
	}
}

func FormatCNPJOrRaw(s string) string {
	if f, err := FormatCNPJ(s); err == nil {
		return f
	}
	return s
}
