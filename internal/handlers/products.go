package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/meta"
	inertia "github.com/romsar/gonertia/v3"

	"github.com/puppe1990/leilao-erp/internal/domain"
	"github.com/puppe1990/leilao-erp/internal/models"
	"github.com/puppe1990/leilao-erp/internal/store"
)

type ProductsHandler struct {
	renderer  *cais.Renderer
	store     store.Store
	site      meta.Site
	cfg       cais.Config
	inertia   *inertia.Inertia
	staticDir string // web/static absolute path for product uploads
}

func NewProductsHandler(renderer *cais.Renderer, s store.Store, site meta.Site, cfg cais.Config, i *inertia.Inertia) *ProductsHandler {
	return &ProductsHandler{renderer: renderer, store: s, site: site, cfg: cfg, inertia: i}
}

// WithStaticDir sets the directory served at /static (for media uploads).
func (h *ProductsHandler) WithStaticDir(dir string) *ProductsHandler {
	h.staticDir = dir
	return h
}

func productKindLabel(kind string) string {
	if kind == "accessory" {
		return "Acessório"
	}
	return "Principal"
}

func productListRow(p models.Product) map[string]any {
	sale := "—"
	saleRaw := ""
	if p.SalePriceHintCents != nil {
		sale = domain.FormatBRL(*p.SalePriceHintCents)
		saleRaw = formatCentsInput(*p.SalePriceHintCents)
	}
	return map[string]any{
		"id":            p.ID,
		"name":          p.Name,
		"kind":          p.Kind,
		"kindLabel":     productKindLabel(p.Kind),
		"qtyInStock":    p.QtyInStock,
		"photoCount":    p.PhotoCount,
		"videoCount":    p.VideoCount,
		"salePriceHint": sale,
		"salePriceRaw":  saleRaw,
		"description":   p.Description,
		"listingText":   p.ListingText,
	}
}

// Index lists the product catalog (listing only — no inline edit/view panels).
func (h *ProductsHandler) Index(w http.ResponseWriter, r *http.Request) {
	products, err := h.store.ListProducts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows := make([]map[string]any, 0, len(products))
	for _, p := range products {
		rows = append(rows, productListRow(p))
	}

	_ = h.inertia.Render(w, r, "Products/Index", withCompany(h.store, inertia.Props{
		"site":     meta.ForRequest(h.site, r),
		"products": rows,
	}))
}

// Show renders a single product (view + media).
func (h *ProductsHandler) Show(w http.ResponseWriter, r *http.Request, id int64) {
	p, err := h.store.FindProduct(id)
	if err != nil {
		http.Error(w, "produto não encontrado", http.StatusNotFound)
		return
	}
	row := productListRow(p)
	media, _ := h.store.ListProductMedia(id)
	mediaRows := make([]map[string]any, 0, len(media))
	for _, m := range media {
		mediaRows = append(mediaRows, map[string]any{
			"id":   m.ID,
			"kind": m.Kind,
			"url":  m.URL,
		})
	}
	row["media"] = mediaRows

	_ = h.inertia.Render(w, r, "Products/Show", withCompany(h.store, inertia.Props{
		"site":    meta.ForRequest(h.site, r),
		"product": row,
	}))
}

// Edit renders the product edit form.
func (h *ProductsHandler) Edit(w http.ResponseWriter, r *http.Request, id int64) {
	p, err := h.store.FindProduct(id)
	if err != nil {
		http.Error(w, "produto não encontrado", http.StatusNotFound)
		return
	}
	_ = h.inertia.Render(w, r, "Products/Edit", withCompany(h.store, inertia.Props{
		"site":    meta.ForRequest(h.site, r),
		"product": productListRow(p),
	}))
}

// Update renames a product and/or sets sale price (syncs all linked stock units).
func (h *ProductsHandler) Update(w http.ResponseWriter, r *http.Request, id int64) {
	if err := parseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	redirectAfter := func() {
		ret := strings.TrimSpace(r.FormValue("return_to"))
		switch {
		case ret == "/stock":
			h.inertia.Redirect(w, r, "/stock", http.StatusSeeOther)
		case strings.HasPrefix(ret, "/products/"):
			h.inertia.Redirect(w, r, ret, http.StatusSeeOther)
		case ret == "/products":
			h.inertia.Redirect(w, r, "/products", http.StatusSeeOther)
		default:
			h.inertia.Redirect(w, r, fmt.Sprintf("/products/%d", id), http.StatusSeeOther)
		}
	}

	renderEditError := func(msg string) {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": msg})
		r = r.WithContext(ctx)
		h.Edit(w, r, id)
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		name = strings.TrimSpace(r.FormValue("title"))
	}
	if name != "" {
		if err := h.store.RenameProduct(id, name); err != nil {
			renderEditError(err.Error())
			return
		}
	}

	raw := strings.TrimSpace(r.FormValue("sale_price_hint"))
	if raw != "" {
		cents, err := domain.ParseBRLToCents(raw)
		if err != nil || cents < 0 {
			renderEditError("Preço de venda inválido")
			return
		}
		if err := h.store.UpdateProductSaleHint(id, &cents); err != nil {
			renderEditError(err.Error())
			return
		}
	} else if r.FormValue("clear_sale_price") == "1" {
		if err := h.store.UpdateProductSaleHint(id, nil); err != nil {
			renderEditError(err.Error())
			return
		}
	}

	if r.FormValue("description") != "" || r.FormValue("listing_text") != "" || r.FormValue("save_descriptions") == "1" {
		if err := h.store.UpdateProductDescriptions(id, r.FormValue("description"), r.FormValue("listing_text")); err != nil {
			renderEditError(err.Error())
			return
		}
	}

	redirectAfter()
}
