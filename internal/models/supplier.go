package models

import "time"

// Supplier is a wholesale contact (25 de Março / Santa Efigênia).
type Supplier struct {
	ID            int64
	Name          string
	WhatsApp      string
	ChatID        string
	Notes         string
	MinOrderCents int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// SupplierMessage is one WhatsApp line saved from a supplier chat.
type SupplierMessage struct {
	ID          int64
	SupplierID  int64
	WAMessageID string
	FromMe      bool
	Body        string
	SentAt      time.Time
	CreatedAt   time.Time
}
