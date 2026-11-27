package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/puppe1990/cais/pkg/cais/meta"
	inertia "github.com/romsar/gonertia/v3"

	"github.com/puppe1990/leilao-erp/internal/domain"
	"github.com/puppe1990/leilao-erp/internal/store"
)

func shopLotOfferProps(offer *store.ShopLotOffer, phone string) map[string]any {
	if offer == nil || offer.Qty < 2 {
		return nil
	}
	price := domain.FormatBRL(offer.PriceCents)
	retail := ""
	if offer.RetailCents > offer.PriceCents {
		retail = domain.FormatBRL(offer.RetailCents)
	}
	items := make([]map[string]any, 0, len(offer.Items))
	thumbs := make([]string, 0, 6)
	for _, it := range offer.Items {
		slug := it.Slug
		if slug == "" {
			slug = domain.ProductSlug(it.Name)
		}
		thumb := domain.CatalogThumbURL(it.ThumbURL)
		items = append(items, map[string]any{
			"productId": it.ProductID,
			"name":      it.Name,
			"slug":      slug,
			"href":      fmt.Sprintf("/produto/%s", slug),
			"qty":       it.Qty,
			"price":     domain.FormatBRL(it.PriceCents),
			"thumbUrl":  thumb,
		})
		if thumb != "" && len(thumbs) < 6 {
			thumbs = append(thumbs, thumb)
		}
	}
	photos := make([]map[string]any, 0, len(offer.Photos))
	for _, m := range offer.Photos {
		photos = append(photos, map[string]any{
			"id": m.ID, "url": m.URL, "kind": "photo",
			"thumbUrl":    domain.CatalogThumbURL(m.URL),
			"productId":   m.ProductID,
			"productName": m.ProductName,
		})
	}
	videos := make([]map[string]any, 0, len(offer.Videos))
	for _, m := range offer.Videos {
		videos = append(videos, map[string]any{
			"id": m.ID, "url": m.URL, "kind": "video",
			"productId":   m.ProductID,
			"productName": m.ProductName,
		})
	}
	name := fmt.Sprintf("Lote de %d monitores", offer.Qty)
	href := "/produto/" + store.ShopLotSlug
	var imageURL, thumbURL string
	if len(offer.Photos) > 0 {
		imageURL = offer.Photos[0].URL
		thumbURL = domain.CatalogThumbURL(imageURL)
	}
	return map[string]any{
		"id":          "lot",
		"slug":        store.ShopLotSlug,
		"href":        href,
		"brand":       "LOTE",
		"category":    "lote",
		"name":        name,
		"qty":         offer.Qty,
		"price":       price,
		"priceCents":  offer.PriceCents,
		"pixPrice":    price,
		"retailPrice": retail,
		"thumbUrl":    thumbURL,
		"imageUrl":    imageURL,
		"thumbs":      thumbs,
		"photoCount":  len(photos),
		"videoCount":  len(videos),
		"photos":      photos,
		"videos":      videos,
		"items":       items,
		"whatsappUrl": domain.WhatsAppLotURL(phone, offer.Qty, price),
		"fixedPix":    true,
		"isLot":       true,
		"qtyInStock":  1,
		"badge":       "LOTE",
		"condition":   "Usados · testados",
		"description": shopLotDescription(offer.Qty, price, retail),
	}
}

func shopLotDescription(qty int, price, retail string) string {
	s := fmt.Sprintf("Leve os %d monitores restantes do leilão por %s.", qty, price)
	if retail != "" {
		s += fmt.Sprintf(" No varejo do catálogo somam %s.", retail)
	}
	s += " Preço fechado do lote (sem 10% extra no PIX). Fotos e vídeos reais de cada unidade."
	return s
}

func isShopLotSlug(slug string) bool {
	s := strings.ToLower(strings.TrimSpace(slug))
	return s == store.ShopLotSlug || s == "lote" || s == "lote-de-monitores"
}

func (h *ShopHandler) loadLotOffer(phone string) (map[string]any, error) {
	offer, err := h.store.ShopLotOffer()
	if err != nil {
		return nil, err
	}
	return shopLotOfferProps(offer, phone), nil
}

// Lot is the public remaining-auction-lot deal with all photos and videos.
func (h *ShopHandler) Lot(w http.ResponseWriter, r *http.Request) {
	phone, err := h.store.WhatsAppPhone()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	row, err := h.loadLotOffer(phone)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if row == nil {
		http.Error(w, "lote indisponível no catálogo", http.StatusNotFound)
		return
	}
	company := companyName(h.store)
	showQty, _ := h.store.ShopShowQty()
	title := fmt.Sprintf("%s — %s", row["name"], company)
	desc, _ := row["description"].(string)
	base := h.baseURL(r)
	img := h.shopOGImage(r)
	if u, ok := row["imageUrl"].(string); ok && u != "" {
		img = domain.AbsoluteURL(base, u)
	}
	pageURL := domain.AbsoluteURL(base, "/produto/"+store.ShopLotSlug)
	r = h.withShopOG(r, domain.OGMeta{
		Title:       title,
		Description: desc,
		ImageURL:    img,
		PageURL:     pageURL,
		SiteName:    company,
		Type:        "product",
	})
	_ = h.inertia.Render(w, r, "Shop/Lot", inertia.Props{
		"site":           meta.ForRequest(h.site, r),
		"companyName":    company,
		"lotOffer":       row,
		"shopShowQty":    showQty,
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
