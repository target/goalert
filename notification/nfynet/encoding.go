package nfynet

import (
	"fmt"
	"strings"
	"unicode"
)

// ValidateIDString will validate a string representation of an ID, only printable characters are allowed.
func ValidateIDString(s string) error {
	for i, r := range s {
		if !unicode.IsPrint(r) {
			return fmt.Errorf("unprintable character at %d: 0x%x", i, r)
		}
	}

	return nil
}

func validateEncoding(s string) error {
	escaped := strings.Count(s, "`d") + strings.Count(s, "`e")
	escapeChars := strings.Count(s, "`")
	if escaped != escapeChars {
		return fmt.Errorf("invalid encoding: %s", s)
	}

	return ValidateIDString(s)
}

func escape(s string) string {
	s = strings.ReplaceAll(s, "`", "`e")
	s = strings.ReplaceAll(s, "|", "`d")
	return s
}

func unescape(s string) string {
	s = strings.ReplaceAll(s, "`d", "|")
	s = strings.ReplaceAll(s, "`e", "`")
	return s
}
