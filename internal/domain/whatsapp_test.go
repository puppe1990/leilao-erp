package domain

import (
	"net/url"
	"strings"
	"testing"
)

func TestNormalizeWhatsAppPhone(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"11999998888", "5511999998888"},
		{"(11) 99999-8888", "5511999998888"},
		{"+55 11 99999-8888", "5511999998888"},
		{"5511999998888", "5511999998888"},
		{"", ""},
		{"abc", ""},
	}
	for _, tc := range cases {
		got := NormalizeWhatsAppPhone(tc.in)
		if got != tc.want {
			t.Errorf("NormalizeWhatsAppPhone(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestWhatsAppOrderURL(t *testing.T) {
	u := WhatsAppOrderURL("11987654321", "Monitor Dell P2016t 19,5\"", "R$ 279,00")
	if u == "" {
		t.Fatal("expected URL")
	}
	if !strings.HasPrefix(u, "https://wa.me/5511987654321?text=") {
		t.Fatalf("prefix: %s", u)
	}
	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatal(err)
	}
	text := parsed.Query().Get("text")
	if !strings.Contains(text, "Monitor Dell P2016t") {
		t.Errorf("text missing product: %q", text)
	}
	if !strings.Contains(text, "279") {
		t.Errorf("text missing price: %q", text)
	}
}

func TestWhatsAppOrderURL_EmptyPhone(t *testing.T) {
	if WhatsAppOrderURL("", "X", "R$ 1") != "" {
		t.Fatal("want empty when no phone")
	}
}
