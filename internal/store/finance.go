package store

import (
	"database/sql"
	"fmt"

	"github.com/puppe1990/leilao-erp/internal/models"
)

func (s *SQLiteStore) ListPayables() ([]models.Payable, error) {
	rows, err := s.db.Query(
		`SELECT id, description, amount_cents, due_on, status, lot_id, paid_at, created_at
		 FROM payables ORDER BY id DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list payables: %w", err)
	}
	defer rows.Close()

	var out []models.Payable
	for rows.Next() {
		var p models.Payable
		var lot sql.NullInt64
		var paidAt sql.NullString
		if err := rows.Scan(
			&p.ID, &p.Description, &p.AmountCents, &p.DueOn, &p.Status,
			&lot, &paidAt, &p.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan payable: %w", err)
		}
		if lot.Valid {
			v := lot.Int64
			p.LotID = &v
		}
		if paidAt.Valid {
			v := paidAt.String
			p.PaidAt = &v
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// SettlePayable marks an open payable as paid and records a cash out entry.
func (s *SQLiteStore) SettlePayable(id, accountID int64, paidAt string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var amountCents int64
	var status string
	var lotID sql.NullInt64
	err = tx.QueryRow(
		`SELECT amount_cents, status, lot_id FROM payables WHERE id = ?`,
		id,
	).Scan(&amountCents, &status, &lotID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("payable %d not found", id)
		}
		return fmt.Errorf("load payable: %w", err)
	}
	if status != "open" {
		return fmt.Errorf("payable %d is not open (status=%s)", id, status)
	}

	if _, err := tx.Exec(
		`UPDATE payables SET status = 'paid', paid_at = ? WHERE id = ?`,
		paidAt, id,
	); err != nil {
		return fmt.Errorf("update payable: %w", err)
	}

	var lot any
	if lotID.Valid {
		lot = lotID.Int64
	}

	if _, err := tx.Exec(
		`INSERT INTO cash_entries
		 (account_id, direction, amount_cents, occurred_at, category, payable_id, lot_id)
		 VALUES (?, 'out', ?, ?, 'pagamento', ?, ?)`,
		accountID, amountCents, paidAt, id, lot,
	); err != nil {
		return fmt.Errorf("insert cash entry: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit settle payable: %w", err)
	}
	return nil
}

// SettleReceivable marks an open receivable as received, records a cash in entry,
// and updates the linked sale's payment_status when sale_id is set.
func (s *SQLiteStore) SettleReceivable(id, accountID int64, receivedAt string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var amountCents int64
	var status string
	var saleID sql.NullInt64
	err = tx.QueryRow(
		`SELECT amount_cents, status, sale_id FROM receivables WHERE id = ?`,
		id,
	).Scan(&amountCents, &status, &saleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("receivable %d not found", id)
		}
		return fmt.Errorf("load receivable: %w", err)
	}
	if status != "open" {
		return fmt.Errorf("receivable %d is not open (status=%s)", id, status)
	}

	if _, err := tx.Exec(
		`UPDATE receivables SET status = 'received', received_at = ? WHERE id = ?`,
		receivedAt, id,
	); err != nil {
		return fmt.Errorf("update receivable: %w", err)
	}

	if _, err := tx.Exec(
		`INSERT INTO cash_entries
		 (account_id, direction, amount_cents, occurred_at, category, receivable_id)
		 VALUES (?, 'in', ?, ?, 'recebimento', ?)`,
		accountID, amountCents, receivedAt, id,
	); err != nil {
		return fmt.Errorf("insert cash entry: %w", err)
	}

	if saleID.Valid {
		if _, err := tx.Exec(
			`UPDATE sales SET payment_status = 'received' WHERE id = ?`,
			saleID.Int64,
		); err != nil {
			return fmt.Errorf("update sale payment_status: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit settle receivable: %w", err)
	}
	return nil
}
