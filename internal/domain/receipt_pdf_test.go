package domain_test

import (
	"bytes"
	"testing"

	"github.com/puppe1990/leilao-erp/internal/domain"
)

func TestBuildReceiptPDF_ContainsSellerBuyerAndItems(t *testing.T) {
	pdf, err := domain.BuildReceiptPDF(validReceipt())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		t.Fatalf("not a PDF: %q", pdf[:min(12, len(pdf))])
	}
	body := string(pdf)
	for _, want := range []string{
		"RECIBO DE VENDA",
		"11.444.777/0001-61",
		"Puppe Leiloes",
		"Maria Silva",
		"529.982.247-25",
		"Monitor Dell P1913Sb 19'",
		"Cabo HDMI",
		"R$ 199,00",
		"22/07/2026",
		"documento fiscal",
		"Admin",
		"Courier",
	} {
		if !bytes.Contains(pdf, []byte(want)) {
			t.Errorf("PDF missing %q\n--- excerpt ---\n%s", want, body)
		}
	}
	if !bytes.Contains(pdf, []byte(" rg")) {
		t.Fatal("PDF missing fill color operators (design uses AuctionHQ palette)")
	}
}

func TestBuildReceiptPDF_RejectsInvalid(t *testing.T) {
	r := validReceipt()
	r.Seller.Document = ""
	if _, err := domain.BuildReceiptPDF(r); err == nil {
		t.Fatal("expected validation error")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
