package store

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"github.com/puppe1990/leilao-erp/internal/models"
)

const (
	FindSourceOLX      = "olx"
	FindSourceFacebook = "facebook"
	FindSourceKwara    = "kwara"
)

// FindInput creates or updates a marketplace find.
type FindInput struct {
	Name      string
	URL       string
	Source    string
	BuyCents  int64
	SellCents int64
}

func normalizeFindInput(in FindInput) (FindInput, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return in, fmt.Errorf("%w: name required", ErrInvalidInput)
	}
	in.URL = strings.TrimSpace(in.URL)
	parsed, err := url.Parse(in.URL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return in, fmt.Errorf("%w: url must be http(s), got %q", ErrInvalidInput, in.URL)
	}
	in.Source = strings.ToLower(strings.TrimSpace(in.Source))
	switch in.Source {
	case FindSourceOLX, FindSourceFacebook, FindSourceKwara:
	default:
		return in, fmt.Errorf("%w: source %q must be olx, facebook or kwara", ErrInvalidInput, in.Source)
	}
	if in.BuyCents < 0 {
		return in, fmt.Errorf("%w: buy_cents %d must be >= 0", ErrInvalidInput, in.BuyCents)
	}
	if in.SellCents < 0 {
		return in, fmt.Errorf("%w: sell_cents %d must be >= 0", ErrInvalidInput, in.SellCents)
	}
	return in, nil
}

// CreateFind inserts a find and returns its id.
func (s *SQLiteStore) CreateFind(in FindInput) (int64, error) {
	in, err := normalizeFindInput(in)
	if err != nil {
		return 0, err
	}
	res, err := s.db.Exec(
		`INSERT INTO finds (name, url, source, buy_cents, sell_cents) VALUES (?, ?, ?, ?, ?)`,
		in.Name, in.URL, in.Source, in.BuyCents, in.SellCents,
	)
	if err != nil {
		return 0, fmt.Errorf("insert find: %w", err)
	}
	return res.LastInsertId()
}

// UpdateFind updates mutable find fields.
func (s *SQLiteStore) UpdateFind(id int64, in FindInput) error {
	in, err := normalizeFindInput(in)
	if err != nil {
		return err
	}
	res, err := s.db.Exec(
		`UPDATE finds SET name = ?, url = ?, source = ?, buy_cents = ?, sell_cents = ?,
		 updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		in.Name, in.URL, in.Source, in.BuyCents, in.SellCents, id,
	)
	if err != nil {
		return fmt.Errorf("update find: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteFind removes a find by id.
func (s *SQLiteStore) DeleteFind(id int64) error {
	res, err := s.db.Exec(`DELETE FROM finds WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete find: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// GetFind returns one find.
func (s *SQLiteStore) GetFind(id int64) (models.Find, error) {
	row := s.db.QueryRow(
		`SELECT id, name, url, source, buy_cents, sell_cents, created_at, updated_at
		 FROM finds WHERE id = ?`, id,
	)
	f, err := scanFind(row)
	if err == sql.ErrNoRows {
		return models.Find{}, ErrNotFound
	}
	if err != nil {
		return models.Find{}, fmt.Errorf("get find %d: %w", id, err)
	}
	return f, nil
}

// ListFinds returns finds newest first. source empty lists all; otherwise filters.
func (s *SQLiteStore) ListFinds(source string) ([]models.Find, error) {
	source = strings.ToLower(strings.TrimSpace(source))
	var (
		rows *sql.Rows
		err  error
	)
	if source == "" {
		rows, err = s.db.Query(
			`SELECT id, name, url, source, buy_cents, sell_cents, created_at, updated_at
			 FROM finds ORDER BY id DESC`,
		)
	} else {
		rows, err = s.db.Query(
			`SELECT id, name, url, source, buy_cents, sell_cents, created_at, updated_at
			 FROM finds WHERE source = ? ORDER BY id DESC`,
			source,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("list finds: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := make([]models.Find, 0)
	for rows.Next() {
		f, err := scanFind(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

type findScanner interface {
	Scan(dest ...any) error
}

func scanFind(row findScanner) (models.Find, error) {
	var f models.Find
	err := row.Scan(&f.ID, &f.Name, &f.URL, &f.Source, &f.BuyCents, &f.SellCents, &f.CreatedAt, &f.UpdatedAt)
	return f, err
}
