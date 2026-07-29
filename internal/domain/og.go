package domain

import (
	"fmt"
	"html"
	"net/url"
	"strings"
)

// AbsoluteURL joins base (APP_URL) with a path or returns absolute URLs as-is.
func AbsoluteURL(base, pathOrURL string) string {
	pathOrURL = strings.TrimSpace(pathOrURL)
	if pathOrURL == "" {
		return strings.TrimRight(strings.TrimSpace(base), "/")
	}
	if strings.HasPrefix(pathOrURL, "http://") || strings.HasPrefix(pathOrURL, "https://") {
		return pathOrURL
	}
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if base == "" {
		if strings.HasPrefix(pathOrURL, "/") {
			return pathOrURL
		}
		return "/" + pathOrURL
	}
	if !strings.HasPrefix(pathOrURL, "/") {
		pathOrURL = "/" + pathOrURL
	}
	return base + pathOrURL
}

// OGMeta describes Open Graph / Twitter card tags for a public page.
type OGMeta struct {
	Title       string
	Description string
	ImageURL    string // absolute preferred
	PageURL     string // absolute preferred
	Type        string // website | product | article
	SiteName    string
}

// OGHeadHTML returns escaped meta tags for injection into the root HTML template.
func OGHeadHTML(m OGMeta) string {
	title := strings.TrimSpace(m.Title)
	if title == "" {
		title = "Catálogo"
	}
	desc := strings.TrimSpace(m.Description)
	if desc == "" {
		desc = "Monitores usados testados. Pedido no WhatsApp. 10% OFF no PIX."
	}
	typ := m.Type
	if typ == "" {
		typ = "website"
	}
	site := m.SiteName
	if site == "" {
		site = "Puppe"
	}
	img := m.ImageURL
	page := m.PageURL

	esc := html.EscapeString
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, `<title>%s</title>`, esc(title))
	_, _ = fmt.Fprintf(&b, `<meta name="description" content="%s"/>`, esc(desc))
	_, _ = fmt.Fprintf(&b, `<meta property="og:type" content="%s"/>`, esc(typ))
	_, _ = fmt.Fprintf(&b, `<meta property="og:site_name" content="%s"/>`, esc(site))
	_, _ = fmt.Fprintf(&b, `<meta property="og:title" content="%s"/>`, esc(title))
	_, _ = fmt.Fprintf(&b, `<meta property="og:description" content="%s"/>`, esc(desc))
	if page != "" {
		_, _ = fmt.Fprintf(&b, `<meta property="og:url" content="%s"/>`, esc(page))
		_, _ = fmt.Fprintf(&b, `<link rel="canonical" href="%s"/>`, esc(page))
	}
	if img != "" {
		_, _ = fmt.Fprintf(&b, `<meta property="og:image" content="%s"/>`, esc(img))
		// Dimensions apply to the default shop card; product photos vary.
		lower := strings.ToLower(img)
		switch {
		case strings.HasSuffix(lower, ".png"), strings.Contains(img, "og-shop"):
			b.WriteString(`<meta property="og:image:width" content="1200"/>`)
			b.WriteString(`<meta property="og:image:height" content="630"/>`)
			b.WriteString(`<meta property="og:image:type" content="image/png"/>`)
		case strings.HasSuffix(lower, ".jpg"), strings.HasSuffix(lower, ".jpeg"):
			b.WriteString(`<meta property="og:image:type" content="image/jpeg"/>`)
		case strings.HasSuffix(lower, ".webp"):
			b.WriteString(`<meta property="og:image:type" content="image/webp"/>`)
		}
	}
	b.WriteString(`<meta name="twitter:card" content="summary_large_image"/>`)
	_, _ = fmt.Fprintf(&b, `<meta name="twitter:title" content="%s"/>`, esc(title))
	_, _ = fmt.Fprintf(&b, `<meta name="twitter:description" content="%s"/>`, esc(desc))
	if img != "" {
		_, _ = fmt.Fprintf(&b, `<meta name="twitter:image" content="%s"/>`, esc(img))
	}
	return b.String()
}

// JoinURLPath is a tiny helper for building page paths with optional query.
func JoinURLPath(parts ...string) string {
	var cleaned []string
	for _, p := range parts {
		p = strings.Trim(p, "/")
		if p != "" {
			cleaned = append(cleaned, p)
		}
	}
	if len(cleaned) == 0 {
		return "/"
	}
	return "/" + strings.Join(cleaned, "/")
}

// IsSafeAbsoluteURL reports whether u is http(s) for optional validation.
func IsSafeAbsoluteURL(u string) bool {
	parsed, err := url.Parse(u)
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}
