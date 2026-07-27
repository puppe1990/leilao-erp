package models

import "time"

type Lot struct {
	ID            int64
	Name          string
	AuctionSource *string
	PurchasedAt   string // YYYY-MM-DD
	Status        string
	Notes         *string
	CreatedAt     time.Time
}
