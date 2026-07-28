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
		        COALESCE((SELECT COUNT(*) FROM items i WHERE i.product_id = p.id AND i.status = 'in_stock'), 0)
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
		if err := rows.Scan(&p.ID, &p.Name, &hint, &p.Kind, &p.CreatedAt, &p.QtyInStock); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		if hint.Valid {
			v := hint.Int64
			p.SalePriceHintCents = &v
		}
		out = append(out, p)
	}
	return out, rows.Err()
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
