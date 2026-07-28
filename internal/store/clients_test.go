package store

import (
	"testing"
)

func TestCreateListUpdateDeleteClient(t *testing.T) {
	st := newTestStore(t)

	list, err := st.ListClients()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("want empty, got %d", len(list))
	}

	id, err := st.CreateClient(ClientInput{
		Name:     "Maria Silva",
		Phone:    "11999998888",
		Email:    "maria@example.com",
		Document: "123.456.789-00",
		Notes:    "Prefere PIX",
	})
	if err != nil {
		t.Fatal(err)
	}
	if id <= 0 {
		t.Fatal("id")
	}

	// name required
	if _, err := st.CreateClient(ClientInput{Name: "  "}); err == nil {
		t.Fatal("expected error empty name")
	}

	list, err = st.ListClients()
	if err != nil || len(list) != 1 {
		t.Fatalf("list=%d err=%v", len(list), err)
	}
	if list[0].Name != "Maria Silva" || list[0].Phone != "11999998888" {
		t.Fatalf("%+v", list[0])
	}

	found, err := st.FindClient(id)
	if err != nil || found.Email != "maria@example.com" {
		t.Fatalf("%+v %v", found, err)
	}

	if err := st.UpdateClient(id, ClientInput{
		Name:  "Maria S.",
		Phone: "11888887777",
		Email: "ms@example.com",
	}); err != nil {
		t.Fatal(err)
	}
	found, _ = st.FindClient(id)
	if found.Name != "Maria S." || found.Phone != "11888887777" {
		t.Fatalf("%+v", found)
	}

	if err := st.DeleteClient(id); err != nil {
		t.Fatal(err)
	}
	if _, err := st.FindClient(id); err == nil {
		t.Fatal("expected not found")
	}
	if err := st.DeleteClient(id); err == nil {
		t.Fatal("second delete should fail")
	}
}

func TestListClients_SearchOrder(t *testing.T) {
	st := newTestStore(t)
	_, _ = st.CreateClient(ClientInput{Name: "Bruno", Phone: "11"})
	_, _ = st.CreateClient(ClientInput{Name: "Ana", Email: "ana@x.com"})
	_, _ = st.CreateClient(ClientInput{Name: "Carlos", Document: "999"})

	all, err := st.ListClients()
	if err != nil || len(all) != 3 {
		t.Fatal(all, err)
	}
	// ordered by name
	if all[0].Name != "Ana" || all[2].Name != "Carlos" {
		t.Fatalf("order: %v %v %v", all[0].Name, all[1].Name, all[2].Name)
	}

	// search by phone fragment
	hit, err := st.SearchClients("11")
	if err != nil || len(hit) != 1 || hit[0].Name != "Bruno" {
		t.Fatalf("search phone: %+v %v", hit, err)
	}
	hit, err = st.SearchClients("ana@")
	if err != nil || len(hit) != 1 {
		t.Fatalf("search email: %+v %v", hit, err)
	}
}
