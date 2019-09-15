package util

import (
	"strings"
)

func isASCIILower(c byte) bool {
	return 'a' <= c && c <= 'z'
}
func isASCIIUpper(c byte) bool {
	return 'A' <= c && c <= 'Z'
}

func isASCIIDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

func CamelCase(s string) string {
	t := ""
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '_' {
			continue // Skip the underscore in s.
		}
		if isASCIIDigit(c) {
			t += string(c)
			continue
		}
		if isASCIILower(c) {
			c ^= ' ' // Make it upper case
		}
		t += string(c)
		// Convert upper case to lower case following an upper case character
		for i+1 < len(s) && isASCIIUpper(s[i+1]) {
			if i+2 < len(s) && isASCIILower(s[i+2]) {
				break
			}
			i++
			t += string(s[i] ^ ' ') // Make it lower case
		}
		// Accept lower case sequence that follows.
		for i+1 < len(s) && isASCIILower(s[i+1]) {
			i++
			t += string(s[i])
		}
	}
	return t
}

func EscapeJson(jsoninput string) string {
	r := strings.NewReplacer(
		`{`, `\{`, `}`, `\}`)
	return r.Replace(jsoninput)
}
