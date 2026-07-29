package domain

import (
	"strings"
	"testing"
)

func TestAbsoluteURL(t *testing.T) {
	if got := AbsoluteURL("http://localhost:8080", "/static/og-shop.png"); got != "http://localhost:8080/static/og-shop.png" {
		t.Fatalf("got %q", got)
	}
	if got := AbsoluteURL("http://localhost:8080/", "static/x.png"); got != "http://localhost:8080/static/x.png" {
		t.Fatalf("got %q", got)
	}
	if got := AbsoluteURL("http://x", "https://cdn/a.jpg"); got != "https://cdn/a.jpg" {
		t.Fatalf("got %q", got)
	}
}

func TestOGHeadHTML(t *testing.T) {
	html := OGHeadHTML(OGMeta{
		Title:       `Monitor "LG"`,
		Description: "Usado · testado",
		ImageURL:    "http://localhost:8080/static/og-shop.png",
		PageURL:     "http://localhost:8080/",
		SiteName:    "Puppe",
		Type:        "website",
	})
	for _, want := range []string{
		`property="og:title"`,
		`Monitor &#34;LG&#34;`,
		`property="og:image" content="http://localhost:8080/static/og-shop.png"`,
		`name="twitter:card" content="summary_large_image"`,
		`rel="canonical"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("missing %q in %s", want, html)
		}
	}
}
