package models

import "time"

// Client is a buyer/contact in the auction resale business.
type Client struct {
	ID        int64
	Name      string
	Phone     string
	Email     string
	Document  string // CPF/CNPJ free text
	Notes     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
