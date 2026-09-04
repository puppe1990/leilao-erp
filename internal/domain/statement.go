package domain

import "sort"

// LedgerEntry is one cash movement used to build a checking-account statement.
type LedgerEntry struct {
	ID          int64
	OccurredAt  string
	Description string
	Direction   string // in | out
	AmountCents int64
}

// StatementLine is one row of a conta-corrente extrato (oldest first).
type StatementLine struct {
	Kind         string // opening | movement
	EntryID      int64
	OccurredAt   string
	Description  string
	CreditCents  int64
	DebitCents   int64
	BalanceCents int64
}

// BuildStatement starts from opening balance and applies entries in time order.
func BuildStatement(openingCents int64, entries []LedgerEntry) []StatementLine {
	ordered := append([]LedgerEntry(nil), entries...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].OccurredAt == ordered[j].OccurredAt {
			return ordered[i].ID < ordered[j].ID
		}
		return ordered[i].OccurredAt < ordered[j].OccurredAt
	})

	out := make([]StatementLine, 0, len(ordered)+1)
	bal := openingCents
	open := StatementLine{Kind: "opening", Description: "Saldo inicial", BalanceCents: bal}
	if openingCents > 0 {
		open.CreditCents = openingCents
	}
	if openingCents < 0 {
		open.DebitCents = -openingCents
	}
	out = append(out, open)

	for _, e := range ordered {
		line := StatementLine{
			Kind:        "movement",
			EntryID:     e.ID,
			OccurredAt:  e.OccurredAt,
			Description: e.Description,
		}
		switch e.Direction {
		case "in":
			line.CreditCents = e.AmountCents
			bal += e.AmountCents
		case "out":
			line.DebitCents = e.AmountCents
			bal -= e.AmountCents
		}
		line.BalanceCents = bal
		out = append(out, line)
	}
	return out
}
