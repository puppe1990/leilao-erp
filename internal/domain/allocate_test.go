package domain_test

import (
	"testing"

	"github.com/puppe1990/leilao-erp/internal/domain"
)

func TestAllocateUnitCosts_Monitors603(t *testing.T) {
	costs := domain.AllocateUnitCosts(60300, 22)
	if len(costs) != 22 {
		t.Fatalf("len=%d", len(costs))
	}
	var sum int64
	var high, low int
	for _, c := range costs {
		sum += c
		switch c {
		case 2741:
			high++
		case 2740:
			low++
		default:
			t.Fatalf("unexpected unit cost %d", c)
		}
	}
	if sum != 60300 {
		t.Fatalf("sum=%d want 60300", sum)
	}
	if high != 20 || low != 2 {
		t.Fatalf("high=%d low=%d", high, low)
	}
}

func TestAllocateUnitCosts_ZeroQty(t *testing.T) {
	if got := domain.AllocateUnitCosts(100, 0); got != nil {
		t.Fatalf("want nil, got %v", got)
	}
}

func TestAllocateUnitCosts_ExactDivision(t *testing.T) {
	costs := domain.AllocateUnitCosts(1000, 10)
	for _, c := range costs {
		if c != 100 {
			t.Fatalf("got %d", c)
		}
	}
}
