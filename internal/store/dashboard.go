package store

import (
	"fmt"
	"time"
)

// AccountBalance is a cash account with its computed balance in cents.
type AccountBalance struct {
	ID    int64
	Name  string
	Cents int64
}

// DashboardSummary holds aggregate figures for the home dashboard.
type DashboardSummary struct {
	CashBalances         []AccountBalance
	TotalCashCents       int64
	OpenPayablesCents    int64
	OpenReceivablesCents int64
	MonthProfitCents     int64 // sum(net - unit_cost_at_sale) for non-cancelled sales in current calendar month (sold_at)
	OverduePayables      int
	OverdueReceivables   int
	LotCount             int
}

// DashboardSummary computes cash balances, open AP/AR, month profit, overdue counts, and lot count.
// Overdue = status open with due_on < today (ISO YYYY-MM-DD string compare).
func (s *SQLiteStore) DashboardSummary() (DashboardSummary, error) {
	var out DashboardSummary

	rows, err := s.db.Query(
		`SELECT id, name, opening_balance_cents FROM cash_accounts ORDER BY id`,
	)
	if err != nil {
		return out, fmt.Errorf("list cash accounts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	type acct struct {
		id      int64
		name    string
		opening int64
	}
	var accounts []acct
	for rows.Next() {
		var a acct
		if err := rows.Scan(&a.id, &a.name, &a.opening); err != nil {
			return out, fmt.Errorf("scan cash account: %w", err)
		}
		accounts = append(accounts, a)
	}
	if err := rows.Err(); err != nil {
		return out, err
	}
	if err := rows.Close(); err != nil {
		return out, err
	}

	out.CashBalances = make([]AccountBalance, 0, len(accounts))
	for _, a := range accounts {
		var net int64
		err := s.db.QueryRow(
			`SELECT COALESCE(SUM(
				CASE direction
					WHEN 'in' THEN amount_cents
					WHEN 'out' THEN -amount_cents
					ELSE 0
				END
			), 0) FROM cash_entries WHERE account_id = ?`,
			a.id,
		).Scan(&net)
		if err != nil {
			return out, fmt.Errorf("cash entries for account %d: %w", a.id, err)
		}
		bal := a.opening + net
		out.CashBalances = append(out.CashBalances, AccountBalance{
			ID:    a.id,
			Name:  a.name,
			Cents: bal,
		})
		out.TotalCashCents += bal
	}

	if err := s.db.QueryRow(
		`SELECT COALESCE(SUM(amount_cents), 0) FROM payables WHERE status = 'open'`,
	).Scan(&out.OpenPayablesCents); err != nil {
		return out, fmt.Errorf("open payables: %w", err)
	}

	if err := s.db.QueryRow(
		`SELECT COALESCE(SUM(amount_cents), 0) FROM receivables WHERE status = 'open'`,
	).Scan(&out.OpenReceivablesCents); err != nil {
		return out, fmt.Errorf("open receivables: %w", err)
	}

	// Current calendar month (UTC) for sold_at filter. ISO prefixes compare correctly.
	monthPrefix := time.Now().UTC().Format("2006-01")
	if err := s.db.QueryRow(
		`SELECT COALESCE(SUM(net_cents - unit_cost_cents_at_sale), 0)
		 FROM sales
		 WHERE payment_status != 'cancelled'
		   AND substr(sold_at, 1, 7) = ?`,
		monthPrefix,
	).Scan(&out.MonthProfitCents); err != nil {
		return out, fmt.Errorf("month profit: %w", err)
	}

	today := time.Now().UTC().Format("2006-01-02")
	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM payables WHERE status = 'open' AND due_on < ?`,
		today,
	).Scan(&out.OverduePayables); err != nil {
		return out, fmt.Errorf("overdue payables: %w", err)
	}
	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM receivables WHERE status = 'open' AND due_on < ?`,
		today,
	).Scan(&out.OverdueReceivables); err != nil {
		return out, fmt.Errorf("overdue receivables: %w", err)
	}

	if err := s.db.QueryRow(`SELECT COUNT(*) FROM lots`).Scan(&out.LotCount); err != nil {
		return out, fmt.Errorf("lot count: %w", err)
	}

	return out, nil
}
