package store

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/puppe1990/leilao-erp/internal/models"
)

func productKindFromTitle(title string) string {
	t := strings.ToLower(strings.TrimSpace(title))
	if strings.HasPrefix(t, "cabo") || strings.Contains(t, " cabo") {
		return "accessory"
	}
	return "principal"
}

// EnsureProductByName returns product id for name, creating the row if needed.
func (s *SQLiteStore) EnsureProductByName(name, kind string, saleHint *int64) (int64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, fmt.Errorf("%w: product name required", ErrInvalidInput)
	}
	if kind != "accessory" {
		kind = "principal"
	}
	var id int64
	err := s.db.QueryRow(`SELECT id FROM products WHERE name = ?`, name).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("find product: %w", err)
	}
	var hint any
	if saleHint != nil {
		hint = *saleHint
	}
	res, err := s.db.Exec(
		`INSERT INTO products (name, sale_price_hint_cents, kind) VALUES (?, ?, ?)`,
		name, hint, kind,
	)
	if err != nil {
		return 0, fmt.Errorf("insert product: %w", err)
	}
	return res.LastInsertId()
}

// ListProducts returns catalog products ordered by name.
func (s *SQLiteStore) ListProducts() ([]models.Product, error) {
	rows, err := s.db.Query(
		`SELECT p.id, p.name, p.sale_price_hint_cents, p.kind, p.created_at,
		        COALESCE(p.description, ''), COALESCE(p.listing_text, ''),
		        COALESCE((SELECT COUNT(*) FROM items i WHERE i.product_id = p.id AND i.status = 'in_stock'), 0),
		        COALESCE((SELECT COUNT(*) FROM product_media m WHERE m.product_id = p.id AND m.kind = 'photo'), 0),
		        COALESCE((SELECT COUNT(*) FROM product_media m WHERE m.product_id = p.id AND m.kind = 'video'), 0),
		        COALESCE((
		          SELECT m.url FROM product_media m
		          WHERE m.product_id = p.id AND m.kind = 'photo'
		          ORDER BY m.sort_order, m.id LIMIT 1
		        ), ''),
		        COALESCE(p.olx_free_shipping, 0)
		 FROM products p
		 ORDER BY p.name`,
	)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.Product
	for rows.Next() {
		var p models.Product
		var hint sql.NullInt64
		var freeShip int
		if err := rows.Scan(
			&p.ID, &p.Name, &hint, &p.Kind, &p.CreatedAt,
			&p.Description, &p.ListingText,
			&p.QtyInStock, &p.PhotoCount, &p.VideoCount,
			&p.FirstPhotoURL,
			&freeShip,
		); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		if hint.Valid {
			v := hint.Int64
			p.SalePriceHintCents = &v
		}
		p.OlxFreeShipping = freeShip != 0
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListProductsWithPhotos returns in-stock catalog products that have at least one photo.
// Used by the public mini shop.
func (s *SQLiteStore) ListProductsWithPhotos() ([]models.Product, error) {
	rows, err := s.db.Query(
		`SELECT p.id, p.name, p.sale_price_hint_cents, p.kind, p.created_at,
		        COALESCE(p.description, ''), COALESCE(p.listing_text, ''),
		        COALESCE(p.item_condition, ''),
		        COALESCE((SELECT COUNT(*) FROM items i WHERE i.product_id = p.id AND i.status = 'in_stock'), 0),
		        COALESCE((SELECT COUNT(*) FROM product_media m WHERE m.product_id = p.id AND m.kind = 'photo'), 0),
		        COALESCE((SELECT COUNT(*) FROM product_media m WHERE m.product_id = p.id AND m.kind = 'video'), 0),
		        COALESCE((
		          SELECT m.url FROM product_media m
		          WHERE m.product_id = p.id AND m.kind = 'photo'
		          ORDER BY m.sort_order, m.id LIMIT 1
		        ), ''),
		        COALESCE(p.olx_free_shipping, 0)
		 FROM products p
		 WHERE EXISTS (
		   SELECT 1 FROM product_media m WHERE m.product_id = p.id AND m.kind = 'photo'
		 )
		 AND EXISTS (
		   SELECT 1 FROM items i WHERE i.product_id = p.id AND i.status = 'in_stock'
		 )
		 ORDER BY p.name`,
	)
	if err != nil {
		return nil, fmt.Errorf("list products with photos: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.Product
	for rows.Next() {
		var p models.Product
		var hint sql.NullInt64
		var freeShip int
		if err := rows.Scan(
			&p.ID, &p.Name, &hint, &p.Kind, &p.CreatedAt,
			&p.Description, &p.ListingText, &p.ItemCondition,
			&p.QtyInStock, &p.PhotoCount, &p.VideoCount,
			&p.FirstPhotoURL,
			&freeShip,
		); err != nil {
			return nil, fmt.Errorf("scan shop product: %w", err)
		}
		if hint.Valid {
			v := hint.Int64
			p.SalePriceHintCents = &v
		}
		p.OlxFreeShipping = freeShip != 0
		out = append(out, p)
	}
	return out, rows.Err()
}

// FindProduct returns one catalog product with stock/media counts.
func (s *SQLiteStore) FindProduct(id int64) (models.Product, error) {
	var p models.Product
	var hint sql.NullInt64
	var curved, box, dp, hdr, wide, cables, audio, hdmi, ultra, freeShip int
	err := s.db.QueryRow(
		`SELECT p.id, p.name, p.sale_price_hint_cents, p.kind, p.created_at,
		        COALESCE(p.description, ''), COALESCE(p.listing_text, ''),
		        COALESCE((SELECT COUNT(*) FROM items i WHERE i.product_id = p.id AND i.status = 'in_stock'), 0),
		        COALESCE((SELECT COUNT(*) FROM product_media m WHERE m.product_id = p.id AND m.kind = 'photo'), 0),
		        COALESCE((SELECT COUNT(*) FROM product_media m WHERE m.product_id = p.id AND m.kind = 'video'), 0),
		        COALESCE(p.screen_type, ''), COALESCE(p.max_resolution, ''),
		        COALESCE(p.refresh_rate, ''), COALESCE(p.item_condition, ''),
		        COALESCE(p.feat_curved, 0), COALESCE(p.feat_includes_box, 0),
		        COALESCE(p.feat_displayport, 0), COALESCE(p.feat_hdr, 0),
		        COALESCE(p.feat_widescreen, 0), COALESCE(p.feat_includes_cables, 0),
		        COALESCE(p.feat_audio, 0), COALESCE(p.feat_hdmi, 0), COALESCE(p.feat_ultrawide, 0),
		        COALESCE(p.olx_free_shipping, 0)
		 FROM products p WHERE p.id = ?`,
		id,
	).Scan(
		&p.ID, &p.Name, &hint, &p.Kind, &p.CreatedAt,
		&p.Description, &p.ListingText,
		&p.QtyInStock, &p.PhotoCount, &p.VideoCount,
		&p.ScreenType, &p.MaxResolution, &p.RefreshRate, &p.ItemCondition,
		&curved, &box, &dp, &hdr, &wide, &cables, &audio, &hdmi, &ultra,
		&freeShip,
	)
	if err == sql.ErrNoRows {
		return models.Product{}, ErrNotFound
	}
	if err != nil {
		return models.Product{}, fmt.Errorf("find product: %w", err)
	}
	if hint.Valid {
		v := hint.Int64
		p.SalePriceHintCents = &v
	}
	p.FeatCurved = curved != 0
	p.FeatIncludesBox = box != 0
	p.FeatDisplayPort = dp != 0
	p.FeatHDR = hdr != 0
	p.FeatWidescreen = wide != 0
	p.FeatIncludesCables = cables != 0
	p.FeatAudio = audio != 0
	p.FeatHDMI = hdmi != 0
	p.FeatUltrawide = ultra != 0
	p.OlxFreeShipping = freeShip != 0
	return p, nil
}

// UpdateProductDescriptions sets technical description and marketplace listing text.
func (s *SQLiteStore) UpdateProductDescriptions(productID int64, description, listingText string) error {
	res, err := s.db.Exec(
		`UPDATE products SET description = ?, listing_text = ? WHERE id = ?`,
		strings.TrimSpace(description), strings.TrimSpace(listingText), productID,
	)
	if err != nil {
		return fmt.Errorf("update product descriptions: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ProductOLXAttrs are structured marketplace fields (OLX-style monitor form).
type ProductOLXAttrs struct {
	ScreenType         string
	MaxResolution      string
	RefreshRate        string
	ItemCondition      string
	FeatCurved         bool
	FeatIncludesBox    bool
	FeatDisplayPort    bool
	FeatHDR            bool
	FeatWidescreen     bool
	FeatIncludesCables bool
	FeatAudio          bool
	FeatHDMI           bool
	FeatUltrawide      bool
	OlxFreeShipping    bool
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

// UpdateProductOLXAttrs saves screen/resolution/condition and feature flags.
func (s *SQLiteStore) UpdateProductOLXAttrs(productID int64, in ProductOLXAttrs) error {
	res, err := s.db.Exec(
		`UPDATE products SET
		   screen_type = ?, max_resolution = ?, refresh_rate = ?, item_condition = ?,
		   feat_curved = ?, feat_includes_box = ?, feat_displayport = ?, feat_hdr = ?,
		   feat_widescreen = ?, feat_includes_cables = ?, feat_audio = ?, feat_hdmi = ?,
		   feat_ultrawide = ?, olx_free_shipping = ?
		 WHERE id = ?`,
		strings.TrimSpace(in.ScreenType),
		strings.TrimSpace(in.MaxResolution),
		strings.TrimSpace(in.RefreshRate),
		strings.TrimSpace(in.ItemCondition),
		boolToInt(in.FeatCurved),
		boolToInt(in.FeatIncludesBox),
		boolToInt(in.FeatDisplayPort),
		boolToInt(in.FeatHDR),
		boolToInt(in.FeatWidescreen),
		boolToInt(in.FeatIncludesCables),
		boolToInt(in.FeatAudio),
		boolToInt(in.FeatHDMI),
		boolToInt(in.FeatUltrawide),
		boolToInt(in.OlxFreeShipping),
		productID,
	)
	if err != nil {
		return fmt.Errorf("update product olx attrs: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ListStockProductGroups aggregates in-stock units by product (or title fallback).
func (s *SQLiteStore) ListStockProductGroups() ([]models.Product, error) {
	rows, err := s.db.Query(
		`SELECT
		    COALESCE(p.id, 0),
		    COALESCE(p.name, i.title),
		    MAX(COALESCE(p.sale_price_hint_cents, i.sale_price_hint_cents)),
		    COALESCE(MAX(p.kind), 'principal'),
		    COUNT(*),
		    CAST(ROUND(AVG(i.unit_cost_cents)) AS INTEGER),
		    MIN(i.id),
		    MIN(i.lot_id)
		 FROM items i
		 LEFT JOIN products p ON p.id = i.product_id
		 WHERE i.status = 'in_stock'
		 GROUP BY COALESCE(p.id, 0), COALESCE(p.name, i.title)
		 ORDER BY COALESCE(p.name, i.title)`,
	)
	if err != nil {
		// Fallback: group by title only (no products table yet)
		return s.listStockGroupsByTitle()
	}
	defer func() { _ = rows.Close() }()

	var out []models.Product
	for rows.Next() {
		var p models.Product
		var hint sql.NullInt64
		if err := rows.Scan(
			&p.ID, &p.Name, &hint, &p.Kind,
			&p.QtyInStock, &p.UnitCostCents,
			&p.SampleItemID, &p.SampleLotID,
		); err != nil {
			return nil, fmt.Errorf("scan stock group: %w", err)
		}
		if hint.Valid {
			v := hint.Int64
			p.SalePriceHintCents = &v
		}
		if p.Kind == "" {
			p.Kind = productKindFromTitle(p.Name)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *SQLiteStore) listStockGroupsByTitle() ([]models.Product, error) {
	rows, err := s.db.Query(
		`SELECT title,
		        MAX(sale_price_hint_cents),
		        COUNT(*),
		        CAST(ROUND(AVG(unit_cost_cents)) AS INTEGER),
		        MIN(id),
		        MIN(lot_id)
		 FROM items WHERE status = 'in_stock'
		 GROUP BY title ORDER BY title`,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []models.Product
	for rows.Next() {
		var p models.Product
		var hint sql.NullInt64
		if err := rows.Scan(&p.Name, &hint, &p.QtyInStock, &p.UnitCostCents, &p.SampleItemID, &p.SampleLotID); err != nil {
			return nil, err
		}
		if hint.Valid {
			v := hint.Int64
			p.SalePriceHintCents = &v
		}
		p.Kind = productKindFromTitle(p.Name)
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListInStockUnitsByProduct returns unit rows for one product (or by title if productID=0).
func (s *SQLiteStore) ListInStockUnitsByProduct(productID int64, title string) ([]models.Item, error) {
	var rows *sql.Rows
	var err error
	if productID > 0 {
		rows, err = s.db.Query(
			`SELECT id, lot_id, product_id, sku, title, condition, unit_cost_cents, status,
			        sale_price_hint_cents, created_at
			 FROM items WHERE status = 'in_stock' AND product_id = ? ORDER BY id`,
			productID,
		)
	} else {
		rows, err = s.db.Query(
			`SELECT id, lot_id, product_id, sku, title, condition, unit_cost_cents, status,
			        sale_price_hint_cents, created_at
			 FROM items WHERE status = 'in_stock' AND title = ? ORDER BY id`,
			title,
		)
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanItems(rows)
}

func scanItems(rows *sql.Rows) ([]models.Item, error) {
	var out []models.Item
	for rows.Next() {
		var it models.Item
		var sku, condition sql.NullString
		var hint, productID sql.NullInt64
		if err := rows.Scan(
			&it.ID, &it.LotID, &productID, &sku, &it.Title, &condition,
			&it.UnitCostCents, &it.Status, &hint, &it.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		if productID.Valid {
			v := productID.Int64
			it.ProductID = &v
		}
		if sku.Valid {
			it.SKU = &sku.String
		}
		if condition.Valid {
			it.Condition = &condition.String
		}
		if hint.Valid {
			v := hint.Int64
			it.SalePriceHintCents = &v
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// UpdateProductSaleHint sets catalog price and syncs all in_stock units of that product.
func (s *SQLiteStore) UpdateProductSaleHint(productID int64, hintCents *int64) error {
	var hint any
	if hintCents != nil {
		hint = *hintCents
	}
	res, err := s.db.Exec(`UPDATE products SET sale_price_hint_cents = ? WHERE id = ?`, hint, productID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	_, err = s.db.Exec(
		`UPDATE items SET sale_price_hint_cents = ? WHERE product_id = ? AND status IN ('in_stock','reserved')`,
		hint, productID,
	)
	return err
}

// RenameProduct renames catalog and all linked items.
func (s *SQLiteStore) RenameProduct(productID int64, newName string) error {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return fmt.Errorf("%w: name required", ErrInvalidInput)
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.Exec(`UPDATE products SET name = ? WHERE id = ?`, newName, productID)
	if err != nil {
		return fmt.Errorf("rename product: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	if _, err := tx.Exec(`UPDATE items SET title = ? WHERE product_id = ?`, newName, productID); err != nil {
		return err
	}
	return tx.Commit()
}
