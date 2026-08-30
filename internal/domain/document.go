package domain

import (
	"fmt"
	"strings"
)

// DigitsOnly keeps only 0-9 from s.
func DigitsOnly(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ValidCNPJ reports whether s is a 14-digit CNPJ with valid check digits.
func ValidCNPJ(s string) bool {
	d := DigitsOnly(s)
	if len(d) != 14 || allSameDigits(d) {
		return false
	}
	nums := digitInts(d)
	w1 := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	w2 := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	return mod11Digit(nums[:12], w1) == nums[12] && mod11Digit(nums[:13], w2) == nums[13]
}

// ValidCPF reports whether s is an 11-digit CPF with valid check digits.
func ValidCPF(s string) bool {
	d := DigitsOnly(s)
	if len(d) != 11 || allSameDigits(d) {
		return false
	}
	nums := digitInts(d)
	w1 := []int{10, 9, 8, 7, 6, 5, 4, 3, 2}
	w2 := []int{11, 10, 9, 8, 7, 6, 5, 4, 3, 2}
	return mod11Digit(nums[:9], w1) == nums[9] && mod11Digit(nums[:10], w2) == nums[10]
}

// FormatCNPJ returns 00.000.000/0000-00 or an error if not 14 digits.
func FormatCNPJ(s string) (string, error) {
	d := DigitsOnly(s)
	if len(d) != 14 {
		return "", fmt.Errorf("CNPJ %q has %d digits, want 14", s, len(d))
	}
	return d[0:2] + "." + d[2:5] + "." + d[5:8] + "/" + d[8:12] + "-" + d[12:14], nil
}

// FormatCPF returns 000.000.000-00 or an error if not 11 digits.
func FormatCPF(s string) (string, error) {
	d := DigitsOnly(s)
	if len(d) != 11 {
		return "", fmt.Errorf("CPF %q has %d digits, want 11", s, len(d))
	}
	return d[0:3] + "." + d[3:6] + "." + d[6:9] + "-" + d[9:11], nil
}

// FormatDocument formats CPF (11) or CNPJ (14).
func FormatDocument(s string) (string, error) {
	d := DigitsOnly(s)
	switch len(d) {
	case 11:
		if !ValidCPF(d) {
			return "", fmt.Errorf("CPF %q failed check digits", s)
		}
		return FormatCPF(d)
	case 14:
		if !ValidCNPJ(d) {
			return "", fmt.Errorf("CNPJ %q failed check digits", s)
		}
		return FormatCNPJ(d)
	default:
		return "", fmt.Errorf("document %q has %d digits, want 11 (CPF) or 14 (CNPJ)", s, len(d))
	}
}

// FormatBRDate turns YYYY-MM-DD or RFC3339-ish values into DD/MM/YYYY.
func FormatBRDate(s string) string {
	s = strings.TrimSpace(s)
	if len(s) < 10 {
		return s
	}
	day := s[:10]
	if len(day) != 10 || day[4] != '-' || day[7] != '-' {
		return s
	}
	return day[8:10] + "/" + day[5:7] + "/" + day[0:4]
}

func allSameDigits(d string) bool {
	if d == "" {
		return true
	}
	for i := 1; i < len(d); i++ {
		if d[i] != d[0] {
			return false
		}
	}
	return true
}

func digitInts(d string) []int {
	out := make([]int, len(d))
	for i := 0; i < len(d); i++ {
		out[i] = int(d[i] - '0')
	}
	return out
}

func mod11Digit(nums, weights []int) int {
	sum := 0
	for i := range weights {
		sum += nums[i] * weights[i]
	}
	r := sum % 11
	if r < 2 {
		return 0
	}
	return 11 - r
}
