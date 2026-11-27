package store

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/puppe1990/leilao-erp/internal/models"
)

// ProductMediaInput creates a photo or video link for a product.
type ProductMediaInput struct {
	Kind      string // photo | video
	URL       string
	SortOrder int
}

// AddProductMedia attaches a photo or video URL/path to a product.
func (s *SQLiteStore) AddProductMedia(productID int64, in ProductMediaInput) (int64, error) {
	kind := strings.TrimSpace(strings.ToLower(in.Kind))
	if kind != "photo" && kind != "video" {
		return 0, fmt.Errorf("%w: kind must be photo or video", ErrInvalidInput)
	}
	url := strings.TrimSpace(in.URL)
	if url == "" {
		return 0, fmt.Errorf("%w: media url required", ErrInvalidInput)
	}
	var exists int
	err := s.db.QueryRow(`SELECT 1 FROM products WHERE id = ?`, productID).Scan(&exists)
	if err == sql.ErrNoRows {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("find product: %w", err)
	}
	res, err := s.db.Exec(
		`INSERT INTO product_media (product_id, kind, url, sort_order) VALUES (?, ?, ?, ?)`,
		productID, kind, url, in.SortOrder,
	)
	if err != nil {
		return 0, fmt.Errorf("insert product media: %w", err)
	}
	return res.LastInsertId()
}

// ListProductMedia returns media for a product ordered by sort_order, id.
func (s *SQLiteStore) ListProductMedia(productID int64) ([]models.ProductMedia, error) {
	rows, err := s.db.Query(
		`SELECT id, product_id, kind, url, sort_order, created_at
		 FROM product_media WHERE product_id = ?
		 ORDER BY sort_order ASC, id ASC`,
		productID,
	)
	if err != nil {
		return nil, fmt.Errorf("list product media: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.ProductMedia
	for rows.Next() {
		var m models.ProductMedia
		if err := rows.Scan(&m.ID, &m.ProductID, &m.Kind, &m.URL, &m.SortOrder, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan product media: %w", err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// DeleteProductMedia removes one media row by id.
func (s *SQLiteStore) DeleteProductMedia(mediaID int64) error {
	res, err := s.db.Exec(`DELETE FROM product_media WHERE id = ?`, mediaID)
	if err != nil {
		return fmt.Errorf("delete product media: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ReorderProductMedia sets sort_order 0..n-1 from orderedIDs.
// orderedIDs must be a permutation of this product's media.
func (s *SQLiteStore) ReorderProductMedia(productID int64, orderedIDs []int64) error {
	current, err := s.ListProductMedia(productID)
	if err != nil {
		return err
	}
	if err := validateMediaPermutation(current, orderedIDs); err != nil {
		return err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin media reorder: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	for i, id := range orderedIDs {
		res, err := tx.Exec(
			`UPDATE product_media SET sort_order = ? WHERE id = ? AND product_id = ?`,
			i, id, productID,
		)
		if err != nil {
			return fmt.Errorf("update media %d sort_order: %w", id, err)
		}
		n, _ := res.RowsAffected()
		if n != 1 {
			return fmt.Errorf("%w: media %d not updated", ErrInvalidInput, id)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit media reorder: %w", err)
	}
	return nil
}

func validateMediaPermutation(current []models.ProductMedia, orderedIDs []int64) error {
	if len(orderedIDs) != len(current) {
		return fmt.Errorf("%w: expected %d media ids, got %d", ErrInvalidInput, len(current), len(orderedIDs))
	}
	have := make(map[int64]struct{}, len(current))
	for _, m := range current {
		have[m.ID] = struct{}{}
	}
	seen := make(map[int64]struct{}, len(orderedIDs))
	for _, id := range orderedIDs {
		if _, ok := have[id]; !ok {
			return fmt.Errorf("%w: media %d does not belong to this product", ErrInvalidInput, id)
		}
		if _, dup := seen[id]; dup {
			return fmt.Errorf("%w: media %d repeated", ErrInvalidInput, id)
		}
		seen[id] = struct{}{}
	}
	return nil
}

// FindProductMedia returns one media row.
func (s *SQLiteStore) FindProductMedia(mediaID int64) (models.ProductMedia, error) {
	var m models.ProductMedia
	err := s.db.QueryRow(
		`SELECT id, product_id, kind, url, sort_order, created_at FROM product_media WHERE id = ?`,
		mediaID,
	).Scan(&m.ID, &m.ProductID, &m.Kind, &m.URL, &m.SortOrder, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return models.ProductMedia{}, ErrNotFound
	}
	if err != nil {
		return models.ProductMedia{}, err
	}
	return m, nil
}
