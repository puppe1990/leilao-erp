package store

import (
	"database/sql"
	"fmt"

	"github.com/puppe1990/leilao-erp/internal/domain"
	"github.com/puppe1990/leilao-erp/internal/models"
)

type CreateSaleInput struct {
	ItemID        int64
	SoldAt        string // ISO datetime or date
	Channel       string // direct, mercadolivre, shopee, olx, other
	GrossCents    int64
	FeeCents      int64
	ShippingCents int64
	PaymentStatus string // "received" | "pending"
	CashAccountID int64  // required if received
	DueOn         string // YYYY-MM-DD required if pending
}

func (s *SQLiteStore) CreateSale(input CreateSaleInput) (saleID int64, err error) {
	switch input.PaymentStatus {
	case "received":
		if input.CashAccountID <= 0 {
			return 0, fmt.Errorf("cash account required when payment is received")
		}
	case "pending":
		if input.DueOn == "" {
			return 0, fmt.Errorf("due_on required when payment is pending")
		}
	default:
		return 0, fmt.Errorf("payment_status must be received or pending")
	}

	net := domain.SaleNet(input.GrossCents, input.FeeCents, input.ShippingCents)
	if net < 0 {
		return 0, fmt.Errorf("net amount cannot be negative (gross=%d fee=%d shipping=%d)",
			input.GrossCents, input.FeeCents, input.ShippingCents)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var lotID, unitCost int64
	var title, status string
	err = tx.QueryRow(
		`SELECT lot_id, title, unit_cost_cents, status FROM items WHERE id = ?`,
		input.ItemID,
	).Scan(&lotID, &title, &unitCost, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("item %d not found", input.ItemID)
		}
		return 0, fmt.Errorf("load item: %w", err)
	}
	if status != "in_stock" {
		return 0, fmt.Errorf("item %d is not in stock (status=%s)", input.ItemID, status)
	}

	res, err := tx.Exec(
		`INSERT INTO sales
		 (item_id, sold_at, channel, gross_cents, fee_cents, shipping_cents,
		  net_cents, payment_status, unit_cost_cents_at_sale)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		input.ItemID, input.SoldAt, input.Channel,
		input.GrossCents, input.FeeCents, input.ShippingCents,
		net, input.PaymentStatus, unitCost,
	)
	if err != nil {
		return 0, fmt.Errorf("insert sale: %w", err)
	}
	saleID, err = res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("sale id: %w", err)
	}

	if _, err := tx.Exec(
		`UPDATE items SET status = 'sold' WHERE id = ?`,
		input.ItemID,
	); err != nil {
		return 0, fmt.Errorf("mark item sold: %w", err)
	}

	switch input.PaymentStatus {
	case "received":
		if net > 0 {
			if _, err := tx.Exec(
				`INSERT INTO cash_entries
				 (account_id, direction, amount_cents, occurred_at, category, sale_id)
				 VALUES (?, 'in', ?, ?, 'venda', ?)`,
				input.CashAccountID, net, input.SoldAt, saleID,
			); err != nil {
				return 0, fmt.Errorf("insert cash entry: %w", err)
			}
		}
	case "pending":
		if net > 0 {
			desc := fmt.Sprintf("Venda — %s", title)
			if _, err := tx.Exec(
				`INSERT INTO receivables
				 (description, amount_cents, due_on, status, sale_id)
				 VALUES (?, ?, ?, 'open', ?)`,
				desc, net, input.DueOn, saleID,
			); err != nil {
				return 0, fmt.Errorf("insert receivable: %w", err)
			}
		}
	}

	var inStockLeft int
	if err := tx.QueryRow(
		`SELECT COUNT(*) FROM items WHERE lot_id = ? AND status = 'in_stock'`,
		lotID,
	).Scan(&inStockLeft); err != nil {
		return 0, fmt.Errorf("count in_stock items: %w", err)
	}
	lotStatus := "sold"
	if inStockLeft > 0 {
		lotStatus = "partial"
	}
	if _, err := tx.Exec(
		`UPDATE lots SET status = ? WHERE id = ?`,
		lotStatus, lotID,
	); err != nil {
		return 0, fmt.Errorf("update lot status: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit sale: %w", err)
	}
	return saleID, nil
}

func (s *SQLiteStore) ListSales() ([]models.Sale, error) {
	rows, err := s.db.Query(
		`SELECT id, item_id, sold_at, channel, gross_cents, fee_cents, shipping_cents,
		        net_cents, payment_status, unit_cost_cents_at_sale, created_at
		 FROM sales ORDER BY id`,
	)
	if err != nil {
		return nil, fmt.Errorf("list sales: %w", err)
	}
	defer rows.Close()

	var out []models.Sale
	for rows.Next() {
		var sale models.Sale
		if err := rows.Scan(
			&sale.ID, &sale.ItemID, &sale.SoldAt, &sale.Channel,
			&sale.GrossCents, &sale.FeeCents, &sale.ShippingCents,
			&sale.NetCents, &sale.PaymentStatus, &sale.UnitCostCentsAtSale,
			&sale.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan sale: %w", err)
		}
		out = append(out, sale)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *SQLiteStore) ListReceivables() ([]models.Receivable, error) {
	rows, err := s.db.Query(
		`SELECT id, description, amount_cents, due_on, status, sale_id, received_at, created_at
		 FROM receivables ORDER BY id`,
	)
	if err != nil {
		return nil, fmt.Errorf("list receivables: %w", err)
	}
	defer rows.Close()

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

func (s *SQLiteStore) FindLot(id int64) (models.Lot, error) {
	var lot models.Lot
	var auctionSource, notes sql.NullString
	err := s.db.QueryRow(
		`SELECT id, name, auction_source, purchased_at, status, notes, created_at
		 FROM lots WHERE id = ?`,
		id,
	).Scan(
		&lot.ID, &lot.Name, &auctionSource, &lot.PurchasedAt,
		&lot.Status, &notes, &lot.CreatedAt,
	)
	if err != nil {
		return models.Lot{}, fmt.Errorf("find lot: %w", err)
	}
	if auctionSource.Valid {
		lot.AuctionSource = &auctionSource.String
	}
	if notes.Valid {
		lot.Notes = &notes.String
	}
	return lot, nil
}
