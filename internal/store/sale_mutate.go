package store

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/puppe1990/leilao-erp/internal/models"
)

func (s *SQLiteStore) FindSaleByID(id int64) (models.Sale, error) {
	var sale models.Sale
	err := s.db.QueryRow(
		`SELECT s.id, s.item_id, s.sold_at, s.channel, s.gross_cents, s.fee_cents, s.shipping_cents,
		        s.net_cents, s.payment_status, s.unit_cost_cents_at_sale, s.created_at,
		        COALESCE(i.title, ''),
		        COALESCE(s.client_id, 0), COALESCE(c.name, ''), COALESCE(c.document, ''),
		        COALESCE(c.phone, ''), COALESCE(c.email, '')
		 FROM sales s
		 LEFT JOIN items i ON i.id = s.item_id
		 LEFT JOIN clients c ON c.id = s.client_id
		 WHERE s.id = ?`, id,
	).Scan(
		&sale.ID, &sale.ItemID, &sale.SoldAt, &sale.Channel, &sale.GrossCents, &sale.FeeCents,
		&sale.ShippingCents, &sale.NetCents, &sale.PaymentStatus, &sale.UnitCostCentsAtSale,
		&sale.CreatedAt, &sale.ItemTitle,
		&sale.ClientID, &sale.ClientName, &sale.ClientDocument,
		&sale.ClientPhone, &sale.ClientEmail,
	)
	if err == sql.ErrNoRows {
		return models.Sale{}, ErrNotFound
	}
	if err != nil {
		return models.Sale{}, err
	}
	return sale, nil
}

type UpdateSaleInput struct {
	SoldAt        string
	Channel       string
	GrossCents    int64
	FeeCents      int64
	ShippingCents int64
	DueOn         string // for pending receivable
}

// UpdateSale edits a pending sale only (not yet received).
func (s *SQLiteStore) UpdateSale(id int64, in UpdateSaleInput) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var status string
	err = tx.QueryRow(`SELECT payment_status FROM sales WHERE id = ?`, id).Scan(&status)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if status != "pending" {
		return fmt.Errorf("%w: only pending sales can be edited", ErrCannotUpdate)
	}

	net := in.GrossCents - in.FeeCents - in.ShippingCents
	if in.GrossCents <= 0 || net < 0 {
		return fmt.Errorf("%w: invalid amounts", ErrInvalidInput)
	}
	channel := strings.TrimSpace(in.Channel)
	if channel == "" {
		channel = "other"
	}
	soldAt := strings.TrimSpace(in.SoldAt)
	if soldAt == "" {
		return fmt.Errorf("%w: sold_at required", ErrInvalidInput)
	}

	_, err = tx.Exec(
		`UPDATE sales SET sold_at = ?, channel = ?, gross_cents = ?, fee_cents = ?, shipping_cents = ?, net_cents = ?
		 WHERE id = ?`,
		soldAt, channel, in.GrossCents, in.FeeCents, in.ShippingCents, net, id,
	)
	if err != nil {
		return err
	}

	// update open receivable amount/due if exists
	dueOn := strings.TrimSpace(in.DueOn)
	if dueOn != "" {
		_, err = tx.Exec(
			`UPDATE receivables SET amount_cents = ?, due_on = ?, description = ?
			 WHERE sale_id = ? AND status = 'open'`,
			net, dueOn, fmt.Sprintf("Venda #%d", id), id,
		)
		if err != nil {
			return err
		}
	} else {
		_, err = tx.Exec(
			`UPDATE receivables SET amount_cents = ? WHERE sale_id = ? AND status = 'open'`,
			net, id,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// DeleteSale cancels pending sales (same as CancelPendingSale) or rejects received ones.
func (s *SQLiteStore) DeleteSale(id int64) error {
	return s.CancelPendingSale(id)
}

// SetSaleClient links a buyer to a non-cancelled sale.
func (s *SQLiteStore) SetSaleClient(saleID, clientID int64) error {
	if clientID <= 0 {
		return fmt.Errorf("%w: client id %d is invalid, want id > 0", ErrInvalidInput, clientID)
	}
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM clients WHERE id = ?`, clientID).Scan(&n); err != nil {
		return fmt.Errorf("check client: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("client %d not found", clientID)
	}
	var status string
	err := s.db.QueryRow(`SELECT payment_status FROM sales WHERE id = ?`, saleID).Scan(&status)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("load sale: %w", err)
	}
	if status == "cancelled" {
		return fmt.Errorf("%w: cannot set client on cancelled sale %d", ErrCannotUpdate, saleID)
	}
	res, err := s.db.Exec(`UPDATE sales SET client_id = ? WHERE id = ?`, clientID, saleID)
	if err != nil {
		return fmt.Errorf("set sale client: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}
