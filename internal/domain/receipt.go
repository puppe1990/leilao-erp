package domain

import (
	"fmt"
	"strings"
)

// ReceiptParty is the seller (empresa) or buyer (cliente) on a sale receipt.
type ReceiptParty struct {
	Name     string
	Document string // CNPJ for seller; CPF or CNPJ for buyer
	Phone    string
	Email    string
}

// ReceiptItem is one line printed on the receipt.
type ReceiptItem struct {
	Description string
}

// Receipt is the data needed to emit a simple (non-fiscal) sale receipt PDF.
type Receipt struct {
	Number     string
	IssuedAt   string
	Channel    string
	Seller     ReceiptParty
	Buyer      ReceiptParty
	Items      []ReceiptItem
	TotalCents int64
}

// ValidateReceipt checks required seller CNPJ, buyer identity, items and total.
func ValidateReceipt(r Receipt) error {
	if strings.TrimSpace(r.Seller.Name) == "" {
		return fmt.Errorf("seller name is empty, want a non-empty company name")
	}
	if !ValidCNPJ(r.Seller.Document) {
		return fmt.Errorf("seller CNPJ %q is invalid, want 14-digit CNPJ with check digits", r.Seller.Document)
	}
	if strings.TrimSpace(r.Buyer.Name) == "" {
		return fmt.Errorf("buyer name is empty, want the client name")
	}
	if _, err := FormatDocument(r.Buyer.Document); err != nil {
		return fmt.Errorf("buyer document: %w", err)
	}
	if len(r.Items) == 0 {
		return fmt.Errorf("items is empty, want at least 1 sold item")
	}
	for i, it := range r.Items {
		if strings.TrimSpace(it.Description) == "" {
			return fmt.Errorf("items[%d] description is empty", i)
		}
	}
	if r.TotalCents <= 0 {
		return fmt.Errorf("total %d cents is not positive, want gross > 0", r.TotalCents)
	}
	return nil
}
