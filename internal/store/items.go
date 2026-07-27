package store

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/puppe1990/leilao-erp/internal/models"
)

type UpdateItemInput struct {
	Title              string
	SKU                string
	SalePriceHintCents *int64 // nil = clear hint
}

// UpdateItem updates mutable fields for an in-stock item.
func (s *SQLiteStore) UpdateItem(id int64, in UpdateItemInput) error {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return fmt.Errorf("%w: title required", ErrInvalidInput)
	}
	var status string
	err := s.db.QueryRow(`SELECT status FROM items WHERE id = ?`, id).Scan(&status)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if status != "in_stock" && status != "reserved" {
		return fmt.Errorf("%w: only in-stock items can be edited", ErrCannotUpdate)
	}
	var skuVal any
	if strings.TrimSpace(in.SKU) != "" {
		skuVal = strings.TrimSpace(in.SKU)
	}
	var hint any
	if in.SalePriceHintCents != nil {
		hint = *in.SalePriceHintCents
	}
	_, err = s.db.Exec(
		`UPDATE items SET title = ?, sku = ?, sale_price_hint_cents = ? WHERE id = ?`,
		title, skuVal, hint, id,
	)
	return err
}

// SetSalePriceHintByTitle sets sale_price_hint_cents on all matching in-stock items.
func (s *SQLiteStore) SetSalePriceHintByTitle(title string, hintCents int64) (int64, error) {
	res, err := s.db.Exec(
		`UPDATE items SET sale_price_hint_cents = ? WHERE title = ? AND status IN ('in_stock', 'reserved')`,
		hintCents, title,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// RenameItemsByTitle renames all items with fromTitle to toTitle (any status).
func (s *SQLiteStore) RenameItemsByTitle(fromTitle, toTitle string) (int64, error) {
	toTitle = strings.TrimSpace(toTitle)
	if toTitle == "" {
		return 0, fmt.Errorf("%w: title required", ErrInvalidInput)
	}
	res, err := s.db.Exec(`UPDATE items SET title = ? WHERE title = ?`, toTitle, fromTitle)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (s *SQLiteStore) FindItemByID(id int64) (models.Item, error) {
	var it models.Item
	var sku, condition sql.NullString
	var hint sql.NullInt64
	err := s.db.QueryRow(
		`SELECT id, lot_id, sku, title, condition, unit_cost_cents, status, sale_price_hint_cents, created_at
		 FROM items WHERE id = ?`, id,
	).Scan(&it.ID, &it.LotID, &sku, &it.Title, &condition, &it.UnitCostCents, &it.Status, &hint, &it.CreatedAt)
	if err == sql.ErrNoRows {
		return models.Item{}, ErrNotFound
	}
	if err != nil {
		return models.Item{}, err
	}
	if sku.Valid {
		v := sku.String
		it.SKU = &v
	}
	if condition.Valid {
		v := condition.String
		it.Condition = &v
	}
	if hint.Valid {
		v := hint.Int64
		it.SalePriceHintCents = &v
	}
	return it, nil
}

// ListItemsInStock returns items available for sale (status = in_stock).
func (s *SQLiteStore) ListItemsInStock() ([]models.Item, error) {
	rows, err := s.db.Query(
		`SELECT id, lot_id, sku, title, condition, unit_cost_cents, status,
		        sale_price_hint_cents, created_at
		 FROM items WHERE status = 'in_stock' ORDER BY id`,
	)
	if err != nil {
		return nil, fmt.Errorf("list items in stock: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.Item
	for rows.Next() {
		var it models.Item
		var sku, condition sql.NullString
		var hint sql.NullInt64
		if err := rows.Scan(
			&it.ID, &it.LotID, &sku, &it.Title, &condition,
			&it.UnitCostCents, &it.Status, &hint, &it.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
