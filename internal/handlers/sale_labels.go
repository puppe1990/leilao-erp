package handlers

import "fmt"

func channelLabel(channel string) string {
	switch channel {
	case "direct":
		return "Direto"
	case "mercadolivre":
		return "Mercado Livre"
	case "shopee":
		return "Shopee"
	case "olx":
		return "OLX"
	case "other":
		return "Outro"
	default:
		return channel
	}
}

func channelOptions() []map[string]string {
	return []map[string]string{
		{"value": "direct", "label": "Direto"},
		{"value": "mercadolivre", "label": "Mercado Livre"},
		{"value": "shopee", "label": "Shopee"},
		{"value": "olx", "label": "OLX"},
		{"value": "other", "label": "Outro"},
	}
}

func validChannel(channel string) bool {
	switch channel {
	case "direct", "mercadolivre", "shopee", "olx", "other":
		return true
	default:
		return false
	}
}

func paymentStatusLabel(status string) string {
	switch status {
	case "received":
		return "Recebido"
	case "pending":
		return "A receber"
	case "cancelled":
		return "Cancelado"
	default:
		return status
	}
}

func formatCentsInput(cents int64) string {
	neg := cents < 0
	if neg {
		cents = -cents
	}
	s := fmt.Sprintf("%d,%02d", cents/100, cents%100)
	if neg {
		return "-" + s
	}
	return s
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
