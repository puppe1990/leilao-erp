package handlers

import (
	"net/http"
	"strings"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/meta"
	inertia "github.com/romsar/gonertia/v3"

	"github.com/puppe1990/leilao-erp/internal/domain"
	"github.com/puppe1990/leilao-erp/internal/models"
	"github.com/puppe1990/leilao-erp/internal/store"
)

type FindsHandler struct {
	renderer *cais.Renderer
	store    store.Store
	site     meta.Site
	cfg      cais.Config
	inertia  *inertia.Inertia
}

func NewFindsHandler(renderer *cais.Renderer, s store.Store, site meta.Site, cfg cais.Config, i *inertia.Inertia) *FindsHandler {
	return &FindsHandler{renderer: renderer, store: s, site: site, cfg: cfg, inertia: i}
}

func findSourceLabel(source string) string {
	switch source {
	case store.FindSourceOLX:
		return "OLX"
	case store.FindSourceFacebook:
		return "Facebook"
	case store.FindSourceKwara:
		return "Kwara"
	default:
		return source
	}
}

func findRow(f models.Find) map[string]any {
	margin := f.SellCents - f.BuyCents
	return map[string]any{
		"id":          f.ID,
		"name":        f.Name,
		"url":         f.URL,
		"source":      f.Source,
		"sourceLabel": findSourceLabel(f.Source),
		"buy":         domain.FormatBRL(f.BuyCents),
		"buyRaw":      formatCashInput(f.BuyCents),
		"sell":        domain.FormatBRL(f.SellCents),
		"sellRaw":     formatCashInput(f.SellCents),
		"margin":      domain.FormatBRL(margin),
	}
}

func findInputFromRequest(r *http.Request) (store.FindInput, inertia.ValidationErrors) {
	errs := inertia.ValidationErrors{}
	in := store.FindInput{
		Name:   r.FormValue("name"),
		URL:    r.FormValue("url"),
		Source: r.FormValue("source"),
	}
	buy, err := domain.ParseBRLToCents(r.FormValue("buy"))
	if err != nil {
		errs["buy"] = "Valor de compra inválido"
	} else {
		in.BuyCents = buy
	}
	sell, err := domain.ParseBRLToCents(r.FormValue("sell"))
	if err != nil {
		errs["sell"] = "Valor de venda inválido"
	} else {
		in.SellCents = sell
	}
	return in, errs
}

func findSourceOptions() []map[string]string {
	return []map[string]string{
		{"value": store.FindSourceOLX, "label": "OLX"},
		{"value": store.FindSourceFacebook, "label": "Facebook"},
		{"value": store.FindSourceKwara, "label": "Kwara"},
	}
}

// Index lists marketplace finds, optionally filtered by ?source=.
func (h *FindsHandler) Index(w http.ResponseWriter, r *http.Request) {
	source := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("source")))
	list, err := h.store.ListFinds(source)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rows := make([]map[string]any, 0, len(list))
	for _, f := range list {
		rows = append(rows, findRow(f))
	}
	_ = h.inertia.Render(w, r, "Finds/Index", withCompany(h.store, inertia.Props{
		"site":    meta.ForRequest(h.site, r),
		"finds":   rows,
		"source":  source,
		"sources": findSourceOptions(),
	}))
}

// Create inserts a find.
func (h *FindsHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := parseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	in, errs := findInputFromRequest(r)
	if len(errs) > 0 {
		ctx := inertia.SetValidationErrors(r.Context(), errs)
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}
	if _, err := h.store.CreateFind(in); err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}
	h.inertia.Redirect(w, r, "/finds", http.StatusSeeOther)
}

// Update updates a find.
func (h *FindsHandler) Update(w http.ResponseWriter, r *http.Request, id int64) {
	if err := parseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	in, errs := findInputFromRequest(r)
	if len(errs) > 0 {
		ctx := inertia.SetValidationErrors(r.Context(), errs)
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}
	if err := h.store.UpdateFind(id, in); err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}
	h.inertia.Redirect(w, r, "/finds", http.StatusSeeOther)
}

// Destroy deletes a find.
func (h *FindsHandler) Destroy(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.store.DeleteFind(id); err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}
	h.inertia.Redirect(w, r, "/finds", http.StatusSeeOther)
}
