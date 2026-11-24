package domain

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var nonDigits = regexp.MustCompile(`\D+`)

// WhatsAppChatID turns a BR number or chat id into 55…@c.us. Empty when invalid.
func WhatsAppChatID(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimSuffix(raw, "@c.us")
	raw = strings.TrimSuffix(raw, "@s.whatsapp.net")
	d := NormalizeWhatsAppPhone(raw)
	if d == "" {
		return ""
	}
	return d + "@c.us"
}

// NormalizeWhatsAppPhone returns digits for wa.me (adds Brazil 55 when 10–11 digits).
func NormalizeWhatsAppPhone(raw string) string {
	d := nonDigits.ReplaceAllString(raw, "")
	if d == "" {
		return ""
	}
	// already has country code
	if strings.HasPrefix(d, "55") && len(d) >= 12 {
		return d
	}
	// Brazilian mobile/landline without country
	if len(d) == 10 || len(d) == 11 {
		return "55" + d
	}
	// other lengths: only accept if purely digits and long enough
	if len(d) >= 12 {
		return d
	}
	return ""
}

// WhatsAppOrderURL builds https://wa.me/<phone>?text=... for a product inquiry.
func WhatsAppOrderURL(phone, productName, priceLabel string) string {
	digits := NormalizeWhatsAppPhone(phone)
	if digits == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString("Olá! Vi o anúncio e quero comprar:\n")
	b.WriteString("• ")
	b.WriteString(strings.TrimSpace(productName))
	if p := strings.TrimSpace(priceLabel); p != "" {
		b.WriteString("\n• Preço: ")
		b.WriteString(p)
	}
	b.WriteString("\n\nPode me passar a disponibilidade e formas de pagamento?")
	return "https://wa.me/" + digits + "?text=" + url.QueryEscape(b.String())
}

// WhatsAppLotURL builds a click-to-chat URL for the remaining auction-lot bundle.
func WhatsAppLotURL(phone string, qty int, priceLabel string) string {
	digits := NormalizeWhatsAppPhone(phone)
	if digits == "" {
		return ""
	}
	n := qty
	if n < 1 {
		n = 1
	}
	var b strings.Builder
	b.WriteString("Olá! Quero o lote completo de ")
	fmt.Fprintf(&b, "%d", n)
	b.WriteString(" monitores")
	if p := strings.TrimSpace(priceLabel); p != "" {
		b.WriteString(" por ")
		b.WriteString(p)
	}
	b.WriteString(".\n\nPode confirmar disponibilidade e combinar entrega/retirada?")
	return "https://wa.me/" + digits + "?text=" + url.QueryEscape(b.String())
}
