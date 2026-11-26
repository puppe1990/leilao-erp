package handlers

import (
	"path/filepath"
	"testing"

	"github.com/puppe1990/leilao-erp/internal/store"
)

func TestCompanyName_emptyFallsBackToPixEletro(t *testing.T) {
	st, err := store.NewSQLiteStore(filepath.Join(t.TempDir(), "p.db"), "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })

	if got := companyName(st); got != "Pix Eletro" {
		t.Fatalf("got %q, want Pix Eletro", got)
	}
}
