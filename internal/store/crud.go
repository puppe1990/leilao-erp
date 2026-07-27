package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/puppe1990/leilao-erp/internal/models"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrCannotDelete = errors.New("cannot delete")
	ErrCannotUpdate = errors.New("cannot update")
	ErrInvalidInput = errors.New("invalid input")
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

// UpdateItem updates mutable fields for an in-stock item.
func (s *SQLiteStore) UpdateItem(id int64, title, sku string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return fmt.Errorf("%w: title required", ErrInvalidInput)
	}
	var status string
	err := s.db.QueryRow(`SELECT status FROM items WHERE id = ?`, id).Scan(&status)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if status != "in_stock" && status != "reserved" {
		return fmt.Errorf("%w: only in-stock items can be edited", ErrCannotUpdate)
	}
	var skuVal any
	if strings.TrimSpace(sku) != "" {
		skuVal = strings.TrimSpace(sku)
	}
	_, err = s.db.Exec(`UPDATE items SET title = ?, sku = ? WHERE id = ?`, title, skuVal, id)
	return err
}

func (s *SQLiteStore) FindItemByID(id int64) (models.Item, error) {
	var it models.Item
	var sku, condition sql.NullString
	var hint sql.NullInt64
	err := s.db.QueryRow(
		`SELECT id, lot_id, sku, title, condition, unit_cost_cents, status, sale_price_hint_cents, created_at
		 FROM items WHERE id = ?`, id,
	).Scan(&it.ID, &it.LotID, &sku, &it.Title, &condition, &it.UnitCostCents, &it.Status, &hint, &it.CreatedAt)
	if err == sql.ErrNoRows {
		return models.Item{}, ErrNotFound
	}
	if err != nil {
		return models.Item{}, err
	}
	if sku.Valid {
		v := sku.String
		it.SKU = &v
	}
	if condition.Valid {
		v := condition.String
		it.Condition = &v
	}
	if hint.Valid {
		v := hint.Int64
		it.SalePriceHintCents = &v
	}
	return it, nil
}

// --- Sales ---

func (s *SQLiteStore) FindSaleByID(id int64) (models.Sale, error) {
	var sale models.Sale
	err := s.db.QueryRow(
		`SELECT s.id, s.item_id, s.sold_at, s.channel, s.gross_cents, s.fee_cents, s.shipping_cents,
		        s.net_cents, s.payment_status, s.unit_cost_cents_at_sale, s.created_at,
		        COALESCE(i.title, '')
		 FROM sales s
		 LEFT JOIN items i ON i.id = s.item_id
		 WHERE s.id = ?`, id,
	).Scan(
		&sale.ID, &sale.ItemID, &sale.SoldAt, &sale.Channel, &sale.GrossCents, &sale.FeeCents,
		&sale.ShippingCents, &sale.NetCents, &sale.PaymentStatus, &sale.UnitCostCentsAtSale,
		&sale.CreatedAt, &sale.ItemTitle,
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

// --- Cash accounts ---

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

// --- Payables / Receivables ---

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
		var pstatus string
		var itemID, lotID int64
		err = tx.QueryRow(
			`SELECT s.payment_status, s.item_id, i.lot_id FROM sales s JOIN items i ON i.id = s.item_id WHERE s.id = ?`,
			saleID.Int64,
		).Scan(&pstatus, &itemID, &lotID)
		if err == nil && pstatus == "pending" {
			if _, err := tx.Exec(`UPDATE sales SET payment_status = 'cancelled' WHERE id = ?`, saleID.Int64); err != nil {
				return err
			}
			if _, err := tx.Exec(`UPDATE items SET status = 'in_stock' WHERE id = ?`, itemID); err != nil {
				return err
			}
			// recompute lot status
			var inStock, total int
			_ = tx.QueryRow(`SELECT COUNT(*) FROM items WHERE lot_id = ? AND status = 'in_stock'`, lotID).Scan(&inStock)
			_ = tx.QueryRow(`SELECT COUNT(*) FROM items WHERE lot_id = ?`, lotID).Scan(&total)
			lotStatus := "open"
			if inStock == 0 && total > 0 {
				lotStatus = "sold"
			} else if inStock < total {
				lotStatus = "partial"
			}
			if _, err := tx.Exec(`UPDATE lots SET status = ? WHERE id = ?`, lotStatus, lotID); err != nil {
				return err
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
