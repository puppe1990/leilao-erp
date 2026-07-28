package handlers

import (
	"net/http"
	"strings"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/meta"
	inertia "github.com/romsar/gonertia/v3"

	"github.com/puppe1990/leilao-erp/internal/domain"
	"github.com/puppe1990/leilao-erp/internal/store"
)

type ProductsHandler struct {
	renderer *cais.Renderer
	store    store.Store
	site     meta.Site
	cfg      cais.Config
	inertia  *inertia.Inertia
}

func NewProductsHandler(renderer *cais.Renderer, s store.Store, site meta.Site, cfg cais.Config, i *inertia.Inertia) *ProductsHandler {
	return &ProductsHandler{renderer: renderer, store: s, site: site, cfg: cfg, inertia: i}
}

// Index lists the product catalog (names reusable across stock units).
func (h *ProductsHandler) Index(w http.ResponseWriter, r *http.Request) {
	products, err := h.store.ListProducts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows := make([]map[string]any, 0, len(products))
	for _, p := range products {
		sale := "—"
		saleRaw := ""
		if p.SalePriceHintCents != nil {
			sale = domain.FormatBRL(*p.SalePriceHintCents)
			saleRaw = formatCentsInput(*p.SalePriceHintCents)
		}
		kindLabel := "Principal"
		if p.Kind == "accessory" {
			kindLabel = "Acessório"
		}
		rows = append(rows, map[string]any{
			"id":            p.ID,
			"name":          p.Name,
			"kind":          p.Kind,
			"kindLabel":     kindLabel,
			"qtyInStock":    p.QtyInStock,
			"salePriceHint": sale,
			"salePriceRaw":  saleRaw,
		})
	}

	_ = h.inertia.Render(w, r, "Products/Index", withCompany(h.store, inertia.Props{
		"site":     meta.ForRequest(h.site, r),
		"products": rows,
	}))
}

// Update renames a product and/or sets sale price (syncs all linked stock units).
func (h *ProductsHandler) Update(w http.ResponseWriter, r *http.Request, id int64) {
	if err := parseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		name = strings.TrimSpace(r.FormValue("title"))
	}
	if name != "" {
		if err := h.store.RenameProduct(id, name); err != nil {
			ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
			r = r.WithContext(ctx)
			h.Index(w, r)
			return
		}
	}

	raw := strings.TrimSpace(r.FormValue("sale_price_hint"))
	if raw != "" {
		cents, err := domain.ParseBRLToCents(raw)
		if err != nil || cents < 0 {
			ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": "Preço de venda inválido"})
			r = r.WithContext(ctx)
			h.Index(w, r)
			return
		}
		if err := h.store.UpdateProductSaleHint(id, &cents); err != nil {
			ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
			r = r.WithContext(ctx)
			h.Index(w, r)
			return
		}
	} else if r.FormValue("clear_sale_price") == "1" {
		if err := h.store.UpdateProductSaleHint(id, nil); err != nil {
			ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
			r = r.WithContext(ctx)
			h.Index(w, r)
			return
		}
	}

	ret := strings.TrimSpace(r.FormValue("return_to"))
	if ret == "/stock" {
		h.inertia.Redirect(w, r, "/stock", http.StatusSeeOther)
		return
	}
	if ret == "/products" || ret == "" {
		h.inertia.Redirect(w, r, "/products", http.StatusSeeOther)
		return
	}
	h.inertia.Redirect(w, r, "/products", http.StatusSeeOther)
}
