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
	s := fmt.Sprintf("R$ %d,%02d", reais, cent)
	if neg {
		return "-" + s
	}
	return s
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
