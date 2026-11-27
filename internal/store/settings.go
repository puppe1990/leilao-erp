package store

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/puppe1990/leilao-erp/internal/models"
)

const (
	settingCompanyName    = "company_name"
	settingCompanyCNPJ    = "company_cnpj"
	settingWhatsAppPhone  = "whatsapp_phone"
	settingShopLotEnabled = "shop_lot_enabled"
	settingShopShowQty    = "shop_show_qty"
)

// GetSetting returns a setting value (empty string if missing).
func (s *SQLiteStore) GetSetting(key string) (string, error) {
	var value string
	err := s.db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get setting %q: %w", key, err)
	}
	return value, nil
}

// SetSetting upserts a setting value.
func (s *SQLiteStore) SetSetting(key, value string) error {
	_, err := s.db.Exec(`
		INSERT INTO settings (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP
	`, key, value)
	if err != nil {
		return fmt.Errorf("set setting %q: %w", key, err)
	}
	return nil
}

// CompanyName returns the configured company display name (may be empty).
func (s *SQLiteStore) CompanyName() (string, error) {
	return s.GetSetting(settingCompanyName)
}

// SetCompanyName stores the company display name.
func (s *SQLiteStore) SetCompanyName(name string) error {
	return s.SetSetting(settingCompanyName, strings.TrimSpace(name))
}

// CompanyCNPJ returns the configured company CNPJ (may be empty).
func (s *SQLiteStore) CompanyCNPJ() (string, error) {
	return s.GetSetting(settingCompanyCNPJ)
}

// SetCompanyCNPJ stores the company CNPJ as typed (trimmed).
func (s *SQLiteStore) SetCompanyCNPJ(cnpj string) error {
	return s.SetSetting(settingCompanyCNPJ, strings.TrimSpace(cnpj))
}

// WhatsAppPhone returns the public shop WhatsApp number (may be empty).
func (s *SQLiteStore) WhatsAppPhone() (string, error) {
	return s.GetSetting(settingWhatsAppPhone)
}

// SetWhatsAppPhone stores the WhatsApp number used for public shop orders.
func (s *SQLiteStore) SetWhatsAppPhone(phone string) error {
	return s.SetSetting(settingWhatsAppPhone, strings.TrimSpace(phone))
}

// ShopLotEnabled is the public remaining-lot bundle. Missing setting means on.
func (s *SQLiteStore) ShopLotEnabled() (bool, error) {
	v, err := s.GetSetting(settingShopLotEnabled)
	if err != nil {
		return true, err
	}
	return settingBool(v, true), nil
}

// SetShopLotEnabled stores "1" or "0" for the public lot bundle.
func (s *SQLiteStore) SetShopLotEnabled(enabled bool) error {
	return s.SetSetting(settingShopLotEnabled, settingBoolValue(enabled))
}

func settingBool(v string, missingDefault bool) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "":
		return missingDefault
	case "0", "false", "off", "no":
		return false
	default:
		return true
	}
}

func settingBoolValue(on bool) string {
	if on {
		return "1"
	}
	return "0"
}

// ShopShowQty is whether catalog cards/PDP show remaining units. Missing setting means on.
func (s *SQLiteStore) ShopShowQty() (bool, error) {
	v, err := s.GetSetting(settingShopShowQty)
	if err != nil {
		return true, err
	}
	return settingBool(v, true), nil
}

// SetShopShowQty stores "1" or "0" for catalog quantity.
func (s *SQLiteStore) SetShopShowQty(enabled bool) error {
	return s.SetSetting(settingShopShowQty, settingBoolValue(enabled))
}

// FindUserByID loads a user by primary key.
func (s *SQLiteStore) FindUserByID(id int64) (models.User, error) {
	var u models.User
	err := s.db.QueryRow(
		`SELECT id, email, password_hash, created_at FROM users WHERE id = ?`,
		id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return models.User{}, fmt.Errorf("find user by id: %w", err)
	}
	return u, nil
}

// UpdateUserPassword replaces the password hash for a user.
func (s *SQLiteStore) UpdateUserPassword(userID int64, passwordHash string) error {
	res, err := s.db.Exec(
		`UPDATE users SET password_hash = ? WHERE id = ?`,
		passwordHash, userID,
	)
	if err != nil {
		return fmt.Errorf("update user password: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("update user password: user %d not found", userID)
	}
	return nil
}
