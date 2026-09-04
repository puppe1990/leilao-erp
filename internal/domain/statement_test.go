package domain_test

import (
	"testing"

	"github.com/puppe1990/leilao-erp/internal/domain"
)

func TestBuildStatement_RunningBalanceOldestFirst(t *testing.T) {
	lines := domain.BuildStatement(10000, []domain.LedgerEntry{
		{ID: 2, OccurredAt: "2026-01-02T12:00:00Z", Description: "Uber", Direction: "out", AmountCents: 2000},
		{ID: 1, OccurredAt: "2026-01-01T12:00:00Z", Description: "Venda", Direction: "in", AmountCents: 5000},
	})
	if len(lines) != 3 {
		t.Fatalf("lines=%d want 3", len(lines))
	}
	if lines[0].Kind != "opening" || lines[0].BalanceCents != 10000 {
		t.Fatalf("opening=%+v", lines[0])
	}
	if lines[1].CreditCents != 5000 || lines[1].BalanceCents != 15000 || lines[1].Description != "Venda" {
		t.Fatalf("credit=%+v", lines[1])
	}
	if lines[2].DebitCents != 2000 || lines[2].BalanceCents != 13000 {
		t.Fatalf("debit=%+v", lines[2])
	}
}

func TestBuildStatement_ZeroOpeningStillListed(t *testing.T) {
	lines := domain.BuildStatement(0, nil)
	if len(lines) != 1 || lines[0].Kind != "opening" || lines[0].BalanceCents != 0 {
		t.Fatalf("got %+v", lines)
	}
}
