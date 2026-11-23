package store

import (
	"errors"
	"testing"
)

func TestCreateListUpdateDeleteFind(t *testing.T) {
	st := newTestStore(t)

	list, err := st.ListFinds("")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("want empty, got %d", len(list))
	}

	id, err := st.CreateFind(FindInput{
		Name:      "Dell P2016t 19,5\"",
		URL:       "https://www.olx.com.br/vi/1555022453",
		Source:    FindSourceOLX,
		BuyCents:  17500,
		SellCents: 25000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if id <= 0 {
		t.Fatal("id")
	}

	if _, err := st.CreateFind(FindInput{Name: "  ", URL: "https://olx.com.br/vi/1", Source: FindSourceOLX}); err == nil {
		t.Fatal("expected error empty name")
	}
	if _, err := st.CreateFind(FindInput{Name: "HP", URL: "olx.com.br/vi/1", Source: FindSourceOLX}); err == nil {
		t.Fatal("expected error url without scheme")
	}
	if _, err := st.CreateFind(FindInput{Name: "HP", URL: "https://olx.com.br/vi/1", Source: "shopee"}); err == nil {
		t.Fatal("expected error unknown source")
	}
	if _, err := st.CreateFind(FindInput{Name: "HP", URL: "https://olx.com.br/vi/1", Source: FindSourceOLX, BuyCents: -1}); err == nil {
		t.Fatal("expected error negative buy")
	}

	list, err = st.ListFinds("")
	if err != nil || len(list) != 1 {
		t.Fatalf("list=%d err=%v", len(list), err)
	}
	got := list[0]
	if got.Name != "Dell P2016t 19,5\"" || got.URL != "https://www.olx.com.br/vi/1555022453" {
		t.Fatalf("%+v", got)
	}
	if got.Source != FindSourceOLX || got.BuyCents != 17500 || got.SellCents != 25000 {
		t.Fatalf("%+v", got)
	}

	found, err := st.GetFind(id)
	if err != nil || found.BuyCents != 17500 {
		t.Fatalf("%+v %v", found, err)
	}

	if err := st.UpdateFind(id, FindInput{
		Name:      "Dell P2016t",
		URL:       "https://www.facebook.com/marketplace/item/1645626870456007/",
		Source:    FindSourceFacebook,
		BuyCents:  15000,
		SellCents: 22000,
	}); err != nil {
		t.Fatal(err)
	}
	found, _ = st.GetFind(id)
	if found.Name != "Dell P2016t" || found.Source != FindSourceFacebook || found.BuyCents != 15000 {
		t.Fatalf("%+v", found)
	}

	if err := st.DeleteFind(id); err != nil {
		t.Fatal(err)
	}
	if _, err := st.GetFind(id); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	if err := st.DeleteFind(id); !errors.Is(err, ErrNotFound) {
		t.Fatal("second delete should fail")
	}
}

func TestListFinds_NewestFirstAndSourceFilter(t *testing.T) {
	st := newTestStore(t)
	_, _ = st.CreateFind(FindInput{
		Name: "Parafusadeira OLX", URL: "https://www.olx.com.br/vi/1", Source: FindSourceOLX, BuyCents: 8000, SellCents: 15000,
	})
	_, _ = st.CreateFind(FindInput{
		Name: "DeWalt Kwara", URL: "https://kwara.com.br/bens/2", Source: FindSourceKwara, BuyCents: 26975, SellCents: 40000,
	})
	_, _ = st.CreateFind(FindInput{
		Name: "Vostro Face", URL: "https://www.facebook.com/marketplace/item/3/", Source: FindSourceFacebook, BuyCents: 0, SellCents: 40000,
	})

	all, err := st.ListFinds("")
	if err != nil || len(all) != 3 {
		t.Fatalf("all=%d err=%v", len(all), err)
	}
	if all[0].Name != "Vostro Face" {
		t.Fatalf("want newest first, got %q", all[0].Name)
	}

	kwara, err := st.ListFinds(FindSourceKwara)
	if err != nil || len(kwara) != 1 || kwara[0].Name != "DeWalt Kwara" {
		t.Fatalf("%+v %v", kwara, err)
	}
}

func TestFindsTableExists(t *testing.T) {
	st := newTestStore(t)
	var name string
	err := st.db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, "finds").Scan(&name)
	if err != nil {
		t.Fatalf("finds table not found: %v", err)
	}
}
