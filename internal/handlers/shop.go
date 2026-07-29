package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/meta"
	inertia "github.com/romsar/gonertia/v3"

	"github.com/puppe1990/leilao-erp/internal/domain"
	"github.com/puppe1990/leilao-erp/internal/models"
	"github.com/puppe1990/leilao-erp/internal/store"
)

func shopBrandFromName(name string) string {
	n := strings.ToLower(name)
	switch {
	case strings.Contains(n, "dell"):
		return "DELL"
	case strings.Contains(n, "samsung"):
		return "SAMSUNG"
	case strings.Contains(n, "lg") || strings.Contains(n, "carcaça lg"):
		return "LG"
	case strings.Contains(n, "lenovo") || strings.Contains(n, "thinkvision"):
		return "LENOVO"
	case strings.Contains(n, "philips"):
		return "PHILIPS"
	case strings.Contains(n, "prizi"):
		return "PRIZI"
	case strings.Contains(n, "hp"):
		return "HP"
	default:
		return "MONITOR"
	}
}

func shopBadge(p models.Product) string {
	n := strings.ToLower(p.Name + " " + p.ItemCondition)
	switch {
	case strings.Contains(n, "defeito") || strings.Contains(n, "amarelada") || strings.Contains(n, "piscando") || strings.Contains(n, "linha vertical"):
		return "DEFEITO"
	case strings.Contains(n, "sem base"):
		return "SEM BASE"
	case p.QtyInStock > 0:
		return "PRONTA ENTREGA"
	default:
		return ""
	}
}

func shopCategory(p models.Product) string {
	n := strings.ToLower(p.Name + " " + p.ItemCondition)
	if strings.Contains(n, "defeito") || strings.Contains(n, "amarelada") || strings.Contains(n, "piscando") || strings.Contains(n, "linha vertical") || strings.Contains(n, "para peças") {
		return "defeito"
	}
	if strings.Contains(n, "sem base") {
		return "sem-base"
	}
	return "com-base"
}

func shopProductRow(p models.Product, phone string) map[string]any {
	var cents int64
	price := "Consulte"
	priceRaw := ""
	pixPrice := ""
	if p.SalePriceHintCents != nil && *p.SalePriceHintCents > 0 {
		cents = *p.SalePriceHintCents
		price = domain.FormatBRL(cents)
		priceRaw = formatCashInput(cents)
		// 10% PIX discount (same as eletro-puppe-draft)
		pix := cents * 90 / 100
		pixPrice = domain.FormatBRL(pix)
	}
	wa := domain.WhatsAppOrderURL(phone, p.Name, price)
	if pixPrice != "" {
		// Prefer PIX price in WhatsApp message
		wa = domain.WhatsAppOrderURL(phone, p.Name, pixPrice+" no PIX")
	}
	cond := p.ItemCondition
	if cond == "" {
		cond = "Usado"
	}
	slug := p.Slug
	if slug == "" {
		slug = domain.ProductSlug(p.Name)
	}
	return map[string]any{
		"id":          p.ID,
		"slug":        slug,
		"href":        fmt.Sprintf("/produto/%s", slug),
		"name":        p.Name,
		"brand":       shopBrandFromName(p.Name),
		"badge":       shopBadge(p),
		"category":    shopCategory(p),
		"price":       price,
		"priceRaw":    priceRaw,
		"priceCents":  cents,
		"pixPrice":    pixPrice,
		"thumbUrl":    p.FirstPhotoURL,
		"qtyInStock":  p.QtyInStock,
		"photoCount":  p.PhotoCount,
		"videoCount":  p.VideoCount,
		"condition":   cond,
		"description": p.Description,
		"whatsappUrl": wa,
	}
}

// ShopHandler is the public mini storefront (products with photos, WhatsApp order).
type ShopHandler struct {
	store   store.Store
	site    meta.Site
	cfg     cais.Config
	inertia *inertia.Inertia
}

func NewShopHandler(s store.Store, site meta.Site, cfg cais.Config, i *inertia.Inertia) *ShopHandler {
	return &ShopHandler{store: s, site: site, cfg: cfg, inertia: i}
}

func (h *ShopHandler) baseURL(r *http.Request) string {
	if u := strings.TrimSpace(h.site.AppURL); u != "" {
		return strings.TrimRight(u, "/")
	}
	if u := strings.TrimSpace(h.cfg.AppURL); u != "" {
		return strings.TrimRight(u, "/")
	}
	if r == nil {
		return ""
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if xf := r.Header.Get("X-Forwarded-Proto"); xf != "" {
		scheme = strings.TrimSpace(strings.Split(xf, ",")[0])
	}
	host := r.Host
	if host == "" {
		return ""
	}
	return scheme + "://" + host
}

func (h *ShopHandler) shopOGImage(r *http.Request) string {
	return domain.AbsoluteURL(h.baseURL(r), "/static/og-shop.png")
}

// withShopOG injects Open Graph tags into the root HTML for link previews (WhatsApp, etc.).
func (h *ShopHandler) withShopOG(r *http.Request, m domain.OGMeta) *http.Request {
	if m.SiteName == "" {
		m.SiteName = companyName(h.store)
	}
	if m.ImageURL == "" {
		m.ImageURL = h.shopOGImage(r)
	}
	if m.Type == "" {
		m.Type = "website"
	}
	head := domain.OGHeadHTML(m)
	ctx := inertia.SetTemplateData(r.Context(), inertia.TemplateData{
		"ogHead":    template.HTML(head),
		"pageTitle": m.Title,
	})
	return r.WithContext(ctx)
}

func (h *ShopHandler) Index(w http.ResponseWriter, r *http.Request) {
	products, err := h.store.ListProductsWithPhotos()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	phone, err := h.store.WhatsAppPhone()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	company := companyName(h.store)

	rows := make([]map[string]any, 0, len(products))
	for _, p := range products {
		rows = append(rows, shopProductRow(p, phone))
	}

	title := company + " — Catálogo de monitores"
	desc := "Monitores usados testados. Monte o carrinho e peça no WhatsApp. 10% OFF no PIX."
	base := h.baseURL(r)
	pageURL := domain.AbsoluteURL(base, "/")
	imgURL := h.shopOGImage(r)
	r = h.withShopOG(r, domain.OGMeta{
		Title:       title,
		Description: desc,
		PageURL:     pageURL,
		ImageURL:    imgURL,
		SiteName:    company,
		Type:        "website",
	})

	_ = h.inertia.Render(w, r, "Shop/Index", inertia.Props{
		"site":           meta.ForRequest(h.site, r),
		"companyName":    company,
		"products":       rows,
		"whatsappSet":    domain.NormalizeWhatsAppPhone(phone) != "",
		"whatsappHint":   phone,
		"whatsappDigits": domain.NormalizeWhatsAppPhone(phone),
		"og": map[string]any{
			"title":       title,
			"description": desc,
			"image":       imgURL,
			"url":         pageURL,
		},
	})
}

// Show is a public product detail with WhatsApp CTA (URL uses slug).
func (h *ShopHandler) Show(w http.ResponseWriter, r *http.Request, slug string) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		http.Error(w, "produto não encontrado", http.StatusNotFound)
		return
	}

	// Legacy /produto/123 → /produto/slug
	if id, err := strconv.ParseInt(slug, 10, 64); err == nil && id > 0 {
		if p, err := h.store.FindProduct(id); err == nil {
			if p.Slug != "" {
				http.Redirect(w, r, "/produto/"+p.Slug, http.StatusMovedPermanently)
				return
			}
		}
	}

	p, err := h.store.FindProductBySlug(slug)
	if err != nil {
		http.Error(w, "produto não encontrado", http.StatusNotFound)
		return
	}
	if !p.ShopVisible || p.PhotoCount < 1 || p.QtyInStock < 1 {
		http.Error(w, "produto indisponível no catálogo", http.StatusNotFound)
		return
	}
	media, _ := h.store.ListProductMedia(p.ID)
	photos := make([]map[string]any, 0)
	videos := make([]map[string]any, 0)
	for _, m := range media {
		row := map[string]any{"id": m.ID, "url": m.URL, "kind": m.Kind}
		switch m.Kind {
		case "photo":
			photos = append(photos, row)
		case "video":
			videos = append(videos, row)
		}
	}
	if len(photos) == 0 {
		http.Error(w, "produto indisponível no catálogo", http.StatusNotFound)
		return
	}

	phone, _ := h.store.WhatsAppPhone()
	company := companyName(h.store)
	row := shopProductRow(p, phone)
	row["photos"] = photos
	row["videos"] = videos
	row["listingText"] = p.ListingText
	row["screenType"] = p.ScreenType
	row["maxResolution"] = p.MaxResolution
	row["refreshRate"] = p.RefreshRate

	title := p.Name + " — " + company
	desc := strings.TrimSpace(p.Description)
	if desc == "" {
		desc = strings.TrimSpace(p.ListingText)
	}
	if desc == "" {
		price := "Consulte"
		if p.SalePriceHintCents != nil && *p.SalePriceHintCents > 0 {
			price = domain.FormatBRL(*p.SalePriceHintCents*90/100) + " no PIX"
		}
		desc = fmt.Sprintf("%s · %s · monitores testados · pedido no WhatsApp", p.Name, price)
	}
	if len(desc) > 200 {
		desc = desc[:197] + "…"
	}
	base := h.baseURL(r)
	img := h.shopOGImage(r)
	if p.FirstPhotoURL != "" {
		img = domain.AbsoluteURL(base, p.FirstPhotoURL)
	} else if len(photos) > 0 {
		if u, ok := photos[0]["url"].(string); ok && u != "" {
			img = domain.AbsoluteURL(base, u)
		}
	}
	pageURL := domain.AbsoluteURL(base, "/produto/"+p.Slug)
	if p.Slug == "" {
		pageURL = domain.AbsoluteURL(base, fmt.Sprintf("/produto/%d", p.ID))
	}

	r = h.withShopOG(r, domain.OGMeta{
		Title:       title,
		Description: desc,
		ImageURL:    img,
		PageURL:     pageURL,
		SiteName:    company,
		Type:        "product",
	})

	_ = h.inertia.Render(w, r, "Shop/Show", inertia.Props{
		"site":           meta.ForRequest(h.site, r),
		"companyName":    company,
		"product":        row,
		"whatsappSet":    domain.NormalizeWhatsAppPhone(phone) != "",
		"whatsappDigits": domain.NormalizeWhatsAppPhone(phone),
		"og": map[string]any{
			"title":       title,
			"description": desc,
			"image":       img,
			"url":         pageURL,
		},
	})
}
