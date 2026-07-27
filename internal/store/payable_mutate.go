package store

import (
	"database/sql"
	"fmt"
	"strings"
)

func (s *SQLiteStore) CancelPayable(id int64) error {
	var status string
	err := s.db.QueryRow(`SELECT status FROM payables WHERE id = ?`, id).Scan(&status)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if status != "open" {
		return fmt.Errorf("%w: only open payables can be cancelled", ErrCannotUpdate)
	}
	_, err = s.db.Exec(`UPDATE payables SET status = 'cancelled' WHERE id = ?`, id)
	return err
}

func (s *SQLiteStore) CancelReceivable(id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var status string
	var saleID sql.NullInt64
	err = tx.QueryRow(`SELECT status, sale_id FROM receivables WHERE id = ?`, id).Scan(&status, &saleID)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if status != "open" {
		return fmt.Errorf("%w: only open receivables can be cancelled", ErrCannotUpdate)
	}
	if _, err := tx.Exec(`UPDATE receivables SET status = 'cancelled' WHERE id = ?`, id); err != nil {
		return err
	}
	// if linked to pending sale, cancel sale + restore item via CancelPendingSale logic inline
	if saleID.Valid {
		// Delegate to CancelPendingSale for multi-line restore (after this tx commits
		// would nest poorly). Inline multi-line restore here instead.
		var pstatus string
		err = tx.QueryRow(`SELECT payment_status FROM sales WHERE id = ?`, saleID.Int64).Scan(&pstatus)
		if err == nil && pstatus == "pending" {
			if _, err := tx.Exec(`UPDATE sales SET payment_status = 'cancelled' WHERE id = ?`, saleID.Int64); err != nil {
				return err
			}
			rows, err := tx.Query(`SELECT item_id FROM sale_lines WHERE sale_id = ?`, saleID.Int64)
			if err != nil {
				return err
			}
			var itemIDs []int64
			for rows.Next() {
				var id int64
				if err := rows.Scan(&id); err != nil {
					_ = rows.Close()
					return err
				}
				itemIDs = append(itemIDs, id)
			}
			_ = rows.Close()
			if len(itemIDs) == 0 {
				var itemID int64
				_ = tx.QueryRow(`SELECT item_id FROM sales WHERE id = ?`, saleID.Int64).Scan(&itemID)
				if itemID > 0 {
					itemIDs = []int64{itemID}
				}
			}
			lotsTouched := map[int64]bool{}
			for _, itemID := range itemIDs {
				var lotID int64
				if err := tx.QueryRow(`SELECT lot_id FROM items WHERE id = ?`, itemID).Scan(&lotID); err != nil {
					return err
				}
				lotsTouched[lotID] = true
				if _, err := tx.Exec(`UPDATE items SET status = 'in_stock' WHERE id = ?`, itemID); err != nil {
					return err
				}
			}
			for lotID := range lotsTouched {
				var inStock, total int
				_ = tx.QueryRow(`SELECT COUNT(*) FROM items WHERE lot_id = ? AND status = 'in_stock'`, lotID).Scan(&inStock)
				_ = tx.QueryRow(`SELECT COUNT(*) FROM items WHERE lot_id = ?`, lotID).Scan(&total)
				lotStatus := "open"
				if inStock == 0 && total > 0 {
					lotStatus = "sold"
				} else if inStock > 0 && inStock < total {
					lotStatus = "partial"
				}
				if _, err := tx.Exec(`UPDATE lots SET status = ? WHERE id = ?`, lotStatus, lotID); err != nil {
					return err
				}
			}
		}
	}
	return tx.Commit()
}

type CreatePayableInput struct {
	Description string
	AmountCents int64
	DueOn       string
	LotID       int64 // optional 0
}

func (s *SQLiteStore) CreatePayable(in CreatePayableInput) (int64, error) {
	desc := strings.TrimSpace(in.Description)
	if desc == "" || in.AmountCents <= 0 || strings.TrimSpace(in.DueOn) == "" {
		return 0, fmt.Errorf("%w: description, amount and due_on required", ErrInvalidInput)
	}
	var lot any
	if in.LotID > 0 {
		lot = in.LotID
	}
	res, err := s.db.Exec(
		`INSERT INTO payables (description, amount_cents, due_on, status, lot_id) VALUES (?, ?, ?, 'open', ?)`,
		desc, in.AmountCents, strings.TrimSpace(in.DueOn), lot,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

type CreateReceivableInput struct {
	Description string
	AmountCents int64
	DueOn       string
}

func (s *SQLiteStore) CreateReceivable(in CreateReceivableInput) (int64, error) {
	desc := strings.TrimSpace(in.Description)
	if desc == "" || in.AmountCents <= 0 || strings.TrimSpace(in.DueOn) == "" {
		return 0, fmt.Errorf("%w: description, amount and due_on required", ErrInvalidInput)
	}
	res, err := s.db.Exec(
		`INSERT INTO receivables (description, amount_cents, due_on, status) VALUES (?, ?, ?, 'open')`,
		desc, in.AmountCents, strings.TrimSpace(in.DueOn),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
