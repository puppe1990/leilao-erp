package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/puppe1990/leilao-erp/internal/store"
)

type dumpSupplier struct {
	Name          string `json:"name"`
	WhatsApp      string `json:"whatsapp"`
	Notes         string `json:"notes"`
	MinOrderCents int64  `json:"min_order_cents"`
	Messages      []struct {
		ID     string `json:"id"`
		FromMe bool   `json:"from_me"`
		Body   string `json:"body"`
		TS     int64  `json:"ts"`
	} `json:"messages"`
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	dumpPath := "/tmp/waha-suppliers.json"
	if len(args) > 0 {
		dumpPath = args[0]
	}
	dbPath := "data/app.db"
	if len(args) > 1 {
		dbPath = args[1]
	}
	raw, err := os.ReadFile(dumpPath)
	if err != nil {
		return fmt.Errorf("read dump: %w", err)
	}
	var dump []dumpSupplier
	if err := json.Unmarshal(raw, &dump); err != nil {
		return fmt.Errorf("json: %w", err)
	}
	st, err := store.NewSQLiteStore(dbPath, "development")
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer func() { _ = st.Close() }()

	for _, s := range dump {
		id, err := upsertSupplier(st, s)
		if err != nil {
			return fmt.Errorf("%s: %w", s.Name, err)
		}
		n := 0
		for _, m := range s.Messages {
			_, err := st.SaveSupplierMessage(id, store.SupplierMessageInput{
				WAMessageID: m.ID,
				FromMe:      m.FromMe,
				Body:        m.Body,
				SentAt:      time.Unix(m.TS, 0).UTC(),
			})
			if err != nil {
				return fmt.Errorf("%s msg %s: %w", s.Name, m.ID, err)
			}
			n++
		}
		fmt.Printf("%s id=%d msgs=%d\n", s.Name, id, n)
	}
	return nil
}

func upsertSupplier(st *store.SQLiteStore, s dumpSupplier) (int64, error) {
	found, err := st.GetSupplierByChatID(s.WhatsApp)
	if err == nil {
		if err := st.UpdateSupplier(found.ID, store.SupplierInput{
			Name:          s.Name,
			WhatsApp:      s.WhatsApp,
			Notes:         s.Notes,
			MinOrderCents: s.MinOrderCents,
		}); err != nil {
			return 0, err
		}
		return found.ID, nil
	}
	if !errors.Is(err, store.ErrNotFound) {
		return 0, err
	}
	return st.CreateSupplier(store.SupplierInput{
		Name:          s.Name,
		WhatsApp:      s.WhatsApp,
		Notes:         s.Notes,
		MinOrderCents: s.MinOrderCents,
	})
}
