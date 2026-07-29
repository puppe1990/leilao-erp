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
	return map[string]any{
		"id":          p.ID,
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

	_ = h.inertia.Render(w, r, "Shop/Index", inertia.Props{
		"site":         meta.ForRequest(h.site, r),
		"companyName":  company,
		"products":     rows,
		"whatsappSet":  domain.NormalizeWhatsAppPhone(phone) != "",
		"whatsappHint": phone,
		"whatsappDigits": domain.NormalizeWhatsAppPhone(phone),
	})
}

// Show is a public product detail with WhatsApp CTA.
func (h *ShopHandler) Show(w http.ResponseWriter, r *http.Request, id int64) {
	p, err := h.store.FindProduct(id)
	if err != nil {
		http.Error(w, "produto não encontrado", http.StatusNotFound)
		return
	}
	if p.PhotoCount < 1 || p.QtyInStock < 1 {
		http.Error(w, "produto indisponível na vitrine", http.StatusNotFound)
		return
	}
	media, _ := h.store.ListProductMedia(id)
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
		http.Error(w, "produto indisponível na vitrine", http.StatusNotFound)
		return
	}

	phone, _ := h.store.WhatsAppPhone()
	row := shopProductRow(p, phone)
	row["photos"] = photos
	row["videos"] = videos
	row["listingText"] = p.ListingText
	row["screenType"] = p.ScreenType
	row["maxResolution"] = p.MaxResolution
	row["refreshRate"] = p.RefreshRate

	_ = h.inertia.Render(w, r, "Shop/Show", inertia.Props{
		"site":           meta.ForRequest(h.site, r),
		"companyName":    companyName(h.store),
		"product":        row,
		"whatsappSet":    domain.NormalizeWhatsAppPhone(phone) != "",
		"whatsappDigits": domain.NormalizeWhatsAppPhone(phone),
	})
}
