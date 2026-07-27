package models

import "time"

type Receivable struct {
	ID          int64
	Description string
	AmountCents int64
	DueOn       string // YYYY-MM-DD
	Status      string
	SaleID      *int64
	ReceivedAt  *string
	CreatedAt   time.Time
}
