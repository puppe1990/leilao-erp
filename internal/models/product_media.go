package models

import "time"

// ProductMedia is a photo or video attached to a catalog product.
type ProductMedia struct {
	ID        int64
	ProductID int64
	Kind      string // photo | video
	URL       string // /static/... or https://...
	SortOrder int
	CreatedAt time.Time
}
