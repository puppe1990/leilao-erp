package models

import "time"

// Address holds Brazilian street-address fields used by Client and Lot.
type Address struct {
	CEP          string
	Street       string
	Number       string
	Complement   string
	Neighborhood string
	City         string
	State        string
}

// Client is a buyer/contact in the auction resale business.
type Client struct {
	ID        int64
	Name      string
	Phone     string
	Email     string
	Document  string // CPF/CNPJ free text
	Type      string // "person" or "company"
	Address   Address
	Notes     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
