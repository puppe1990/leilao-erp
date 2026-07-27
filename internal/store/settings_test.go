package store_test

import (
	"path/filepath"
	"testing"

	"github.com/puppe1990/cais/pkg/cais/session"

	"github.com/puppe1990/leilao-erp/internal/store"
)

func TestSettings_CompanyName(t *testing.T) {
	st, err := store.NewSQLiteStore(filepath.Join(t.TempDir(), "s.db"), "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })

	name, err := st.CompanyName()
	if err != nil {
		t.Fatal(err)
	}
	if name != "" {
		t.Fatalf("want empty default, got %q", name)
	}

	if err := st.SetCompanyName("  Puppe Leilões  "); err != nil {
		t.Fatal(err)
	}
	name, err = st.CompanyName()
	if err != nil {
		t.Fatal(err)
	}
	if name != "Puppe Leilões" {
		t.Fatalf("got %q", name)
	}

	// upsert
	if err := st.SetCompanyName("Leilão ERP"); err != nil {
		t.Fatal(err)
	}
	name, _ = st.CompanyName()
	if name != "Leilão ERP" {
		t.Fatalf("got %q", name)
	}
}

func TestUpdateUserPassword(t *testing.T) {
	st, err := store.NewSQLiteStore(filepath.Join(t.TempDir(), "s.db"), "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })

	user, err := st.FindUserByEmail("admin@leilao.local")
	if err != nil {
		t.Fatal(err)
	}
	if !session.VerifyPassword(user.PasswordHash, "change-me-now") {
		t.Fatal("seed password mismatch")
	}

	hash, err := session.HashPassword("nova-senha-forte")
	if err != nil {
		t.Fatal(err)
	}
	if err := st.UpdateUserPassword(user.ID, hash); err != nil {
		t.Fatal(err)
	}

	updated, err := st.FindUserByID(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !session.VerifyPassword(updated.PasswordHash, "nova-senha-forte") {
		t.Fatal("new password not applied")
	}
	if session.VerifyPassword(updated.PasswordHash, "change-me-now") {
		t.Fatal("old password still works")
	}
}
