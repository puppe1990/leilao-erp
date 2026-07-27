package models

import "time"

type Sale struct {
	ID                  int64
	ItemID              int64
	ItemTitle           string // main item title (joined); not a sales column
	LineCount           int    // total sale_lines (0 if unknown)
	Composition         string // e.g. "Monitor + 2 acessórios"
	SoldAt              string
	Channel             string
	GrossCents          int64
	FeeCents            int64
	ShippingCents       int64
	NetCents            int64
	PaymentStatus       string
	UnitCostCentsAtSale int64 // total cost of all lines
	CreatedAt           time.Time
}
