package utils

import (
	"strings"
)

func EnsureSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		return s
	}
	return s + suffix
}

func SanitizeName(s string) string {
	return strings.Map(func(r rune) rune {
		if IsLetter(r) || IsDigit(r) || r == '_' || r == '.' || r == '-' {
			return r
		}
		return '-'
	}, s)
}

func IsLetter(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z')
}

func IsDigit(r rune) bool {
	return '0' <= r && r <= '9'
}
