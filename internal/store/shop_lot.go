package store

import (
	"fmt"
	"strings"
)

// ShopLotPriceCents is the catalog price for the remaining auction-lot bundle (R$ 1.500).
const ShopLotPriceCents int64 = 150000

// ShopLotSlug is the public catalog URL segment (/produto/lote-monitores).
const ShopLotSlug = "lote-monitores"

// ShopLotOffer is the public catalog bundle for remaining leilão monitors.
type ShopLotOffer struct {
	Qty         int
	PriceCents  int64
	RetailCents int64
	Items       []ShopLotItem
	Photos      []ShopLotMedia
	Videos      []ShopLotMedia
}

// ShopLotItem is one product line inside the lot offer.
type ShopLotItem struct {
	ProductID  int64
	Name       string
	Slug       string
	Qty        int
	PriceCents int64
	ThumbURL   string
}

// ShopLotMedia is a photo or video belonging to a remaining lot unit.
type ShopLotMedia struct {
	ID          int64
	ProductID   int64
	ProductName string
	Kind        string
	URL         string
}

const shopLotLotFilter = `(l.name LIKE '%leil%' AND lower(l.name) NOT LIKE '%pessoal%')`

// ShopLotOffer returns the public lot deal for remaining leilão stock.
// Nil when fewer than 2 units remain (not a lot).
func (s *SQLiteStore) ShopLotOffer() (*ShopLotOffer, error) {
	items, err := s.shopLotItems()
	if err != nil {
		return nil, err
	}
	var qty int
	var retail int64
	for _, it := range items {
		qty += it.Qty
		retail += it.PriceCents * int64(it.Qty)
	}
	if qty < 2 {
		return nil, nil
	}
	photos, videos, err := s.shopLotMedia()
	if err != nil {
		return nil, err
	}
	return &ShopLotOffer{
		Qty:         qty,
		PriceCents:  ShopLotPriceCents,
		RetailCents: retail,
		Items:       items,
		Photos:      photos,
		Videos:      videos,
	}, nil
}

func (s *SQLiteStore) shopLotItems() ([]ShopLotItem, error) {
	rows, err := s.db.Query(`
		SELECT p.id, p.name, COALESCE(p.slug,''),
		       COUNT(i.id),
		       COALESCE(p.sale_price_hint_cents, 0),
		       COALESCE((
		         SELECT m.url FROM product_media m
		         WHERE m.product_id = p.id AND m.kind = 'photo'
		         ORDER BY m.sort_order, m.id LIMIT 1
		       ), '')
		FROM items i
		JOIN lots l ON l.id = i.lot_id
		JOIN products p ON p.id = i.product_id
		WHERE i.status = 'in_stock'
		  AND COALESCE(p.shop_visible, 0) = 1 -- hidden catalog units stay out of the lot
		  AND ` + shopLotLotFilter + `
		GROUP BY p.id
		ORDER BY p.name`)
	if err != nil {
		return nil, fmt.Errorf("shop lot items: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []ShopLotItem
	for rows.Next() {
		var it ShopLotItem
		if err := rows.Scan(&it.ProductID, &it.Name, &it.Slug, &it.Qty, &it.PriceCents, &it.ThumbURL); err != nil {
			return nil, fmt.Errorf("scan shop lot item: %w", err)
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (s *SQLiteStore) shopLotMedia() (photos, videos []ShopLotMedia, err error) {
	rows, err := s.db.Query(`
		SELECT m.id, m.product_id, p.name, m.kind, m.url
		FROM product_media m
		JOIN products p ON p.id = m.product_id
		WHERE COALESCE(p.shop_visible, 0) = 1
		  AND p.id IN (
		  SELECT DISTINCT i.product_id
		  FROM items i
		  JOIN lots l ON l.id = i.lot_id
		  WHERE i.status = 'in_stock'
		    AND i.product_id IS NOT NULL
		    AND ` + shopLotLotFilter + `
		)
		ORDER BY CASE m.kind WHEN 'photo' THEN 0 ELSE 1 END, p.name, m.sort_order, m.id`)
	if err != nil {
		return nil, nil, fmt.Errorf("shop lot media: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var m ShopLotMedia
		if err := rows.Scan(&m.ID, &m.ProductID, &m.ProductName, &m.Kind, &m.URL); err != nil {
			return nil, nil, fmt.Errorf("scan shop lot media: %w", err)
		}
		switch strings.ToLower(m.Kind) {
		case "video":
			videos = append(videos, m)
		default:
			photos = append(photos, m)
		}
	}
	return photos, videos, rows.Err()
}
