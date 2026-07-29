package domain

import "testing"

func TestProductSlug(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Monitor Dell P1914Sf 19\" (sem base)", "monitor-dell-p1914sf-19-sem-base"},
		{"Monitor LG E1941S-PN 18,5\" (sem base)", "monitor-lg-e1941s-pn-185-sem-base"},
		{"  Ação & Peças  ", "acao-pecas"},
		{"", "produto"},
		{"!!!", "produto"},
	}
	for _, tc := range cases {
		got := ProductSlug(tc.in)
		if got != tc.want {
			t.Errorf("ProductSlug(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}

func TestUniqueProductSlug(t *testing.T) {
	taken := map[string]bool{"monitor-x": true}
	got := UniqueProductSlug("Monitor X", 7, func(s string) bool { return taken[s] })
	if got != "monitor-x-7" {
		t.Fatalf("got %q", got)
	}
	got = UniqueProductSlug("Monitor Y", 1, func(s string) bool { return taken[s] })
	if got != "monitor-y" {
		t.Fatalf("got %q", got)
	}
}
