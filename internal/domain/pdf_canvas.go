package domain

import (
	"bytes"
	"fmt"
	"strings"
)

// AuctionHQ print palette (light theme tokens — readable on paper).
var (
	colBg          = rgb{0.957, 0.965, 0.973} // #F4F6F8
	colInk         = rgb{0.067, 0.094, 0.153} // #111827
	colMuted       = rgb{0.294, 0.333, 0.388} // #4B5563
	colGreen       = rgb{0.086, 0.639, 0.290} // #16A34A
	colHeader      = rgb{0, 0, 0}
	colWhite       = rgb{1, 1, 1}
	colCard        = rgb{1, 1, 1}
	colBorder      = rgb{0.898, 0.906, 0.922} // #E5E7EB
	colHeaderMuted = rgb{0.64, 0.64, 0.64}
)

type rgb struct{ r, g, b float64 }

const (
	fontSans     = "/F1"
	fontSansBold = "/F2"
	fontMono     = "/F3"
	fontMonoBold = "/F4"
)

type pdfCanvas struct {
	ops []string
}

func (c *pdfCanvas) fillRect(x, y, w, h float64, col rgb) {
	c.ops = append(c.ops, fmt.Sprintf(
		"%.3f %.3f %.3f rg %.1f %.1f %.1f %.1f re f",
		col.r, col.g, col.b, x, y, w, h,
	))
}

func (c *pdfCanvas) strokeRect(x, y, w, h, lw float64, col rgb) {
	c.ops = append(c.ops, fmt.Sprintf(
		"%.3f %.3f %.3f RG %.2f w %.1f %.1f %.1f %.1f re S",
		col.r, col.g, col.b, lw, x, y, w, h,
	))
}

func (c *pdfCanvas) fillStrokeRect(x, y, w, h, lw float64, fill, stroke rgb) {
	c.ops = append(c.ops, fmt.Sprintf(
		"%.3f %.3f %.3f rg %.3f %.3f %.3f RG %.2f w %.1f %.1f %.1f %.1f re B",
		fill.r, fill.g, fill.b, stroke.r, stroke.g, stroke.b, lw, x, y, w, h,
	))
}

func (c *pdfCanvas) hline(x1, x2, y, lw float64, col rgb) {
	c.line(x1, y, x2, y, lw, col)
}

func (c *pdfCanvas) vline(x, y1, y2, lw float64, col rgb) {
	c.line(x, y1, x, y2, lw, col)
}

func (c *pdfCanvas) line(x1, y1, x2, y2, lw float64, col rgb) {
	c.ops = append(c.ops, fmt.Sprintf(
		"%.3f %.3f %.3f RG %.2f w %.1f %.1f m %.1f %.1f l S",
		col.r, col.g, col.b, lw, x1, y1, x2, y2,
	))
}

func (c *pdfCanvas) text(x, y, size float64, font string, col rgb, s string) {
	c.ops = append(c.ops, fmt.Sprintf(
		"%.3f %.3f %.3f rg BT %s %.1f Tf 1 0 0 1 %.1f %.1f Tm %s Tj ET",
		col.r, col.g, col.b, font, size, x, y, pdfLiteral(s),
	))
}

func (c *pdfCanvas) textRight(right, y, size float64, font string, col rgb, s string) {
	c.text(right-approxWidth(s, size, font), y, size, font, col, s)
}

func (c *pdfCanvas) bytes() []byte {
	content := strings.Join(c.ops, "\n")
	objs := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		fmt.Sprintf(
			"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.0f %.0f] /Resources << /Font << /F1 4 0 R /F2 5 0 R /F3 6 0 R /F4 7 0 R >> >> /Contents 8 0 R >>",
			pdfPageW, pdfPageH,
		),
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold /Encoding /WinAnsiEncoding >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Courier /Encoding /WinAnsiEncoding >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Courier-Bold /Encoding /WinAnsiEncoding >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content), content),
	}
	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n%\xE2\xE3\xCF\xD3\n")
	offsets := make([]int, len(objs)+1)
	for i, obj := range objs {
		offsets[i+1] = buf.Len()
		fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}
	xref := buf.Len()
	fmt.Fprintf(&buf, "xref\n0 %d\n", len(objs)+1)
	buf.WriteString("0000000000 65535 f \n")
	for i := 1; i <= len(objs); i++ {
		fmt.Fprintf(&buf, "%010d 00000 n \n", offsets[i])
	}
	fmt.Fprintf(&buf, "trailer << /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objs)+1, xref)
	return buf.Bytes()
}

func approxWidth(s string, size float64, font string) float64 {
	em := 0.50
	switch font {
	case fontSansBold:
		em = 0.57
	case fontMono, fontMonoBold:
		em = 0.60
	}
	return float64(len(toWinAnsi(s))) * size * em
}

func pdfLiteral(s string) string {
	enc := toWinAnsi(s)
	var b strings.Builder
	b.WriteByte('(')
	for i := 0; i < len(enc); i++ {
		ch := enc[i]
		switch ch {
		case '\\', '(', ')':
			b.WriteByte('\\')
			b.WriteByte(ch)
		default:
			b.WriteByte(ch)
		}
	}
	b.WriteByte(')')
	return b.String()
}

func toWinAnsi(s string) []byte {
	var out []byte
	for _, r := range s {
		if r < 128 {
			out = append(out, byte(r))
			continue
		}
		if b, ok := winAnsi[r]; ok {
			out = append(out, b)
			continue
		}
		out = append(out, '?')
	}
	return out
}

// WinAnsi / Windows-1252 extras used in Portuguese copy.
var winAnsi = map[rune]byte{
	'Á': 0xC1, 'À': 0xC0, 'Â': 0xC2, 'Ã': 0xC3, 'Ä': 0xC4,
	'É': 0xC9, 'Ê': 0xCA, 'Í': 0xCD, 'Ó': 0xD3, 'Ô': 0xD4,
	'Õ': 0xD5, 'Ú': 0xDA, 'Ü': 0xDC, 'Ç': 0xC7,
	'á': 0xE1, 'à': 0xE0, 'â': 0xE2, 'ã': 0xE3, 'ä': 0xE4,
	'é': 0xE9, 'ê': 0xEA, 'í': 0xED, 'ó': 0xF3, 'ô': 0xF4,
	'õ': 0xF5, 'ú': 0xFA, 'ü': 0xFC, 'ç': 0xE7,
	'º': 0xBA, 'ª': 0xAA, '–': 0x96, '—': 0x97, '’': 0x92,
	'·': 0xB7,
}
