package db

import (
	"fmt"

	"github.com/puppe1990/leilao-erp/internal/models"
	"github.com/puppe1990/leilao-erp/internal/store"
)

const (
	seedPIXAccountName  = "PIX principal"
	seedMonitorsLotName = "Monitores — leilão Jul/2026"
	seedMonitorTitle    = "Monitor"
)

// RunSeeds populates demo data. Safe to run multiple times.
//
// - Ensures cash account "PIX principal" (kind=pix, opening 0)
// - Ensures lot "Monitores — leilão Jul/2026" with 22 monitors + arremate + transporte
// - Admin user is seeded separately in development via store.NewSQLiteStore
// Model/sale price (ex. P2219H) is set manually per unit in Estoque — not bulk-seeded.
func RunSeeds(s store.Store) error {
	// cais:recurring-seeds
	// cais:seeds
	// Optional scaffold demo contact once (skip if any contact exists).
	if n, err := s.CountContacts(); err != nil {
		return err
	} else if n == 0 {
		if _, err := s.InsertContact(models.Contact{
			Name:  "Demo",
			Email: "demo@example.com",
		}); err != nil {
			return err
		}
	}

	if err := ensurePIXAccount(s); err != nil {
		return err
	}
	if err := ensureMonitorsLot(s); err != nil {
		return err
	}
	if err := ensureMonitorsTransportCosts(s); err != nil {
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
		ItemTitle:   seedMonitorTitle,
		ItemQty:     22,
		Costs: []store.CostInput{
			{Label: "Arremate", AmountCents: 60300, AlreadyPaid: true},
			{Label: "Uber — transporte monitores", AmountCents: 1435, AlreadyPaid: true},
			{Label: "Uber — transporte monitores", AmountCents: 1452, AlreadyPaid: true},
			{Label: "Lalamove — transporte monitores", AmountCents: 5476, AlreadyPaid: true},
		},
		CashAccountID: accountID,
		PaidAt:        "2026-07-20T12:00:00Z",
	})
	if err != nil {
		return fmt.Errorf("create monitors lot: %w", err)
	}
	return nil
}

// ensureMonitorsTransportCosts adds freight costs if the monitores lot exists without them.
func ensureMonitorsTransportCosts(s store.Store) error {
	lots, err := s.ListLots()
	if err != nil {
		return err
	}
	var lotID int64
	for _, lot := range lots {
		if lot.Name == seedMonitorsLotName {
			lotID = lot.ID
			break
		}
	}
	if lotID == 0 {
		return nil
	}

	accounts, err := s.ListCashAccounts()
	if err != nil {
		return err
	}
	var accountID int64
	for _, a := range accounts {
		if a.Name == seedPIXAccountName {
			accountID = a.ID
			break
		}
	}
	if accountID == 0 && len(accounts) > 0 {
		accountID = accounts[0].ID
	}
	if accountID == 0 {
		return nil
	}

	existing, err := s.ListPurchaseCostsByLot(lotID)
	if err != nil {
		return err
	}
	has := func(label string, amount int64) bool {
		for _, e := range existing {
			if e.Label == label && e.AmountCents == amount {
				return true
			}
		}
		return false
	}

	costs := []store.CostInput{
		{Label: "Uber — transporte monitores", AmountCents: 1435, AlreadyPaid: true},
		{Label: "Uber — transporte monitores", AmountCents: 1452, AlreadyPaid: true},
		{Label: "Lalamove — transporte monitores", AmountCents: 5476, AlreadyPaid: true},
	}
	for _, c := range costs {
		if has(c.Label, c.AmountCents) {
			continue
		}
		if err := s.AddPurchaseCost(lotID, c, accountID, "2026-07-20T15:00:00Z"); err != nil {
			return fmt.Errorf("seed transport cost %s: %w", c.Label, err)
		}
	}
	return nil
}
