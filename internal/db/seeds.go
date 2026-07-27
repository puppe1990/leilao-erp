package db

import (
	"fmt"

	"github.com/puppe1990/leilao-erp/internal/models"
	"github.com/puppe1990/leilao-erp/internal/store"
)

const (
	seedPIXAccountName = "PIX principal"
	seedMonitorsLotName = "Monitores — leilão Jul/2026"
)

// RunSeeds populates demo data. Safe to run multiple times.
//
// - Ensures cash account "PIX principal" (kind=pix, opening 0)
// - Ensures lot "Monitores — leilão Jul/2026" with 22 monitors, Arremate 60300 already paid
// - Admin user is seeded separately in development via store.NewSQLiteStore
func RunSeeds(s store.Store) error {
	// cais:recurring-seeds
	// cais:seeds
	if _, err := s.InsertContact(models.Contact{
		Name:  "Demo",
		Email: "demo@example.com",
	}); err != nil {
		return err
	}

	if err := ensurePIXAccount(s); err != nil {
		return err
	}
	if err := ensureMonitorsLot(s); err != nil {
		return err
	}
	return nil
}

func ensurePIXAccount(s store.Store) error {
	accounts, err := s.ListCashAccounts()
	if err != nil {
		return fmt.Errorf("list cash accounts: %w", err)
	}
	for _, a := range accounts {
		if a.Name == seedPIXAccountName {
			return nil
		}
	}
	_, err = s.InsertCashAccount(seedPIXAccountName, "pix", 0)
	if err != nil {
		return fmt.Errorf("insert PIX principal: %w", err)
	}
	return nil
}

func ensureMonitorsLot(s store.Store) error {
	lots, err := s.ListLots()
	if err != nil {
		return fmt.Errorf("list lots: %w", err)
	}
	for _, lot := range lots {
		if lot.Name == seedMonitorsLotName {
			return nil
		}
	}

	accounts, err := s.ListCashAccounts()
	if err != nil {
		return fmt.Errorf("list cash accounts for lot: %w", err)
	}
	var accountID int64
	for _, a := range accounts {
		if a.Name == seedPIXAccountName {
			accountID = a.ID
			break
		}
	}
	if accountID == 0 {
		if len(accounts) == 0 {
			return fmt.Errorf("no cash account available for monitors lot seed")
		}
		accountID = accounts[0].ID
	}

	_, err = s.CreateLotPurchase(store.CreateLotInput{
		Name:        seedMonitorsLotName,
		PurchasedAt: "2026-07-20",
		ItemTitle:   "Monitor",
		ItemQty:     22,
		Costs: []store.CostInput{
			{Label: "Arremate", AmountCents: 60300, AlreadyPaid: true},
		},
		CashAccountID: accountID,
		PaidAt:        "2026-07-20T12:00:00Z",
	})
	if err != nil {
		return fmt.Errorf("create monitors lot: %w", err)
	}
	return nil
}
