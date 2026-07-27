package models

import "time"

type CashEntry struct {
	ID           int64
	AccountID    int64
	Direction    string
	AmountCents  int64
	OccurredAt   string
	Category     string
	Memo         *string
	SaleID       *int64
	PayableID    *int64
	ReceivableID *int64
	LotID        *int64
	CreatedAt    time.Time
}
