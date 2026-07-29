package domain

import (
	"fmt"
	"strings"
	"unicode"
)

// ProductSlug turns a product name into a URL-safe slug (ASCII, hyphenated).
func ProductSlug(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	if s == "" {
		return "produto"
	}
	var b strings.Builder
	b.Grow(len(s))
	prevDash := false
	for _, r := range s {
		r = foldAccent(r)
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		case unicode.IsSpace(r) || r == '-' || r == '_' || r == '/' || r == '.':
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "produto"
	}
	// keep URLs reasonable
	if len(out) > 80 {
		out = strings.Trim(out[:80], "-")
	}
	return out
}

// UniqueProductSlug prefers base; if taken, appends -id (or -n).
func UniqueProductSlug(base string, id int64, taken func(slug string) bool) string {
	base = ProductSlug(base)
	if base == "" {
		base = "produto"
	}
	if !taken(base) {
		return base
	}
	if id > 0 {
		cand := fmt.Sprintf("%s-%d", base, id)
		if !taken(cand) {
			return cand
		}
	}
	for n := 2; n < 1000; n++ {
		cand := fmt.Sprintf("%s-%d", base, n)
		if !taken(cand) {
			return cand
		}
	}
	return fmt.Sprintf("%s-%d", base, id)
}

func foldAccent(r rune) rune {
	switch r {
	case 'á', 'à', 'ã', 'â', 'ä':
		return 'a'
	case 'é', 'è', 'ê', 'ë':
		return 'e'
	case 'í', 'ì', 'î', 'ï':
		return 'i'
	case 'ó', 'ò', 'õ', 'ô', 'ö':
		return 'o'
	case 'ú', 'ù', 'û', 'ü':
		return 'u'
	case 'ç':
		return 'c'
	case 'ñ':
		return 'n'
	default:
		return r
	}
}
