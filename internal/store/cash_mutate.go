package store

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/puppe1990/leilao-erp/internal/models"
)

func (s *SQLiteStore) FindCashAccount(id int64) (models.CashAccount, error) {
	var a models.CashAccount
	err := s.db.QueryRow(
		`SELECT id, name, kind, opening_balance_cents, created_at FROM cash_accounts WHERE id = ?`, id,
	).Scan(&a.ID, &a.Name, &a.Kind, &a.OpeningBalanceCents, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return models.CashAccount{}, ErrNotFound
	}
	if err != nil {
		return models.CashAccount{}, err
	}
	return a, nil
}

func (s *SQLiteStore) UpdateCashAccount(id int64, name, kind string, openingBalanceCents int64) error {
	name = strings.TrimSpace(name)
	kind = strings.TrimSpace(kind)
	if name == "" {
		return fmt.Errorf("%w: name required", ErrInvalidInput)
	}
	switch kind {
	case "pix", "bank", "cash", "other":
	default:
		return fmt.Errorf("%w: invalid kind", ErrInvalidInput)
	}
	res, err := s.db.Exec(
		`UPDATE cash_accounts SET name = ?, kind = ?, opening_balance_cents = ? WHERE id = ?`,
		name, kind, openingBalanceCents, id,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *SQLiteStore) DeleteCashAccount(id int64) error {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM cash_entries WHERE account_id = ?`, id).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("%w: account has cash entries", ErrCannotDelete)
	}
	res, err := s.db.Exec(`DELETE FROM cash_accounts WHERE id = ?`, id)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteCashEntry removes only manual adjustment entries.
func (s *SQLiteStore) DeleteCashEntry(id int64) error {
	var cat string
	err := s.db.QueryRow(`SELECT category FROM cash_entries WHERE id = ?`, id).Scan(&cat)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if cat != "ajuste" {
		return fmt.Errorf("%w: only manual adjustments can be deleted", ErrCannotDelete)
	}
	_, err = s.db.Exec(`DELETE FROM cash_entries WHERE id = ?`, id)
	return err
}
