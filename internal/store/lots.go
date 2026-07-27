package store

import (
	"database/sql"
	"fmt"

	"github.com/puppe1990/leilao-erp/internal/domain"
	"github.com/puppe1990/leilao-erp/internal/models"
)

type CostInput struct {
	Label       string
	AmountCents int64
	AlreadyPaid bool
}

type CreateLotInput struct {
	Name          string
	AuctionSource string // optional empty
	PurchasedAt   string // YYYY-MM-DD
	Notes         string // optional
	ItemTitle     string
	ItemQty       int
	Costs         []CostInput
	CashAccountID int64  // required if any cost AlreadyPaid
	PaidAt        string // ISO datetime for cash/payable paid_at
}

func (s *SQLiteStore) InsertCashAccount(name, kind string, openingBalanceCents int64) (int64, error) {
	result, err := s.db.Exec(
		`INSERT INTO cash_accounts (name, kind, opening_balance_cents) VALUES (?, ?, ?)`,
		name, kind, openingBalanceCents,
	)
	if err != nil {
		return 0, fmt.Errorf("insert cash account: %w", err)
	}
	return result.LastInsertId()
}

func (s *SQLiteStore) CreateLotPurchase(input CreateLotInput) (lotID int64, err error) {
	if input.ItemQty <= 0 {
		return 0, fmt.Errorf("item qty must be > 0")
	}
	var total int64
	var anyPositive bool
	var anyPaid bool
	for _, c := range input.Costs {
		if c.AmountCents > 0 {
			anyPositive = true
			total += c.AmountCents
		}
		if c.AlreadyPaid {
			anyPaid = true
		}
	}
	if !anyPositive || total <= 0 {
		return 0, fmt.Errorf("at least one cost with amount_cents > 0 is required")
	}
	if anyPaid && input.CashAccountID <= 0 {
		return 0, fmt.Errorf("cash account required when any cost is already paid")
	}
	if anyPaid && input.PaidAt == "" {
		return 0, fmt.Errorf("paid_at required when any cost is already paid")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var auctionSource, notes any
	if input.AuctionSource != "" {
		auctionSource = input.AuctionSource
	}
	if input.Notes != "" {
		notes = input.Notes
	}

	res, err := tx.Exec(
		`INSERT INTO lots (name, auction_source, purchased_at, status, notes)
		 VALUES (?, ?, ?, 'open', ?)`,
		input.Name, auctionSource, input.PurchasedAt, notes,
	)
	if err != nil {
		return 0, fmt.Errorf("insert lot: %w", err)
	}
	lotID, err = res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("lot id: %w", err)
	}

	for _, c := range input.Costs {
		if c.AmountCents <= 0 {
			continue
		}

		status := "open"
		var paidAt any
		if c.AlreadyPaid {
			status = "paid"
			paidAt = input.PaidAt
		}

		desc := c.Label
		if input.Name != "" {
			desc = fmt.Sprintf("%s — %s", c.Label, input.Name)
		}

		pres, err := tx.Exec(
			`INSERT INTO payables (description, amount_cents, due_on, status, lot_id, paid_at)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			desc, c.AmountCents, input.PurchasedAt, status, lotID, paidAt,
		)
		if err != nil {
			return 0, fmt.Errorf("insert payable: %w", err)
		}
		payableID, err := pres.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("payable id: %w", err)
		}

		if _, err := tx.Exec(
			`INSERT INTO purchase_costs (lot_id, label, amount_cents, payable_id)
			 VALUES (?, ?, ?, ?)`,
			lotID, c.Label, c.AmountCents, payableID,
		); err != nil {
			return 0, fmt.Errorf("insert purchase cost: %w", err)
		}

		if c.AlreadyPaid {
			if _, err := tx.Exec(
				`INSERT INTO cash_entries
				 (account_id, direction, amount_cents, occurred_at, category, payable_id, lot_id)
				 VALUES (?, 'out', ?, ?, 'compra_lote', ?, ?)`,
				input.CashAccountID, c.AmountCents, input.PaidAt, payableID, lotID,
			); err != nil {
				return 0, fmt.Errorf("insert cash entry: %w", err)
			}
		}
	}

	units := domain.AllocateUnitCosts(total, input.ItemQty)
	for i := 0; i < input.ItemQty; i++ {
		if _, err := tx.Exec(
			`INSERT INTO items (lot_id, title, unit_cost_cents, status)
			 VALUES (?, ?, ?, 'in_stock')`,
			lotID, input.ItemTitle, units[i],
		); err != nil {
			return 0, fmt.Errorf("insert item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit lot purchase: %w", err)
	}
	return lotID, nil
}

// AddPurchaseCost appends a cost to an existing lot, inserts payable (+ cash if paid),
// and reallocates unit_cost_cents across non-sold items only.
// Sold items keep their unit_cost_cents; remaining cost is allocated to in_stock + reserved.
func (s *SQLiteStore) AddPurchaseCost(lotID int64, cost CostInput, cashAccountID int64, paidAt string) error {
	if cost.AmountCents <= 0 {
		return fmt.Errorf("cost amount_cents must be > 0")
	}
	if cost.AlreadyPaid && cashAccountID <= 0 {
		return fmt.Errorf("cash account required when cost is already paid")
	}
	if cost.AlreadyPaid && paidAt == "" {
		return fmt.Errorf("paid_at required when cost is already paid")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var lotName, purchasedAt string
	err = tx.QueryRow(
		`SELECT name, purchased_at FROM lots WHERE id = ?`,
		lotID,
	).Scan(&lotName, &purchasedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("lot %d not found", lotID)
		}
		return fmt.Errorf("load lot: %w", err)
	}

	status := "open"
	var paidAtVal any
	if cost.AlreadyPaid {
		status = "paid"
		paidAtVal = paidAt
	}

	desc := cost.Label
	if lotName != "" {
		desc = fmt.Sprintf("%s — %s", cost.Label, lotName)
	}

	pres, err := tx.Exec(
		`INSERT INTO payables (description, amount_cents, due_on, status, lot_id, paid_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		desc, cost.AmountCents, purchasedAt, status, lotID, paidAtVal,
	)
	if err != nil {
		return fmt.Errorf("insert payable: %w", err)
	}
	payableID, err := pres.LastInsertId()
	if err != nil {
		return fmt.Errorf("payable id: %w", err)
	}

	if _, err := tx.Exec(
		`INSERT INTO purchase_costs (lot_id, label, amount_cents, payable_id)
		 VALUES (?, ?, ?, ?)`,
		lotID, cost.Label, cost.AmountCents, payableID,
	); err != nil {
		return fmt.Errorf("insert purchase cost: %w", err)
	}

	if cost.AlreadyPaid {
		if _, err := tx.Exec(
			`INSERT INTO cash_entries
			 (account_id, direction, amount_cents, occurred_at, category, payable_id, lot_id)
			 VALUES (?, 'out', ?, ?, 'compra_lote', ?, ?)`,
			cashAccountID, cost.AmountCents, paidAt, payableID, lotID,
		); err != nil {
			return fmt.Errorf("insert cash entry: %w", err)
		}
	}

	// totalCosts = sum of all purchase_costs for the lot
	var totalCosts int64
	if err := tx.QueryRow(
		`SELECT COALESCE(SUM(amount_cents), 0) FROM purchase_costs WHERE lot_id = ?`,
		lotID,
	).Scan(&totalCosts); err != nil {
		return fmt.Errorf("sum purchase costs: %w", err)
	}

	// sold items keep unit_cost; remaining allocated to in_stock + reserved
	rows, err := tx.Query(
		`SELECT id, unit_cost_cents, status FROM items WHERE lot_id = ? ORDER BY id`,
		lotID,
	)
	if err != nil {
		return fmt.Errorf("list items for reallocation: %w", err)
	}
	defer rows.Close()

	var soldCostSum int64
	type openItem struct {
		id int64
	}
	var open []openItem
	for rows.Next() {
		var id, unitCost int64
		var itemStatus string
		if err := rows.Scan(&id, &unitCost, &itemStatus); err != nil {
			return fmt.Errorf("scan item: %w", err)
		}
		if itemStatus == "sold" {
			soldCostSum += unitCost
			continue
		}
		// in_stock + reserved
		open = append(open, openItem{id: id})
	}
	if err := rows.Err(); err != nil {
		return err
	}
	// close rows before further Exec on same connection
	if err := rows.Close(); err != nil {
		return err
	}

	if len(open) > 0 {
		remaining := totalCosts - soldCostSum
		units := domain.AllocateUnitCosts(remaining, len(open))
		for i, it := range open {
			if _, err := tx.Exec(
				`UPDATE items SET unit_cost_cents = ? WHERE id = ?`,
				units[i], it.id,
			); err != nil {
				return fmt.Errorf("update item unit cost: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit add purchase cost: %w", err)
	}
	return nil
}

func (s *SQLiteStore) ListItemsByLot(lotID int64) ([]models.Item, error) {
	rows, err := s.db.Query(
		`SELECT id, lot_id, sku, title, condition, unit_cost_cents, status,
		        sale_price_hint_cents, created_at
		 FROM items WHERE lot_id = ? ORDER BY id`,
		lotID,
	)
	if err != nil {
		return nil, fmt.Errorf("list items by lot: %w", err)
	}
	defer rows.Close()

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

func (s *SQLiteStore) ListPayablesByLot(lotID int64) ([]models.Payable, error) {
	rows, err := s.db.Query(
		`SELECT id, description, amount_cents, due_on, status, lot_id, paid_at, created_at
		 FROM payables WHERE lot_id = ? ORDER BY id`,
		lotID,
	)
	if err != nil {
		return nil, fmt.Errorf("list payables by lot: %w", err)
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

// CashBalance returns opening_balance + sum(in) - sum(out) for the account.
func (s *SQLiteStore) CashBalance(accountID int64) (int64, error) {
	var opening int64
	err := s.db.QueryRow(
		`SELECT opening_balance_cents FROM cash_accounts WHERE id = ?`,
		accountID,
	).Scan(&opening)
	if err != nil {
		return 0, fmt.Errorf("cash account: %w", err)
	}

	var net int64
	err = s.db.QueryRow(
		`SELECT COALESCE(SUM(
			CASE direction
				WHEN 'in' THEN amount_cents
				WHEN 'out' THEN -amount_cents
				ELSE 0
			END
		), 0) FROM cash_entries WHERE account_id = ?`,
		accountID,
	).Scan(&net)
	if err != nil {
		return 0, fmt.Errorf("cash entries: %w", err)
	}
	return opening + net, nil
}
