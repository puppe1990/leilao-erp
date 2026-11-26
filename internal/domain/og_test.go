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

func TestOGHeadHTML_emptySiteNameFallsBackToPixEletro(t *testing.T) {
	html := OGHeadHTML(OGMeta{Title: "Catálogo"})
	if !strings.Contains(html, `property="og:site_name" content="Pix Eletro"`) {
		t.Fatalf("missing Pix Eletro site_name in %s", html)
	}
	if strings.Contains(html, `content="Puppe"`) {
		t.Fatal("legacy Puppe site_name still present")
	}
}

func TestOGHeadHTML_emptyDescriptionMentionsTechAndWhatsApp(t *testing.T) {
	html := OGHeadHTML(OGMeta{Title: "Catálogo"})
	if !strings.Contains(html, "Monitores e tech") {
		t.Fatalf("fallback description should mention tech in %s", html)
	}
	if !strings.Contains(html, "WhatsApp") {
		t.Fatalf("fallback description should mention WhatsApp in %s", html)
	}
	if strings.Contains(html, "Monitores usados testados") {
		t.Fatal("legacy monitor-only description still present")
	}
}

func TestOGHeadHTML(t *testing.T) {
	html := OGHeadHTML(OGMeta{
		Title:       `Monitor "LG"`,
		Description: "Usado · testado",
		ImageURL:    "http://localhost:8080/static/og-shop.png",
		PageURL:     "http://localhost:8080/",
		SiteName:    "Pix Eletro",
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
