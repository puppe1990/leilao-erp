package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/puppe1990/leilao-erp/internal/domain"
	"github.com/puppe1990/leilao-erp/internal/models"
)

// SupplierInput creates or updates a wholesale supplier.
type SupplierInput struct {
	Name          string
	WhatsApp      string
	Notes         string
	MinOrderCents int64
}

// SupplierMessageInput is one WhatsApp line to persist.
type SupplierMessageInput struct {
	WAMessageID string
	FromMe      bool
	Body        string
	SentAt      time.Time
}

// SupplierListItem is a supplier plus last-message preview.
type SupplierListItem struct {
	models.Supplier
	LastBody     string
	LastFromMe   bool
	LastAt       time.Time
	MessageCount int
}

func normalizeSupplierInput(in SupplierInput) (SupplierInput, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return in, fmt.Errorf("%w: name required", ErrInvalidInput)
	}
	chatID := domain.WhatsAppChatID(in.WhatsApp)
	if chatID == "" {
		return in, fmt.Errorf("%w: whatsapp %q must be a Brazilian number", ErrInvalidInput, strings.TrimSpace(in.WhatsApp))
	}
	in.WhatsApp = strings.TrimSuffix(chatID, "@c.us")
	in.Notes = strings.TrimSpace(in.Notes)
	if in.MinOrderCents < 0 {
		return in, fmt.Errorf("%w: min_order_cents %d must be >= 0", ErrInvalidInput, in.MinOrderCents)
	}
	return in, nil
}

func supplierChatID(in SupplierInput) string {
	return in.WhatsApp + "@c.us"
}

// CreateSupplier inserts a supplier and returns its id.
func (s *SQLiteStore) CreateSupplier(in SupplierInput) (int64, error) {
	in, err := normalizeSupplierInput(in)
	if err != nil {
		return 0, err
	}
	res, err := s.db.Exec(
		`INSERT INTO suppliers (name, whatsapp, chat_id, notes, min_order_cents) VALUES (?, ?, ?, ?, ?)`,
		in.Name, in.WhatsApp, supplierChatID(in), in.Notes, in.MinOrderCents,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return 0, fmt.Errorf("%w: whatsapp already registered", ErrInvalidInput)
		}
		return 0, fmt.Errorf("insert supplier: %w", err)
	}
	return res.LastInsertId()
}

// UpdateSupplier updates mutable supplier fields.
func (s *SQLiteStore) UpdateSupplier(id int64, in SupplierInput) error {
	in, err := normalizeSupplierInput(in)
	if err != nil {
		return err
	}
	res, err := s.db.Exec(
		`UPDATE suppliers SET name = ?, whatsapp = ?, chat_id = ?, notes = ?, min_order_cents = ?,
		 updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		in.Name, in.WhatsApp, supplierChatID(in), in.Notes, in.MinOrderCents, id,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return fmt.Errorf("%w: whatsapp already registered", ErrInvalidInput)
		}
		return fmt.Errorf("update supplier: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteSupplier removes a supplier and its messages.
func (s *SQLiteStore) DeleteSupplier(id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin delete supplier: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Exec(`DELETE FROM supplier_messages WHERE supplier_id = ?`, id); err != nil {
		return fmt.Errorf("delete supplier messages %d: %w", id, err)
	}
	res, err := tx.Exec(`DELETE FROM suppliers WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete supplier %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit delete supplier: %w", err)
	}
	return nil
}

// GetSupplier returns one supplier.
func (s *SQLiteStore) GetSupplier(id int64) (models.Supplier, error) {
	row := s.db.QueryRow(
		`SELECT id, name, whatsapp, chat_id, notes, min_order_cents, created_at, updated_at
		 FROM suppliers WHERE id = ?`, id,
	)
	sup, err := scanSupplier(row)
	if err == sql.ErrNoRows {
		return models.Supplier{}, ErrNotFound
	}
	if err != nil {
		return models.Supplier{}, fmt.Errorf("get supplier %d: %w", id, err)
	}
	return sup, nil
}

// GetSupplierByChatID finds a supplier by WhatsApp chat id.
func (s *SQLiteStore) GetSupplierByChatID(chatID string) (models.Supplier, error) {
	chatID = domain.WhatsAppChatID(chatID)
	if chatID == "" {
		return models.Supplier{}, fmt.Errorf("%w: chat_id required", ErrInvalidInput)
	}
	row := s.db.QueryRow(
		`SELECT id, name, whatsapp, chat_id, notes, min_order_cents, created_at, updated_at
		 FROM suppliers WHERE chat_id = ?`, chatID,
	)
	sup, err := scanSupplier(row)
	if err == sql.ErrNoRows {
		return models.Supplier{}, ErrNotFound
	}
	if err != nil {
		return models.Supplier{}, fmt.Errorf("get supplier chat %s: %w", chatID, err)
	}
	return sup, nil
}

// ListSuppliers returns suppliers newest first, with last-message preview.
func (s *SQLiteStore) ListSuppliers() ([]SupplierListItem, error) {
	rows, err := s.db.Query(
		`SELECT s.id, s.name, s.whatsapp, s.chat_id, s.notes, s.min_order_cents, s.created_at, s.updated_at,
		        (SELECT m.body FROM supplier_messages m WHERE m.supplier_id = s.id ORDER BY m.sent_at DESC, m.id DESC LIMIT 1),
		        (SELECT m.from_me FROM supplier_messages m WHERE m.supplier_id = s.id ORDER BY m.sent_at DESC, m.id DESC LIMIT 1),
		        (SELECT m.sent_at FROM supplier_messages m WHERE m.supplier_id = s.id ORDER BY m.sent_at DESC, m.id DESC LIMIT 1),
		        (SELECT COUNT(*) FROM supplier_messages m WHERE m.supplier_id = s.id)
		 FROM suppliers s ORDER BY s.id DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list suppliers: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := make([]SupplierListItem, 0)
	for rows.Next() {
		item, err := scanSupplierListItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// SaveSupplierMessage inserts a WhatsApp line. Same wa_message_id returns the existing id.
func (s *SQLiteStore) SaveSupplierMessage(supplierID int64, in SupplierMessageInput) (int64, error) {
	in.WAMessageID = strings.TrimSpace(in.WAMessageID)
	if in.WAMessageID == "" {
		return 0, fmt.Errorf("%w: wa_message_id required", ErrInvalidInput)
	}
	if in.SentAt.IsZero() {
		return 0, fmt.Errorf("%w: sent_at required", ErrInvalidInput)
	}
	if _, err := s.GetSupplier(supplierID); err != nil {
		return 0, err
	}
	fromMe := 0
	if in.FromMe {
		fromMe = 1
	}
	_, err := s.db.Exec(
		`INSERT INTO supplier_messages (supplier_id, wa_message_id, from_me, body, sent_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(wa_message_id) DO NOTHING`,
		supplierID, in.WAMessageID, fromMe, in.Body, in.SentAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return 0, fmt.Errorf("insert supplier message: %w", err)
	}
	var id int64
	err = s.db.QueryRow(`SELECT id FROM supplier_messages WHERE wa_message_id = ?`, in.WAMessageID).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("get supplier message %s: %w", in.WAMessageID, err)
	}
	return id, nil
}

// ListSupplierMessages returns a supplier conversation oldest first.
func (s *SQLiteStore) ListSupplierMessages(supplierID int64) ([]models.SupplierMessage, error) {
	rows, err := s.db.Query(
		`SELECT id, supplier_id, wa_message_id, from_me, body, sent_at, created_at
		 FROM supplier_messages WHERE supplier_id = ? ORDER BY sent_at ASC, id ASC`,
		supplierID,
	)
	if err != nil {
		return nil, fmt.Errorf("list supplier messages %d: %w", supplierID, err)
	}
	defer func() { _ = rows.Close() }()
	out := make([]models.SupplierMessage, 0)
	for rows.Next() {
		m, err := scanSupplierMessage(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

type supplierScanner interface {
	Scan(dest ...any) error
}

func scanSupplier(row supplierScanner) (models.Supplier, error) {
	var s models.Supplier
	err := row.Scan(&s.ID, &s.Name, &s.WhatsApp, &s.ChatID, &s.Notes, &s.MinOrderCents, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func scanSupplierListItem(row supplierScanner) (SupplierListItem, error) {
	var (
		item     SupplierListItem
		lastBody sql.NullString
		lastFrom sql.NullInt64
		lastAt   sql.NullString
	)
	err := row.Scan(
		&item.ID, &item.Name, &item.WhatsApp, &item.ChatID, &item.Notes, &item.MinOrderCents,
		&item.CreatedAt, &item.UpdatedAt, &lastBody, &lastFrom, &lastAt, &item.MessageCount,
	)
	if err != nil {
		return item, err
	}
	if lastBody.Valid {
		item.LastBody = lastBody.String
	}
	if lastFrom.Valid {
		item.LastFromMe = lastFrom.Int64 == 1
	}
	if lastAt.Valid {
		item.LastAt = parseSupplierTime(lastAt.String)
	}
	return item, nil
}

func scanSupplierMessage(row supplierScanner) (models.SupplierMessage, error) {
	var (
		m      models.SupplierMessage
		fromMe int
		sentAt string
	)
	err := row.Scan(&m.ID, &m.SupplierID, &m.WAMessageID, &fromMe, &m.Body, &sentAt, &m.CreatedAt)
	if err != nil {
		return m, err
	}
	m.FromMe = fromMe == 1
	m.SentAt = parseSupplierTime(sentAt)
	return m, nil
}

func parseSupplierTime(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", time.RFC3339Nano} {
		if t, err := time.Parse(layout, raw); err == nil {
			return t
		}
	}
	return time.Time{}
}
