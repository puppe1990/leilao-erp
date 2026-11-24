package store

import (
	"errors"
	"testing"
	"time"
)

func TestCreateListUpdateDeleteSupplier(t *testing.T) {
	st := newTestStore(t)

	list, err := st.ListSuppliers()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("want empty, got %d", len(list))
	}

	id, err := st.CreateSupplier(SupplierInput{
		Name:          "PMCELL São Paulo",
		WhatsApp:      "11911304693",
		Notes:         "power bank, carregador, fone TWS, suporte",
		MinOrderCents: 30000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if id <= 0 {
		t.Fatal("id")
	}

	if _, err := st.CreateSupplier(SupplierInput{Name: "  ", WhatsApp: "11911304693"}); err == nil {
		t.Fatal("expected error empty name")
	}
	if _, err := st.CreateSupplier(SupplierInput{Name: "X", WhatsApp: "abc"}); err == nil {
		t.Fatal("expected error bad whatsapp")
	}
	if _, err := st.CreateSupplier(SupplierInput{Name: "X", WhatsApp: "11911304693", MinOrderCents: -1}); err == nil {
		t.Fatal("expected error negative min")
	}

	list, err = st.ListSuppliers()
	if err != nil || len(list) != 1 {
		t.Fatalf("list=%d err=%v", len(list), err)
	}
	got := list[0]
	if got.Name != "PMCELL São Paulo" || got.ChatID != "5511911304693@c.us" {
		t.Fatalf("%+v", got)
	}
	if got.WhatsApp != "5511911304693" || got.MinOrderCents != 30000 {
		t.Fatalf("%+v", got)
	}
	if got.Notes != "power bank, carregador, fone TWS, suporte" {
		t.Fatalf("%+v", got)
	}

	found, err := st.GetSupplier(id)
	if err != nil || found.ChatID != "5511911304693@c.us" {
		t.Fatalf("%+v %v", found, err)
	}

	byChat, err := st.GetSupplierByChatID("5511911304693@c.us")
	if err != nil || byChat.ID != id {
		t.Fatalf("%+v %v", byChat, err)
	}

	if err := st.UpdateSupplier(id, SupplierInput{
		Name:          "PMCELL",
		WhatsApp:      "5511911304693",
		Notes:         "25 de Março",
		MinOrderCents: 30000,
	}); err != nil {
		t.Fatal(err)
	}
	found, _ = st.GetSupplier(id)
	if found.Name != "PMCELL" || found.Notes != "25 de Março" {
		t.Fatalf("%+v", found)
	}

	if err := st.DeleteSupplier(id); err != nil {
		t.Fatal(err)
	}
	if _, err := st.GetSupplier(id); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	if err := st.DeleteSupplier(id); !errors.Is(err, ErrNotFound) {
		t.Fatal("second delete should fail")
	}
}

func TestCreateSupplier_DuplicateChatID(t *testing.T) {
	st := newTestStore(t)
	_, err := st.CreateSupplier(SupplierInput{Name: "Awei", WhatsApp: "11972711435"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = st.CreateSupplier(SupplierInput{Name: "Awei 2", WhatsApp: "5511972711435@c.us"})
	if err == nil {
		t.Fatal("expected duplicate")
	}
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
}

func TestSaveSupplierMessage_IdempotentAndChronological(t *testing.T) {
	st := newTestStore(t)
	id, err := st.CreateSupplier(SupplierInput{Name: "Ecooda", WhatsApp: "11986089696"})
	if err != nil {
		t.Fatal(err)
	}
	t0 := time.Date(2026, 11, 24, 14, 4, 0, 0, time.UTC)
	t1 := t0.Add(2 * time.Minute)

	mid, err := st.SaveSupplierMessage(id, SupplierMessageInput{
		WAMessageID: "true_5511986089696@c.us_3EB03F723C1ACFF871B69B",
		FromMe:      true,
		Body:        "Vocês têm tabela de atacado?",
		SentAt:      t0,
	})
	if err != nil || mid <= 0 {
		t.Fatalf("id=%d err=%v", mid, err)
	}
	again, err := st.SaveSupplierMessage(id, SupplierMessageInput{
		WAMessageID: "true_5511986089696@c.us_3EB03F723C1ACFF871B69B",
		FromMe:      true,
		Body:        "Vocês têm tabela de atacado?",
		SentAt:      t0,
	})
	if err != nil || again != mid {
		t.Fatalf("idempotent want %d got %d err=%v", mid, again, err)
	}

	_, err = st.SaveSupplierMessage(id, SupplierMessageInput{
		WAMessageID: "false_5511986089696@c.us_AAA",
		FromMe:      false,
		Body:        "Mínimo R$ 300",
		SentAt:      t1,
	})
	if err != nil {
		t.Fatal(err)
	}

	msgs, err := st.ListSupplierMessages(id)
	if err != nil || len(msgs) != 2 {
		t.Fatalf("len=%d err=%v", len(msgs), err)
	}
	if !msgs[0].FromMe || msgs[0].Body != "Vocês têm tabela de atacado?" {
		t.Fatalf("first %+v", msgs[0])
	}
	if msgs[1].FromMe || msgs[1].Body != "Mínimo R$ 300" {
		t.Fatalf("second %+v", msgs[1])
	}

	list, err := st.ListSuppliers()
	if err != nil || len(list) != 1 {
		t.Fatalf("%+v %v", list, err)
	}
	if list[0].LastBody != "Mínimo R$ 300" || list[0].LastFromMe || list[0].MessageCount != 2 {
		t.Fatalf("preview %+v", list[0])
	}
}

func TestDeleteSupplier_RemovesMessages(t *testing.T) {
	st := newTestStore(t)
	id, err := st.CreateSupplier(SupplierInput{Name: "Victor", WhatsApp: "11958310194"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = st.SaveSupplierMessage(id, SupplierMessageInput{
		WAMessageID: "true_x_1",
		FromMe:      true,
		Body:        "oi",
		SentAt:      time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := st.DeleteSupplier(id); err != nil {
		t.Fatal(err)
	}
	msgs, err := st.ListSupplierMessages(id)
	if err != nil || len(msgs) != 0 {
		t.Fatalf("want no messages after delete, got %d err=%v", len(msgs), err)
	}
}

func TestSuppliersTablesExist(t *testing.T) {
	st := newTestStore(t)
	for _, table := range []string{"suppliers", "supplier_messages"} {
		var name string
		err := st.db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name)
		if err != nil {
			t.Fatalf("%s table not found: %v", table, err)
		}
	}
}
