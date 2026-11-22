package domain

import (
	"fmt"
	"strconv"
	"strings"
)

func FormatBRL(cents int64) string {
	neg := cents < 0
	if neg {
		cents = -cents
	}
	reais := cents / 100
	cent := cents % 100
	s := fmt.Sprintf("R$ %s,%02d", thousands(reais), cent)
	if neg {
		return "-" + s
	}
	return s
}

func thousands(n int64) string {
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	lead := len(s) % 3
	if lead == 0 {
		lead = 3
	}
	b.WriteString(s[:lead])
	for i := lead; i < len(s); i += 3 {
		b.WriteByte('.')
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

func ParseBRLToCents(s string) (int64, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "R$")
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", ".")
	if s == "" {
		return 0, fmt.Errorf("empty amount")
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return int64(f*100 + 0.5), nil
}

func SaleNet(gross, fee, shipping int64) int64 {
	return gross - fee - shipping
}

func Margin(net, unitCostAtSale int64) int64 {
	return net - unitCostAtSale
}
