package store

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/puppe1990/leilao-erp/internal/models"
)

// ListCashEntries returns cash ledger rows, newest first.
// If accountID > 0, filters by that account; otherwise returns all accounts.
func (s *SQLiteStore) ListCashEntries(accountID int64) ([]models.CashEntry, error) {
	q := `SELECT id, account_id, direction, amount_cents, occurred_at, category, memo,
	             sale_id, payable_id, receivable_id, lot_id, created_at
	      FROM cash_entries`
	var args []any
	if accountID > 0 {
		q += ` WHERE account_id = ?`
		args = append(args, accountID)
	}
	q += ` ORDER BY id DESC`

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("list cash entries: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.CashEntry
	for rows.Next() {
		e, err := scanCashEntry(rows)
		if err != nil {
			return nil, fmt.Errorf("scan cash entry: %w", err)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ManualCashCategories are categories allowed for free-form cash ledger rows.
var ManualCashCategories = map[string]bool{
	"ajuste":  true,
	"despesa": true,
	"frete":   true,
	"taxa":    true,
}

func normalizeManualCashCategory(category string) (string, error) {
	c := strings.TrimSpace(strings.ToLower(category))
	if c == "" {
		c = "ajuste"
	}
	if !ManualCashCategories[c] {
		return "", fmt.Errorf("%w: category must be ajuste, despesa, frete or taxa", ErrInvalidInput)
	}
	return c, nil
}

// InsertManualCashEntry records a free-form ledger row (not linked to sale/lot/payable).
func (s *SQLiteStore) InsertManualCashEntry(accountID int64, direction string, amountCents int64, occurredAt, category, memo string) (int64, error) {
	direction = strings.TrimSpace(direction)
	if direction != "in" && direction != "out" {
		return 0, fmt.Errorf("direction must be in or out")
	}
	if amountCents <= 0 {
		return 0, fmt.Errorf("amount must be positive")
	}
	if accountID <= 0 {
		return 0, fmt.Errorf("account_id required")
	}
	if strings.TrimSpace(occurredAt) == "" {
		return 0, fmt.Errorf("occurred_at required")
	}
	cat, err := normalizeManualCashCategory(category)
	if err != nil {
		return 0, err
	}

	var memoArg any
	if m := strings.TrimSpace(memo); m != "" {
		memoArg = m
	}

	result, err := s.db.Exec(
		`INSERT INTO cash_entries (account_id, direction, amount_cents, occurred_at, category, memo)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		accountID, direction, amountCents, occurredAt, cat, memoArg,
	)
	if err != nil {
		return 0, fmt.Errorf("insert manual cash entry: %w", err)
	}
	return result.LastInsertId()
}

func scanCashEntry(row interface {
	Scan(dest ...any) error
}) (models.CashEntry, error) {
	var e models.CashEntry
	var memo sql.NullString
	var saleID, payableID, receivableID, lotID sql.NullInt64
	if err := row.Scan(
		&e.ID, &e.AccountID, &e.Direction, &e.AmountCents, &e.OccurredAt, &e.Category, &memo,
		&saleID, &payableID, &receivableID, &lotID, &e.CreatedAt,
	); err != nil {
		return models.CashEntry{}, err
	}
	if memo.Valid {
		v := memo.String
		e.Memo = &v
	}
	if saleID.Valid {
		v := saleID.Int64
		e.SaleID = &v
	}
	if payableID.Valid {
		v := payableID.Int64
		e.PayableID = &v
	}
	if receivableID.Valid {
		v := receivableID.Int64
		e.ReceivableID = &v
	}
	if lotID.Valid {
		v := lotID.Int64
		e.LotID = &v
	}
	return e, nil
}

// CashEntryIsManual reports whether the entry can be freely edited/deleted.
func CashEntryIsManual(e models.CashEntry) bool {
	if e.SaleID != nil || e.PayableID != nil || e.ReceivableID != nil || e.LotID != nil {
		return false
	}
	return ManualCashCategories[e.Category]
}

func (s *SQLiteStore) FindCashEntry(id int64) (models.CashEntry, error) {
	e, err := scanCashEntry(s.db.QueryRow(
		`SELECT id, account_id, direction, amount_cents, occurred_at, category, memo,
		        sale_id, payable_id, receivable_id, lot_id, created_at
		 FROM cash_entries WHERE id = ?`, id,
	))
	if err == sql.ErrNoRows {
		return models.CashEntry{}, ErrNotFound
	}
	if err != nil {
		return models.CashEntry{}, fmt.Errorf("find cash entry: %w", err)
	}
	return e, nil
}

// UpdateCashEntry updates a free-form ledger row.
func (s *SQLiteStore) UpdateCashEntry(id int64, accountID int64, direction string, amountCents int64, occurredAt, category, memo string) error {
	cur, err := s.FindCashEntry(id)
	if err != nil {
		return err
	}
	if !CashEntryIsManual(cur) {
		return fmt.Errorf("%w: only manual cash entries can be edited", ErrCannotUpdate)
	}
	direction = strings.TrimSpace(direction)
	if direction != "in" && direction != "out" {
		return fmt.Errorf("%w: direction must be in or out", ErrInvalidInput)
	}
	if amountCents <= 0 {
		return fmt.Errorf("%w: amount must be positive", ErrInvalidInput)
	}
	if accountID <= 0 {
		return fmt.Errorf("%w: account_id required", ErrInvalidInput)
	}
	if strings.TrimSpace(occurredAt) == "" {
		return fmt.Errorf("%w: occurred_at required", ErrInvalidInput)
	}
	cat, err := normalizeManualCashCategory(category)
	if err != nil {
		return err
	}
	var memoArg any
	if m := strings.TrimSpace(memo); m != "" {
		memoArg = m
	}
	res, err := s.db.Exec(
		`UPDATE cash_entries
		 SET account_id = ?, direction = ?, amount_cents = ?, occurred_at = ?, category = ?, memo = ?
		 WHERE id = ?`,
		accountID, direction, amountCents, occurredAt, cat, memoArg, id,
	)
	if err != nil {
		return fmt.Errorf("update cash entry: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *SQLiteStore) ListPayables() ([]models.Payable, error) {
	rows, err := s.db.Query(
		`SELECT id, description, amount_cents, due_on, status, lot_id, paid_at, created_at
		 FROM payables ORDER BY id DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list payables: %w", err)
	}
	defer func() { _ = rows.Close() }()

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
