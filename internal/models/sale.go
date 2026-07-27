package models

import "time"

type Sale struct {
	ID                   int64
	ItemID               int64
	SoldAt               string
	Channel              string
	GrossCents           int64
	FeeCents             int64
	ShippingCents        int64
	NetCents             int64
	PaymentStatus        string
	UnitCostCentsAtSale  int64
	CreatedAt            time.Time
}
