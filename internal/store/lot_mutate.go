package store

import (
	"fmt"
	"strings"
)

// --- Lots ---

type UpdateLotInput struct {
	Name          string
	PurchasedAt   string
	AuctionSource string
	Notes         string
}

func (s *SQLiteStore) UpdateLot(id int64, in UpdateLotInput) error {
	name := strings.TrimSpace(in.Name)
	purchasedAt := strings.TrimSpace(in.PurchasedAt)
	if name == "" || purchasedAt == "" {
		return fmt.Errorf("%w: name and purchased_at required", ErrInvalidInput)
	}
	var src, notes any
	if strings.TrimSpace(in.AuctionSource) != "" {
		src = strings.TrimSpace(in.AuctionSource)
	}
	if strings.TrimSpace(in.Notes) != "" {
		notes = strings.TrimSpace(in.Notes)
	}
	res, err := s.db.Exec(
		`UPDATE lots SET name = ?, purchased_at = ?, auction_source = ?, notes = ? WHERE id = ?`,
		name, purchasedAt, src, notes, id,
	)
	if err != nil {
		return fmt.Errorf("update lot: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteLot removes a lot only when no items are sold (and cascades related open data).
func (s *SQLiteStore) DeleteLot(id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var sold int
	if err := tx.QueryRow(
		`SELECT COUNT(*) FROM items WHERE lot_id = ? AND status = 'sold'`, id,
	).Scan(&sold); err != nil {
		return err
	}
	if sold > 0 {
		return fmt.Errorf("%w: lot has sold items", ErrCannotDelete)
	}

	// cash entries linked to this lot
	if _, err := tx.Exec(`DELETE FROM cash_entries WHERE lot_id = ?`, id); err != nil {
		return fmt.Errorf("delete lot cash entries: %w", err)
	}
	// purchase costs
	if _, err := tx.Exec(`DELETE FROM purchase_costs WHERE lot_id = ?`, id); err != nil {
		return fmt.Errorf("delete purchase costs: %w", err)
	}
	// payables for lot
	if _, err := tx.Exec(`DELETE FROM payables WHERE lot_id = ?`, id); err != nil {
		return fmt.Errorf("delete payables: %w", err)
	}
	// items
	if _, err := tx.Exec(`DELETE FROM items WHERE lot_id = ?`, id); err != nil {
		return fmt.Errorf("delete items: %w", err)
	}
	res, err := tx.Exec(`DELETE FROM lots WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete lot: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return tx.Commit()
}
