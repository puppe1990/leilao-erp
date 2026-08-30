package domain

import "strings"

const (
	pdfPageW = 595.0
	pdfPageH = 842.0
	pdfLeft  = 40.0
	pdfRight = 555.0
)

// BuildReceiptPDF renders a one-page A4 sale receipt in AuctionHQ print style.
func BuildReceiptPDF(r Receipt) ([]byte, error) {
	if err := ValidateReceipt(r); err != nil {
		return nil, err
	}
	sellerCNPJ, _ := FormatCNPJ(r.Seller.Document)
	buyerDoc, _ := FormatDocument(r.Buyer.Document)

	var c pdfCanvas
	c.fillRect(0, 0, pdfPageW, pdfPageH, colBg)
	c.drawHeader(r)

	y := 740.0
	y = c.partyCard(y, "EMITENTE", strings.TrimSpace(r.Seller.Name), "CNPJ "+sellerCNPJ, "", "")
	y = c.partyCard(y, "CLIENTE", strings.TrimSpace(r.Buyer.Name), "CPF/CNPJ "+buyerDoc,
		phoneLine(r.Buyer.Phone), emailLine(r.Buyer.Email))
	y = c.itemsCard(y, r.Items)
	y = c.totalCard(y, r)
	c.footer(y, r, sellerCNPJ)
	return c.bytes(), nil
}

func (c *pdfCanvas) drawHeader(r Receipt) {
	c.fillRect(0, 766, pdfPageW, 76, colHeader)
	c.fillRect(0, 762, pdfPageW, 4, colGreen)
	c.brandMark(36, 794)
	brand := brandMain(r.Seller.Name)
	c.text(66, 814, 16, fontSansBold, colWhite, brand)
	dotX := 66 + approxWidth(brand, 16, fontSansBold) + 1.5
	c.fillRect(dotX, 818, 3.2, 3.2, colGreen)
	c.text(66, 798, 8, fontSansBold, colHeaderMuted, "Admin  ·  ERP")
	c.textRight(pdfRight, 812, 14, fontSansBold, colWhite, "RECIBO DE VENDA")
	c.textRight(pdfRight, 796, 10, fontMono, colGreen, "Nº "+strings.TrimSpace(r.Number))
}

func (c *pdfCanvas) brandMark(x, y float64) {
	c.fillRect(x, y, 22, 22, colGreen)
	c.strokeRect(x+4, y+9, 14, 9, 1.1, colWhite)
	c.hline(x+8, x+14, y+5.5, 1.1, colWhite)
	c.vline(x+11, y+9, y+5.5, 1.1, colWhite)
}

func (c *pdfCanvas) partyCard(top float64, label, name, doc, extra1, extra2 string) float64 {
	lines := 2
	if extra1 != "" {
		lines++
	}
	if extra2 != "" {
		lines++
	}
	h := 28 + float64(lines)*14
	bottom := top - h
	c.fillStrokeRect(pdfLeft, bottom, pdfRight-pdfLeft, h, 0.6, colCard, colBorder)
	c.fillRect(pdfLeft, bottom, 3, h, colGreen)
	c.text(pdfLeft+14, top-16, 8, fontSansBold, colMuted, label)
	c.text(pdfLeft+14, top-32, 12, fontSansBold, colInk, name)
	y := top - 46
	c.text(pdfLeft+14, y, 10, fontMono, colInk, doc)
	if extra1 != "" {
		y -= 14
		c.text(pdfLeft+14, y, 9, fontSans, colMuted, extra1)
	}
	if extra2 != "" {
		y -= 14
		c.text(pdfLeft+14, y, 9, fontSans, colMuted, extra2)
	}
	return bottom - 12
}

func (c *pdfCanvas) itemsCard(top float64, items []ReceiptItem) float64 {
	h := 28 + float64(len(items))*16 + 8
	bottom := top - h
	c.fillStrokeRect(pdfLeft, bottom, pdfRight-pdfLeft, h, 0.6, colCard, colBorder)
	c.fillRect(pdfLeft, bottom, 3, h, colGreen)
	c.text(pdfLeft+14, top-16, 8, fontSansBold, colMuted, "REFERENTE A")
	y := top - 34
	for _, it := range items {
		c.text(pdfLeft+14, y, 10, fontSans, colInk, "- "+strings.TrimSpace(it.Description))
		y -= 16
	}
	return bottom - 12
}

func (c *pdfCanvas) totalCard(top float64, r Receipt) float64 {
	h := 58.0
	bottom := top - h
	c.fillStrokeRect(pdfLeft, bottom, pdfRight-pdfLeft, h, 0.6, colCard, colBorder)
	c.fillRect(pdfLeft, bottom, 3, h, colGreen)
	c.text(pdfLeft+14, top-18, 8, fontSansBold, colMuted, "VALOR RECEBIDO")
	c.text(pdfLeft+14, top-40, 18, fontMonoBold, colGreen, FormatBRL(r.TotalCents))
	meta := ""
	if d := FormatBRDate(r.IssuedAt); d != "" {
		meta = "Data " + d
	}
	if ch := strings.TrimSpace(r.Channel); ch != "" {
		if meta != "" {
			meta += "  ·  "
		}
		meta += ch
	}
	if meta != "" {
		c.textRight(pdfRight-14, top-36, 9, fontSans, colMuted, meta)
	}
	return bottom - 18
}

func (c *pdfCanvas) footer(top float64, r Receipt, sellerCNPJ string) {
	c.hline(pdfLeft, pdfLeft+200, top-8, 0.7, colInk)
	c.text(pdfLeft, top-22, 9, fontSans, colInk, strings.TrimSpace(r.Seller.Name))
	c.text(pdfLeft, top-34, 8, fontMono, colMuted, "CNPJ "+sellerCNPJ)
	c.text(pdfLeft, 36, 8, fontSans, colMuted, "Este recibo não é um documento fiscal.")
}

func brandMain(name string) string {
	fields := strings.Fields(strings.TrimSpace(name))
	if len(fields) == 0 {
		return "Admin"
	}
	return fields[0]
}

func phoneLine(phone string) string {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return ""
	}
	return "Telefone " + phone
}

func emailLine(email string) string {
	email = strings.TrimSpace(email)
	if email == "" {
		return ""
	}
	return "E-mail " + email
}
