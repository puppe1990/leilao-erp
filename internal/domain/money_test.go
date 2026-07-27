package domain_test

import (
	"testing"

	"github.com/puppe1990/leilao-erp/internal/domain"
)

func TestFormatBRL(t *testing.T) {
	if got := domain.FormatBRL(60300); got != "R$ 603,00" {
		t.Fatalf("got %q", got)
	}
	if got := domain.FormatBRL(2741); got != "R$ 27,41" {
		t.Fatalf("got %q", got)
	}
	if got := domain.FormatBRL(0); got != "R$ 0,00" {
		t.Fatalf("got %q", got)
	}
}

func TestParseBRLToCents(t *testing.T) {
	cents, err := domain.ParseBRLToCents("603,00")
	if err != nil || cents != 60300 {
		t.Fatalf("got %d %v", cents, err)
	}
	cents, err = domain.ParseBRLToCents("R$ 27,41")
	if err != nil || cents != 2741 {
		t.Fatalf("got %d %v", cents, err)
	}
	if _, err := domain.ParseBRLToCents(""); err == nil {
		t.Fatal("expected error")
	}
}

func TestSaleNet(t *testing.T) {
	if got := domain.SaleNet(18000, 3000, 2000); got != 13000 {
		t.Fatalf("got %d", got)
	}
}

func TestMargin(t *testing.T) {
	if got := domain.Margin(13000, 2741); got != 10259 {
		t.Fatalf("got %d", got)
	}
}
