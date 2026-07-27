package store

import (
	"database/sql"
	"fmt"

	"github.com/puppe1990/leilao-erp/internal/domain"
	"github.com/puppe1990/leilao-erp/internal/models"
)

type CreateSaleInput struct {
	// ItemID is the main item (usually the monitor). Required.
	ItemID int64
	// AccessoryIDs are optional extra items sold with the main (cables, etc.).
	AccessoryIDs  []int64
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
	if input.ItemID <= 0 {
		return 0, fmt.Errorf("main item is required")
	}

	// Deduplicate IDs while preserving main first.
	seen := map[int64]bool{input.ItemID: true}
	allIDs := []int64{input.ItemID}
	for _, id := range input.AccessoryIDs {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		allIDs = append(allIDs, id)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	type lineItem struct {
		id, lotID, unitCost int64
		title, status, role string
	}
	lines := make([]lineItem, 0, len(allIDs))
	var totalCost int64
	var mainTitle string
	lotsTouched := map[int64]bool{}

	for i, itemID := range allIDs {
		var li lineItem
		li.id = itemID
		li.role = "accessory"
		if i == 0 {
			li.role = "main"
		}
		err = tx.QueryRow(
			`SELECT lot_id, title, unit_cost_cents, status FROM items WHERE id = ?`,
			itemID,
		).Scan(&li.lotID, &li.title, &li.unitCost, &li.status)
		if err != nil {
			if err == sql.ErrNoRows {
				return 0, fmt.Errorf("item %d not found", itemID)
			}
			return 0, fmt.Errorf("load item: %w", err)
		}
		if li.status != "in_stock" {
			return 0, fmt.Errorf("item %d is not in stock (status=%s)", itemID, li.status)
		}
		if i == 0 {
			mainTitle = li.title
		}
		totalCost += li.unitCost
		lotsTouched[li.lotID] = true
		lines = append(lines, li)
	}

	res, err := tx.Exec(
		`INSERT INTO sales
		 (item_id, sold_at, channel, gross_cents, fee_cents, shipping_cents,
		  net_cents, payment_status, unit_cost_cents_at_sale)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		input.ItemID, input.SoldAt, input.Channel,
		input.GrossCents, input.FeeCents, input.ShippingCents,
		net, input.PaymentStatus, totalCost,
	)
	if err != nil {
		return 0, fmt.Errorf("insert sale: %w", err)
	}
	saleID, err = res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("sale id: %w", err)
	}

	for _, li := range lines {
		if _, err := tx.Exec(
			`INSERT INTO sale_lines (sale_id, item_id, unit_cost_cents_at_sale, role)
			 VALUES (?, ?, ?, ?)`,
			saleID, li.id, li.unitCost, li.role,
		); err != nil {
			return 0, fmt.Errorf("insert sale line: %w", err)
		}
		if _, err := tx.Exec(`UPDATE items SET status = 'sold' WHERE id = ?`, li.id); err != nil {
			return 0, fmt.Errorf("mark item sold: %w", err)
		}
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
			desc := fmt.Sprintf("Venda — %s", mainTitle)
			if len(lines) > 1 {
				desc = fmt.Sprintf("Venda — %s + %d acessório(s)", mainTitle, len(lines)-1)
			}
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

	for lotID := range lotsTouched {
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
		if _, err := tx.Exec(`UPDATE lots SET status = ? WHERE id = ?`, lotStatus, lotID); err != nil {
			return 0, fmt.Errorf("update lot status: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit sale: %w", err)
	}
	return saleID, nil
}

// CancelPendingSale cancels a pending sale: reverts all sale_lines items to in_stock,
// cancels the open receivable (if any), and refreshes lot statuses.
// Received sales cannot be cancelled in v1.
func (s *SQLiteStore) CancelPendingSale(saleID int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var paymentStatus string
	err = tx.QueryRow(
		`SELECT payment_status FROM sales WHERE id = ?`,
		saleID,
	).Scan(&paymentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("sale %d not found", saleID)
		}
		return fmt.Errorf("load sale: %w", err)
	}

	if paymentStatus == "received" {
		return fmt.Errorf("cannot cancel sale %d: payment already received", saleID)
	}
	if paymentStatus != "pending" {
		return fmt.Errorf("cannot cancel sale %d: payment_status=%s", saleID, paymentStatus)
	}

	var recID sql.NullInt64
	var recStatus sql.NullString
	err = tx.QueryRow(
		`SELECT id, status FROM receivables WHERE sale_id = ?`,
		saleID,
	).Scan(&recID, &recStatus)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("load receivable: %w", err)
	}
	if err == nil {
		if !recStatus.Valid || recStatus.String != "open" {
			status := "unknown"
			if recStatus.Valid {
				status = recStatus.String
			}
			return fmt.Errorf("cannot cancel sale %d: receivable is %s (must be open)", saleID, status)
		}
	}

	if _, err := tx.Exec(
		`UPDATE sales SET payment_status = 'cancelled' WHERE id = ?`,
		saleID,
	); err != nil {
		return fmt.Errorf("cancel sale: %w", err)
	}

	if recID.Valid {
		if _, err := tx.Exec(
			`UPDATE receivables SET status = 'cancelled' WHERE id = ?`,
			recID.Int64,
		); err != nil {
			return fmt.Errorf("cancel receivable: %w", err)
		}
	}

	// Prefer sale_lines; fall back to sales.item_id for edge cases.
	itemRows, err := tx.Query(`SELECT item_id FROM sale_lines WHERE sale_id = ?`, saleID)
	if err != nil {
		return fmt.Errorf("list sale lines: %w", err)
	}
	var itemIDs []int64
	for itemRows.Next() {
		var id int64
		if err := itemRows.Scan(&id); err != nil {
			_ = itemRows.Close()
			return err
		}
		itemIDs = append(itemIDs, id)
	}
	_ = itemRows.Close()
	if err := itemRows.Err(); err != nil {
		return err
	}
	if len(itemIDs) == 0 {
		var itemID int64
		if err := tx.QueryRow(`SELECT item_id FROM sales WHERE id = ?`, saleID).Scan(&itemID); err != nil {
			return err
		}
		itemIDs = []int64{itemID}
	}

	lotsTouched := map[int64]bool{}
	for _, itemID := range itemIDs {
		var lotID int64
		if err := tx.QueryRow(`SELECT lot_id FROM items WHERE id = ?`, itemID).Scan(&lotID); err != nil {
			return fmt.Errorf("load item lot: %w", err)
		}
		lotsTouched[lotID] = true
		if _, err := tx.Exec(`UPDATE items SET status = 'in_stock' WHERE id = ?`, itemID); err != nil {
			return fmt.Errorf("restore item in_stock: %w", err)
		}
	}

	for lotID := range lotsTouched {
		var total, inStock, sold int
		if err := tx.QueryRow(
			`SELECT COUNT(*),
			        COALESCE(SUM(CASE WHEN status = 'in_stock' THEN 1 ELSE 0 END), 0),
			        COALESCE(SUM(CASE WHEN status = 'sold' THEN 1 ELSE 0 END), 0)
			 FROM items WHERE lot_id = ?`,
			lotID,
		).Scan(&total, &inStock, &sold); err != nil {
			return fmt.Errorf("count lot items: %w", err)
		}
		lotStatus := "partial"
		switch {
		case total > 0 && inStock == total:
			lotStatus = "open"
		case total > 0 && sold == total:
			lotStatus = "sold"
		}
		if _, err := tx.Exec(`UPDATE lots SET status = ? WHERE id = ?`, lotStatus, lotID); err != nil {
			return fmt.Errorf("update lot status: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit cancel sale: %w", err)
	}
	return nil
}

func (s *SQLiteStore) ListSales() ([]models.Sale, error) {
	rows, err := s.db.Query(
		`SELECT s.id, s.item_id, COALESCE(i.title, ''), s.sold_at, s.channel,
		        s.gross_cents, s.fee_cents, s.shipping_cents,
		        s.net_cents, s.payment_status, s.unit_cost_cents_at_sale, s.created_at,
		        COALESCE((SELECT COUNT(*) FROM sale_lines sl WHERE sl.sale_id = s.id), 0)
		 FROM sales s
		 LEFT JOIN items i ON i.id = s.item_id
		 ORDER BY s.id DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list sales: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.Sale
	for rows.Next() {
		var sale models.Sale
		if err := rows.Scan(
			&sale.ID, &sale.ItemID, &sale.ItemTitle, &sale.SoldAt, &sale.Channel,
			&sale.GrossCents, &sale.FeeCents, &sale.ShippingCents,
			&sale.NetCents, &sale.PaymentStatus, &sale.UnitCostCentsAtSale,
			&sale.CreatedAt, &sale.LineCount,
		); err != nil {
			return nil, fmt.Errorf("scan sale: %w", err)
		}
		if sale.LineCount <= 1 {
			sale.Composition = sale.ItemTitle
		} else {
			sale.Composition = fmt.Sprintf("%s + %d acessório(s)", sale.ItemTitle, sale.LineCount-1)
		}
		out = append(out, sale)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
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
	if err == sql.ErrNoRows {
		return models.Lot{}, ErrNotFound
	}
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

// ListSaleLines returns line items for a sale (main + accessories).
func (s *SQLiteStore) ListSaleLines(saleID int64) ([]models.SaleLine, error) {
	rows, err := s.db.Query(
		`SELECT sl.id, sl.sale_id, sl.item_id, COALESCE(i.title, ''), sl.unit_cost_cents_at_sale, sl.role, sl.created_at
		 FROM sale_lines sl
		 LEFT JOIN items i ON i.id = sl.item_id
		 WHERE sl.sale_id = ?
		 ORDER BY CASE sl.role WHEN 'main' THEN 0 ELSE 1 END, sl.id`,
		saleID,
	)
	if err != nil {
		return nil, fmt.Errorf("list sale lines: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.SaleLine
	for rows.Next() {
		var ln models.SaleLine
		if err := rows.Scan(
			&ln.ID, &ln.SaleID, &ln.ItemID, &ln.ItemTitle,
			&ln.UnitCostCentsAtSale, &ln.Role, &ln.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan sale line: %w", err)
		}
		out = append(out, ln)
	}
	return out, rows.Err()
}
