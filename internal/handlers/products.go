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
		"id":              p.ID,
		"name":            p.Name,
		"kind":            p.Kind,
		"kindLabel":       productKindLabel(p.Kind),
		"qtyInStock":      p.QtyInStock,
		"photoCount":      p.PhotoCount,
		"videoCount":      p.VideoCount,
		"thumbUrl":        p.FirstPhotoURL,
		"salePriceHint":   sale,
		"salePriceRaw":    saleRaw,
		"description":     p.Description,
		"listingText":     p.ListingText,
		"screenType":      p.ScreenType,
		"maxResolution":   p.MaxResolution,
		"refreshRate":     p.RefreshRate,
		"condition":       p.ItemCondition,
		"olxFreeShipping": p.OlxFreeShipping,
		"features": map[string]bool{
			"curved":         p.FeatCurved,
			"includesBox":    p.FeatIncludesBox,
			"displayPort":    p.FeatDisplayPort,
			"hdr":            p.FeatHDR,
			"widescreen":     p.FeatWidescreen,
			"includesCables": p.FeatIncludesCables,
			"audio":          p.FeatAudio,
			"hdmi":           p.FeatHDMI,
			"ultrawide":      p.FeatUltrawide,
		},
	}
}

// olxFormOptions are select/checkbox choices aligned with OLX monitor ads.
func olxFormOptions() map[string]any {
	opt := func(values ...string) []map[string]string {
		out := make([]map[string]string, 0, len(values)+1)
		out = append(out, map[string]string{"value": "", "label": "Selecione"})
		for _, v := range values {
			out = append(out, map[string]string{"value": v, "label": v})
		}
		return out
	}
	return map[string]any{
		"screenTypes": opt("LED", "LCD", "IPS", "VA", "TN", "OLED", "QLED", "Plasma"),
		"resolutions": opt(
			"1280x1024 (SXGA)",
			"1366x768 (HD)",
			"1440x900 (HD+)",
			"1600x900 (HD+)",
			"1680x1050 (WSXGA+)",
			"1920x1080 (Full HD)",
			"2560x1080 (Ultrawide FHD)",
			"2560x1440 (QHD)",
			"3440x1440 (Ultrawide QHD)",
			"3840x2160 (4K UHD)",
		),
		"refreshRates": opt("60 Hz", "75 Hz", "100 Hz", "120 Hz", "144 Hz", "165 Hz", "180 Hz", "200 Hz", "240 Hz"),
		"conditions": opt(
			"Novo",
			"Usado - Excelente",
			"Usado - Bom",
			"Usado - Aceitável",
			"Recondicionado",
			"Para peças / com defeito",
		),
		"features": []map[string]string{
			{"key": "curved", "label": "Curvo"},
			{"key": "includesBox", "label": "Inclui caixa"},
			{"key": "displayPort", "label": "Possui DisplayPort"},
			{"key": "hdr", "label": "Possui HDR"},
			{"key": "widescreen", "label": "Widescreen"},
			{"key": "includesCables", "label": "Inclui cabos"},
			{"key": "audio", "label": "Possui áudio"},
			{"key": "hdmi", "label": "Possui HDMI"},
			{"key": "ultrawide", "label": "Ultrawide"},
		},
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
		"site":       meta.ForRequest(h.site, r),
		"product":    productListRow(p),
		"olxOptions": olxFormOptions(),
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

	if r.FormValue("save_olx") == "1" {
		feat := func(key string) bool {
			v := strings.TrimSpace(r.FormValue("feat_" + key))
			return v == "1" || v == "true" || v == "on" || v == "yes"
		}
		freeShip := strings.TrimSpace(r.FormValue("olx_free_shipping"))
		if err := h.store.UpdateProductOLXAttrs(id, store.ProductOLXAttrs{
			ScreenType:         r.FormValue("screen_type"),
			MaxResolution:      r.FormValue("max_resolution"),
			RefreshRate:        r.FormValue("refresh_rate"),
			ItemCondition:      r.FormValue("condition"),
			FeatCurved:         feat("curved"),
			FeatIncludesBox:    feat("includesBox"),
			FeatDisplayPort:    feat("displayPort"),
			FeatHDR:            feat("hdr"),
			FeatWidescreen:     feat("widescreen"),
			FeatIncludesCables: feat("includesCables"),
			FeatAudio:          feat("audio"),
			FeatHDMI:           feat("hdmi"),
			FeatUltrawide:      feat("ultrawide"),
			OlxFreeShipping:    freeShip == "1" || freeShip == "true" || freeShip == "on" || freeShip == "yes",
		}); err != nil {
			renderEditError(err.Error())
			return
		}
	}

	redirectAfter()
}
