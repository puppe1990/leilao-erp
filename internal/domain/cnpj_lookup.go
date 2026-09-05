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
	CNPJ                      string `json:"cnpj"`
	RazaoSocial               string `json:"razao_social"`
	NomeFantasia              string `json:"nome_fantasia"`
	Logradouro                string `json:"logradouro"`
	Numero                    string `json:"numero"`
	Complemento               string `json:"complemento"`
	Bairro                    string `json:"bairro"`
	Municipio                 string `json:"municipio"`
	UF                        string `json:"uf"`
	CEP                       string `json:"cep"`
	Email                     string `json:"email"`
	DDDTelefone1              string `json:"ddd_telefone_1"`
	Telefone1                 string `json:"telefone_1"`
	DescricaoTipoDeLogradouro string `json:"descricao_tipo_de_logradouro"`
}

// openCNPJResponse is the public OpenCNPJ payload. BrasilAPI (Minha Receita)
// often ships MEI rows with CEP+bairro and a blank logradouro; OpenCNPJ still
// has tipo/logradouro/numero/email/telefone from the same RFB files.
type openCNPJResponse struct {
	RazaoSocial    string          `json:"razao_social"`
	NomeFantasia   string          `json:"nome_fantasia"`
	TipoLogradouro string          `json:"tipo_logradouro"`
	Logradouro     string          `json:"logradouro"`
	Numero         string          `json:"numero"`
	Complemento    string          `json:"complemento"`
	Bairro         string          `json:"bairro"`
	Municipio      string          `json:"municipio"`
	UF             string          `json:"uf"`
	CEP            string          `json:"cep"`
	Email          string          `json:"email"`
	Telefones      []openCNPJPhone `json:"telefones"`
}

type openCNPJPhone struct {
	DDD    string `json:"ddd"`
	Numero string `json:"numero"`
	IsFax  bool   `json:"is_fax"`
}

// ErrCNPJNotFound is returned when the Receita has no record for the CNPJ.
var ErrCNPJNotFound = errors.New("CNPJ não encontrado")

// CNPJLookupClient looks up a CNPJ on BrasilAPI and fills gaps via OpenCNPJ.
type CNPJLookupClient struct {
	client cnpjLookupDoer
}

// NewCNPJLookupClient builds a client with a short timeout (AGENTS: keep requests short).
func NewCNPJLookupClient() *CNPJLookupClient {
	return &CNPJLookupClient{client: &http.Client{Timeout: 5 * time.Second}}
}

// Lookup fetches company data from BrasilAPI, then fills blank street/contact
// fields from OpenCNPJ. Only digits from cnpj are used.
func (h *CNPJLookupClient) Lookup(ctx context.Context, cnpj string) (CompanyFromCNPJ, error) {
	digits := DigitsOnly(cnpj)
	if len(digits) != 14 {
		return CompanyFromCNPJ{}, fmt.Errorf("CNPJ %q deve ter 14 dígitos", cnpj)
	}
	brasil, brasilErr := h.lookupBrasilAPI(ctx, digits)
	if brasilErr == nil && brasil.Street != "" {
		return brasil, nil
	}
	open, openErr := h.lookupOpenCNPJ(ctx, digits)
	if openErr == nil {
		if brasilErr == nil {
			return fillCompanyGaps(brasil, open), nil
		}
		return open, nil
	}
	if brasilErr != nil {
		return CompanyFromCNPJ{}, brasilErr
	}
	return brasil, nil
}

func (h *CNPJLookupClient) lookupBrasilAPI(ctx context.Context, digits string) (CompanyFromCNPJ, error) {
	resp, err := h.get(ctx, "https://brasilapi.com.br/api/cnpj/v1/"+digits)
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

func (h *CNPJLookupClient) lookupOpenCNPJ(ctx context.Context, digits string) (CompanyFromCNPJ, error) {
	resp, err := h.get(ctx, "https://api.opencnpj.org/"+digits)
	if err != nil {
		return CompanyFromCNPJ{}, fmt.Errorf("consultar OpenCNPJ %s: %w", digits, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNotFound {
		return CompanyFromCNPJ{}, ErrCNPJNotFound
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return CompanyFromCNPJ{}, fmt.Errorf("OpenCNPJ respondeu %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	var raw openCNPJResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return CompanyFromCNPJ{}, fmt.Errorf("decodificar resposta da OpenCNPJ: %w", err)
	}
	return normalizeOpenCNPJResult(raw, digits), nil
}

func (h *CNPJLookupClient) get(ctx context.Context, rawURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("montar requisição CNPJ: %w", err)
	}
	return h.client.Do(req)
}

func normalizeCNPJResult(raw cnpjResponse, digits string) CompanyFromCNPJ {
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
		Street:       composeStreet(raw.DescricaoTipoDeLogradouro, raw.Logradouro),
		Number:       strings.TrimSpace(raw.Numero),
		Complement:   strings.TrimSpace(raw.Complemento),
		Neighborhood: strings.TrimSpace(raw.Bairro),
		City:         strings.TrimSpace(raw.Municipio),
		State:        strings.TrimSpace(raw.UF),
		CEP:          formatCEP(raw.CEP),
		Phone:        phone,
		Email:        strings.TrimSpace(raw.Email),
	}
}

func normalizeOpenCNPJResult(raw openCNPJResponse, digits string) CompanyFromCNPJ {
	return CompanyFromCNPJ{
		CNPJ:         FormatCNPJOrRaw(digits),
		LegalName:    strings.TrimSpace(raw.RazaoSocial),
		TradeName:    strings.TrimSpace(raw.NomeFantasia),
		Street:       composeStreet(raw.TipoLogradouro, raw.Logradouro),
		Number:       strings.TrimSpace(raw.Numero),
		Complement:   strings.TrimSpace(raw.Complemento),
		Neighborhood: strings.TrimSpace(raw.Bairro),
		City:         strings.TrimSpace(raw.Municipio),
		State:        strings.TrimSpace(raw.UF),
		CEP:          formatCEP(raw.CEP),
		Phone:        phoneFromOpenCNPJ(raw.Telefones),
		Email:        strings.TrimSpace(raw.Email),
	}
}

func composeStreet(tipo, logradouro string) string {
	tipo = strings.TrimSpace(tipo)
	logradouro = strings.TrimSpace(logradouro)
	if logradouro == "" {
		return ""
	}
	if tipo == "" {
		return logradouro
	}
	upperLog := strings.ToUpper(logradouro)
	upperTipo := strings.ToUpper(tipo)
	if upperLog == upperTipo || strings.HasPrefix(upperLog, upperTipo+" ") {
		return logradouro
	}
	return tipo + " " + logradouro
}

func formatCEP(cep string) string {
	cep = strings.TrimSpace(cep)
	if len(cep) == 8 {
		return cep[0:2] + "." + cep[2:5] + "-" + cep[5:8]
	}
	return cep
}

func phoneFromOpenCNPJ(phones []openCNPJPhone) string {
	for _, p := range phones {
		if p.IsFax {
			continue
		}
		ddd := strings.TrimSpace(p.DDD)
		num := strings.TrimSpace(p.Numero)
		if ddd == "" && num == "" {
			continue
		}
		if ddd != "" {
			return "(" + ddd + ") " + num
		}
		return num
	}
	return ""
}

func fillCompanyGaps(base, extra CompanyFromCNPJ) CompanyFromCNPJ {
	if base.LegalName == "" {
		base.LegalName = extra.LegalName
	}
	if base.TradeName == "" {
		base.TradeName = extra.TradeName
	}
	if base.Street == "" {
		base.Street = extra.Street
	}
	if base.Number == "" {
		base.Number = extra.Number
	}
	if base.Complement == "" {
		base.Complement = extra.Complement
	}
	if base.Neighborhood == "" {
		base.Neighborhood = extra.Neighborhood
	}
	if base.City == "" {
		base.City = extra.City
	}
	if base.State == "" {
		base.State = extra.State
	}
	if base.CEP == "" {
		base.CEP = extra.CEP
	}
	if base.Phone == "" {
		base.Phone = extra.Phone
	}
	if base.Email == "" {
		base.Email = extra.Email
	}
	return base
}

func FormatCNPJOrRaw(s string) string {
	if f, err := FormatCNPJ(s); err == nil {
		return f
	}
	return s
}
