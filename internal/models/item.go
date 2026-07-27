package models

import "time"

type Item struct {
	ID                 int64
	LotID              int64
	SKU                *string
	Title              string
	Condition          *string
	UnitCostCents      int64
	Status             string
	SalePriceHintCents *int64
	CreatedAt          time.Time
}
