package domain_test

import (
	"testing"

	"github.com/puppe1990/leilao-erp/internal/domain"
)

func TestDigitsOnly(t *testing.T) {
	got := domain.DigitsOnly("11.444.777/0001-61")
	if got != "11444777000161" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatCNPJ(t *testing.T) {
	got, err := domain.FormatCNPJ("11444777000161")
	if err != nil {
		t.Fatal(err)
	}
	if got != "11.444.777/0001-61" {
		t.Fatalf("got %q", got)
	}
	if _, err := domain.FormatCNPJ("123"); err == nil {
		t.Fatal("expected error for short CNPJ")
	}
}

func TestValidCNPJ(t *testing.T) {
	if !domain.ValidCNPJ("11.444.777/0001-61") {
		t.Fatal("want valid CNPJ")
	}
	if domain.ValidCNPJ("11.444.777/0001-00") {
		t.Fatal("want invalid check digits")
	}
	if domain.ValidCNPJ("00.000.000/0000-00") {
		t.Fatal("want reject repeated digits")
	}
}

func TestFormatDocument_CPFAndCNPJ(t *testing.T) {
	cpf, err := domain.FormatDocument("52998224725")
	if err != nil {
		t.Fatal(err)
	}
	if cpf != "529.982.247-25" {
		t.Fatalf("cpf = %q", cpf)
	}
	cnpj, err := domain.FormatDocument("11.444.777/0001-61")
	if err != nil {
		t.Fatal(err)
	}
	if cnpj != "11.444.777/0001-61" {
		t.Fatalf("cnpj = %q", cnpj)
	}
	if _, err := domain.FormatDocument("123"); err == nil {
		t.Fatal("expected error for short document")
	}
}

func TestValidCPF(t *testing.T) {
	if !domain.ValidCPF("529.982.247-25") {
		t.Fatal("want valid CPF")
	}
	if domain.ValidCPF("111.111.111-11") {
		t.Fatal("want reject repeated digits")
	}
}

func TestFormatBRDate(t *testing.T) {
	if got := domain.FormatBRDate("2026-07-22T12:00:00Z"); got != "22/07/2026" {
		t.Fatalf("got %q", got)
	}
	if got := domain.FormatBRDate("2026-07-22"); got != "22/07/2026" {
		t.Fatalf("got %q", got)
	}
	if got := domain.FormatBRDate(""); got != "" {
		t.Fatalf("empty = %q", got)
	}
}
