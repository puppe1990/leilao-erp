package store

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/puppe1990/leilao-erp/internal/models"
)

// ClientInput creates or updates a client.
type ClientInput struct {
	Name     string
	Phone    string
	Email    string
	Document string
	Type     string
	Address  models.Address
	Notes    string
}

func normalizeClientInput(in ClientInput) (ClientInput, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return in, fmt.Errorf("%w: name required", ErrInvalidInput)
	}
	in.Phone = strings.TrimSpace(in.Phone)
	in.Email = strings.TrimSpace(in.Email)
	in.Document = strings.TrimSpace(in.Document)
	if in.Type != "company" {
		in.Type = "person"
	}
	in.Address.CEP = strings.TrimSpace(in.Address.CEP)
	in.Address.Street = strings.TrimSpace(in.Address.Street)
	in.Address.Number = strings.TrimSpace(in.Address.Number)
	in.Address.Complement = strings.TrimSpace(in.Address.Complement)
	in.Address.Neighborhood = strings.TrimSpace(in.Address.Neighborhood)
	in.Address.City = strings.TrimSpace(in.Address.City)
	in.Address.State = strings.TrimSpace(in.Address.State)
	in.Notes = strings.TrimSpace(in.Notes)
	return in, nil
}

// CreateClient inserts a client and returns its id.
func (s *SQLiteStore) CreateClient(in ClientInput) (int64, error) {
	in, err := normalizeClientInput(in)
	if err != nil {
		return 0, err
	}
	a := in.Address
	res, err := s.db.Exec(
		`INSERT INTO clients (name, phone, email, document, client_type, cep, street, number, complement, neighborhood, city, state, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		in.Name, nullStr(in.Phone), nullStr(in.Email), nullStr(in.Document), in.Type,
		nullStr(a.CEP), nullStr(a.Street), nullStr(a.Number), nullStr(a.Complement),
		nullStr(a.Neighborhood), nullStr(a.City), nullStr(a.State), nullStr(in.Notes),
	)
	if err != nil {
		return 0, fmt.Errorf("insert client: %w", err)
	}
	return res.LastInsertId()
}

// UpdateClient updates mutable client fields.
func (s *SQLiteStore) UpdateClient(id int64, in ClientInput) error {
	in, err := normalizeClientInput(in)
	if err != nil {
		return err
	}
	a := in.Address
	res, err := s.db.Exec(
		`UPDATE clients SET name = ?, phone = ?, email = ?, document = ?, client_type = ?,
		 cep = ?, street = ?, number = ?, complement = ?, neighborhood = ?, city = ?, state = ?, notes = ?,
		 updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		in.Name, nullStr(in.Phone), nullStr(in.Email), nullStr(in.Document), in.Type,
		nullStr(a.CEP), nullStr(a.Street), nullStr(a.Number), nullStr(a.Complement),
		nullStr(a.Neighborhood), nullStr(a.City), nullStr(a.State), nullStr(in.Notes),
		id,
	)
	if err != nil {
		return fmt.Errorf("update client: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteClient removes a client by id.
func (s *SQLiteStore) DeleteClient(id int64) error {
	res, err := s.db.Exec(`DELETE FROM clients WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete client: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// FindClient returns one client.
func (s *SQLiteStore) FindClient(id int64) (models.Client, error) {
	var c models.Client
	var phone, email, document, notes sql.NullString
	var cep, street, number, complement, neighborhood, city, state sql.NullString
	err := s.db.QueryRow(
		`SELECT id, name, phone, email, document, client_type,
		 cep, street, number, complement, neighborhood, city, state,
		 notes, created_at, updated_at
		 FROM clients WHERE id = ?`, id,
	).Scan(&c.ID, &c.Name, &phone, &email, &document, &c.Type,
		&cep, &street, &number, &complement, &neighborhood, &city, &state,
		&notes, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return models.Client{}, ErrNotFound
	}
	if err != nil {
		return models.Client{}, fmt.Errorf("find client: %w", err)
	}
	c.Phone = phone.String
	c.Email = email.String
	c.Document = document.String
	c.Notes = notes.String
	c.Address = models.Address{
		CEP:          cep.String,
		Street:       street.String,
		Number:       number.String,
		Complement:   complement.String,
		Neighborhood: neighborhood.String,
		City:         city.String,
		State:        state.String,
	}
	return c, nil
}

// ListClients returns all clients ordered by name.
func (s *SQLiteStore) ListClients() ([]models.Client, error) {
	return s.SearchClients("")
}

// SearchClients filters by name/phone/email/document (case-insensitive substring).
func (s *SQLiteStore) SearchClients(query string) ([]models.Client, error) {
	q := strings.TrimSpace(query)
	var rows *sql.Rows
	var err error
	selectClause := `SELECT id, name, phone, email, document, client_type,
		 cep, street, number, complement, neighborhood, city, state,
		 notes, created_at, updated_at
		 FROM clients`
	if q == "" {
		rows, err = s.db.Query(selectClause + ` ORDER BY name COLLATE NOCASE, id`)
	} else {
		like := "%" + q + "%"
		rows, err = s.db.Query(selectClause+`
			 WHERE name LIKE ? COLLATE NOCASE
			    OR IFNULL(phone,'') LIKE ?
			    OR IFNULL(email,'') LIKE ? COLLATE NOCASE
			    OR IFNULL(document,'') LIKE ?
			 ORDER BY name COLLATE NOCASE, id`,
			like, like, like, like,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("list clients: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanClients(rows)
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func scanClients(rows *sql.Rows) ([]models.Client, error) {
	var out []models.Client
	for rows.Next() {
		var c models.Client
		var phone, email, document, notes sql.NullString
		var cep, street, number, complement, neighborhood, city, state sql.NullString
		if err := rows.Scan(
			&c.ID, &c.Name, &phone, &email, &document, &c.Type,
			&cep, &street, &number, &complement, &neighborhood, &city, &state,
			&notes, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan client: %w", err)
		}
		c.Phone = phone.String
		c.Email = email.String
		c.Document = document.String
		c.Notes = notes.String
		c.Address = models.Address{
			CEP:          cep.String,
			Street:       street.String,
			Number:       number.String,
			Complement:   complement.String,
			Neighborhood: neighborhood.String,
			City:         city.String,
			State:        state.String,
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
