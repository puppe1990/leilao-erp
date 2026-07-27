package models

import "time"

type SaleLine struct {
	ID                  int64
	SaleID              int64
	ItemID              int64
	ItemTitle           string // joined
	UnitCostCentsAtSale int64
	Role                string // main | accessory
	CreatedAt           time.Time
}
