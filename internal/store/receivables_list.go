package store

import (
	"database/sql"
	"fmt"

	"github.com/puppe1990/leilao-erp/internal/models"
)

func (s *SQLiteStore) ListReceivables() ([]models.Receivable, error) {
	rows, err := s.db.Query(
		`SELECT id, description, amount_cents, due_on, status, sale_id, received_at, created_at
		 FROM receivables ORDER BY id`,
	)
	if err != nil {
		return nil, fmt.Errorf("list receivables: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.Receivable
	for rows.Next() {
		var r models.Receivable
		var saleID sql.NullInt64
		var receivedAt sql.NullString
		if err := rows.Scan(
			&r.ID, &r.Description, &r.AmountCents, &r.DueOn, &r.Status,
			&saleID, &receivedAt, &r.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan receivable: %w", err)
		}
		if saleID.Valid {
			v := saleID.Int64
			r.SaleID = &v
		}
		if receivedAt.Valid {
			v := receivedAt.String
			r.ReceivedAt = &v
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
