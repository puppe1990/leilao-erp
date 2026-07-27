package models

import "time"

type Payable struct {
	ID          int64
	Description string
	AmountCents int64
	DueOn       string // YYYY-MM-DD
	Status      string
	LotID       *int64
	PaidAt      *string
	CreatedAt   time.Time
}
