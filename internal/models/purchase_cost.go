package models

import "time"

type PurchaseCost struct {
	ID          int64
	LotID       int64
	Label       string
	AmountCents int64
	PayableID   *int64
	CreatedAt   time.Time
}
