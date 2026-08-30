package domain_test

import (
	"strings"
	"testing"

	"github.com/puppe1990/leilao-erp/internal/domain"
)

func validReceipt() domain.Receipt {
	return domain.Receipt{
		Number:   "42",
		IssuedAt: "2026-07-22T12:00:00Z",
		Channel:  "Direto",
		Seller: domain.ReceiptParty{
			Name:     "Puppe Leiloes",
			Document: "11.444.777/0001-61",
		},
		Buyer: domain.ReceiptParty{
			Name:     "Maria Silva",
			Document: "529.982.247-25",
			Phone:    "11 99999-0000",
			Email:    "maria@example.com",
		},
		Items: []domain.ReceiptItem{
			{Description: "Monitor Dell P1913Sb 19'"},
			{Description: "Cabo HDMI"},
		},
		TotalCents: 19900,
	}
}

func TestValidateReceipt_OK(t *testing.T) {
	if err := domain.ValidateReceipt(validReceipt()); err != nil {
		t.Fatal(err)
	}
}

func TestValidateReceipt_RequiresSellerCNPJ(t *testing.T) {
	r := validReceipt()
	r.Seller.Document = ""
	err := domain.ValidateReceipt(r)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "CNPJ") {
		t.Fatalf("error %q should mention CNPJ", err)
	}
}

func TestValidateReceipt_RejectsInvalidSellerCNPJ(t *testing.T) {
	r := validReceipt()
	r.Seller.Document = "11.444.777/0001-00"
	if err := domain.ValidateReceipt(r); err == nil {
		t.Fatal("expected invalid CNPJ")
	}
}

func TestValidateReceipt_RequiresBuyerNameAndDocument(t *testing.T) {
	r := validReceipt()
	r.Buyer.Name = ""
	if err := domain.ValidateReceipt(r); err == nil {
		t.Fatal("expected buyer name error")
	}
	r = validReceipt()
	r.Buyer.Document = "123"
	if err := domain.ValidateReceipt(r); err == nil {
		t.Fatal("expected buyer document error")
	}
}

func TestValidateReceipt_RequiresItemsAndTotal(t *testing.T) {
	r := validReceipt()
	r.Items = nil
	if err := domain.ValidateReceipt(r); err == nil {
		t.Fatal("expected items error")
	}
	r = validReceipt()
	r.TotalCents = 0
	if err := domain.ValidateReceipt(r); err == nil {
		t.Fatal("expected total error")
	}
}
