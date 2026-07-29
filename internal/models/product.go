package models

import "time"

// Product is a catalog name reused across stock units (monitors/cables).
type Product struct {
	ID                 int64
	Name               string
	Slug               string // public catalog URL segment (/produto/{slug})
	SalePriceHintCents *int64
	Kind               string // principal | accessory
	// QtyInStock is filled by ListStockProductGroups only.
	QtyInStock    int
	UnitCostCents int64 // representative unit cost (avg or first)
	// SampleItemID first in_stock unit for "Vender"
	SampleItemID int64
	SampleLotID  int64
	// PhotoCount / VideoCount filled by ListProducts.
	PhotoCount int
	VideoCount int
	// FirstPhotoURL is the first photo (sort_order, id) for list thumbnails.
	FirstPhotoURL string
	// Description: technical specs; ListingText: marketplace copy (ML/OLX).
	Description string
	ListingText string
	// OLX marketplace attributes (monitors / displays).
	ScreenType         string
	MaxResolution      string
	RefreshRate        string
	ItemCondition      string
	FeatCurved         bool
	FeatIncludesBox    bool
	FeatDisplayPort    bool
	FeatHDR            bool
	FeatWidescreen     bool
	FeatIncludesCables bool
	FeatAudio          bool
	FeatHDMI           bool
	FeatUltrawide      bool
	// OlxFreeShipping: offer "Entregar grátis pela OLX" on the listing.
	OlxFreeShipping bool
	// ShopVisible: when true (and photo+stock), product appears on the public catalog.
	ShopVisible bool
	CreatedAt   time.Time
}
