package common

import (
	"strings"
	"unicode"
)
// NormalizeInput membersihkan teks input
func NormalizeInput(text string) string {
	return strings.TrimSpace(strings.ToLower(text))
}

func IsNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return s != ""
}
