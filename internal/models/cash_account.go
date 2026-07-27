package models

import "time"

type CashAccount struct {
	ID                  int64
	Name                string
	Kind                string
	OpeningBalanceCents int64
	CreatedAt           time.Time
}
