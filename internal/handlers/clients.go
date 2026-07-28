package handlers

import (
	"net/http"
	"strings"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/meta"
	inertia "github.com/romsar/gonertia/v3"

	"github.com/puppe1990/leilao-erp/internal/store"
)

type ClientsHandler struct {
	renderer *cais.Renderer
	store    store.Store
	site     meta.Site
	cfg      cais.Config
	inertia  *inertia.Inertia
}

func NewClientsHandler(renderer *cais.Renderer, s store.Store, site meta.Site, cfg cais.Config, i *inertia.Inertia) *ClientsHandler {
	return &ClientsHandler{renderer: renderer, store: s, site: site, cfg: cfg, inertia: i}
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
			"id":       c.ID,
			"name":     c.Name,
			"phone":    c.Phone,
			"email":    c.Email,
			"document": c.Document,
			"notes":    c.Notes,
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
