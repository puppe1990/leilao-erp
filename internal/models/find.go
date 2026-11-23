package models

import "time"

// Find is a marketplace listing the shop is considering buying.
type Find struct {
	ID        int64
	Name      string
	URL       string
	Source    string // olx, facebook, kwara
	BuyCents  int64
	SellCents int64
	CreatedAt time.Time
	UpdatedAt time.Time
}
