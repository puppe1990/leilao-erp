package models

import "time"

// Product is a catalog name reused across stock units (monitors/cables).
type Product struct {
	ID                 int64
	Name               string
	SalePriceHintCents *int64
	Kind               string // principal | accessory
	// QtyInStock is filled by ListStockProductGroups only.
	QtyInStock    int
	UnitCostCents int64 // representative unit cost (avg or first)
	// SampleItemID first in_stock unit for "Vender"
	SampleItemID int64
	SampleLotID  int64
	CreatedAt    time.Time
}
