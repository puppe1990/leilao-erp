package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/meta"
	inertia "github.com/romsar/gonertia/v3"

	"github.com/puppe1990/leilao-erp/internal/domain"
	"github.com/puppe1990/leilao-erp/internal/store"
)

// CNPJLookuper fetches company data from an external CNPJ API.
type CNPJLookuper interface {
	Lookup(r *http.Request) (domain.CompanyFromCNPJ, error)
}

type ClientsHandler struct {
	renderer *cais.Renderer
	store    store.Store
	site     meta.Site
	cfg      cais.Config
	inertia  *inertia.Inertia
	cnpj     CNPJLookuper
}

func NewClientsHandler(renderer *cais.Renderer, s store.Store, site meta.Site, cfg cais.Config, i *inertia.Inertia) *ClientsHandler {
	return &ClientsHandler{renderer: renderer, store: s, site: site, cfg: cfg, inertia: i, cnpj: cnpjHTTPLookuper{}}
}

// cnpjHTTPLookuper is the production adapter for domain.CNPJLookupClient.
type cnpjHTTPLookuper struct{}

func (cnpjHTTPLookuper) Lookup(r *http.Request) (domain.CompanyFromCNPJ, error) {
	return domain.NewCNPJLookupClient().Lookup(r.Context(), r.URL.Query().Get("cnpj"))
}

// CNPJLookup resolves a CNPJ via the injected lookuper and returns JSON for the form.
func (h *ClientsHandler) CNPJLookup(w http.ResponseWriter, r *http.Request) {
	cnpj := strings.TrimSpace(r.URL.Query().Get("cnpj"))
	if !domain.ValidCNPJ(cnpj) {
		writeJSONError(w, http.StatusBadRequest, "CNPJ inválido — confira os 14 dígitos")
		return
	}
	company, err := h.cnpj.Lookup(r)
	if err != nil {
		status := http.StatusBadGateway
		if errors.Is(err, domain.ErrCNPJNotFound) {
			status = http.StatusNotFound
		}
		writeJSONError(w, status, friendlyCNPJError(err))
		return
	}
	writeJSON(w, map[string]any{
		"cnpj":         company.CNPJ,
		"legal_name":   company.LegalName,
		"trade_name":   company.TradeName,
		"name":         company.LegalName,
		"street":       company.Street,
		"number":       company.Number,
		"complement":   company.Complement,
		"neighborhood": company.Neighborhood,
		"city":         company.City,
		"state":        company.State,
		"cep":          company.CEP,
		"phone":        company.Phone,
		"email":        company.Email,
	})
}

// friendlyCNPJError maps raw lookup errors to a user-safe message.
func friendlyCNPJError(err error) string {
	switch {
	case strings.Contains(err.Error(), "não encontrado"):
		return "CNPJ não encontrado na base da Receita Federal"
	case strings.Contains(err.Error(), "deve ter 14 dígitos"):
		return "CNPJ inválido — confira os 14 dígitos"
	default:
		return "Falha ao consultar o CNPJ — tente novamente em instantes"
	}
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// Index lists clients with optional ?q= search.
func (h *ClientsHandler) Index(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))

	list, err := h.store.SearchClients(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows := make([]map[string]any, 0, len(list))
	for _, c := range list {
		rows = append(rows, map[string]any{
			"id":           c.ID,
			"name":         c.Name,
			"type":         c.Type,
			"phone":        c.Phone,
			"email":        c.Email,
			"document":     c.Document,
			"cep":          c.Address.CEP,
			"street":       c.Address.Street,
			"number":       c.Address.Number,
			"complement":   c.Address.Complement,
			"neighborhood": c.Address.Neighborhood,
			"city":         c.Address.City,
			"state":        c.Address.State,
			"notes":        c.Notes,
		})
	}

	_ = h.inertia.Render(w, r, "Clients/Index", withCompany(h.store, inertia.Props{
		"site":    meta.ForRequest(h.site, r),
		"clients": rows,
		"query":   q,
	}))
}

// Create inserts a client.
func (h *ClientsHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := parseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	in := store.ClientInput{
		Name:     r.FormValue("name"),
		Phone:    r.FormValue("phone"),
		Email:    r.FormValue("email"),
		Document: r.FormValue("document"),
		Notes:    r.FormValue("notes"),
	}
	if _, err := h.store.CreateClient(in); err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}
	h.inertia.Redirect(w, r, "/clients", http.StatusSeeOther)
}

// Update updates a client.
func (h *ClientsHandler) Update(w http.ResponseWriter, r *http.Request, id int64) {
	if err := parseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	in := store.ClientInput{
		Name:     r.FormValue("name"),
		Phone:    r.FormValue("phone"),
		Email:    r.FormValue("email"),
		Document: r.FormValue("document"),
		Notes:    r.FormValue("notes"),
	}
	if err := h.store.UpdateClient(id, in); err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}
	h.inertia.Redirect(w, r, "/clients", http.StatusSeeOther)
}

// Destroy deletes a client.
func (h *ClientsHandler) Destroy(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.store.DeleteClient(id); err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}
	h.inertia.Redirect(w, r, "/clients", http.StatusSeeOther)
}
