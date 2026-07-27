package store

import (
	"database/sql"
	"fmt"

	"github.com/puppe1990/leilao-erp/internal/models"
)

func (s *SQLiteStore) FindLot(id int64) (models.Lot, error) {
	var lot models.Lot
	var auctionSource, notes sql.NullString
	err := s.db.QueryRow(
		`SELECT id, name, auction_source, purchased_at, status, notes, created_at
		 FROM lots WHERE id = ?`,
		id,
	).Scan(
		&lot.ID, &lot.Name, &auctionSource, &lot.PurchasedAt,
		&lot.Status, &notes, &lot.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return models.Lot{}, ErrNotFound
	}
	if err != nil {
		return models.Lot{}, fmt.Errorf("find lot: %w", err)
	}
	if auctionSource.Valid {
		lot.AuctionSource = &auctionSource.String
	}
	if notes.Valid {
		lot.Notes = &notes.String
	}
	return lot, nil
}
