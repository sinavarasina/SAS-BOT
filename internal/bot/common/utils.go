package common

import "strings"

// NormalizeInput membersihkan teks input
func NormalizeInput(text string) string {
	return strings.TrimSpace(strings.ToLower(text))
}
