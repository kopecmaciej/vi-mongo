package utils

import (
	"regexp"
	"strings"
	"unicode"
)

// TrimJson removes whitespace from a JSON string, except within quotes
func TrimJson(s string) string {
	var result strings.Builder
	inQuotes := false
	prevChar := ' '

	for _, char := range s {
		if char == '"' && prevChar != '\\' {
			inQuotes = !inQuotes
		}

		if inQuotes {
			result.WriteRune(char)
		} else if !unicode.IsSpace(char) {
			result.WriteRune(char)
		}

		prevChar = char
	}

	return result.String()
}

// QuoteUnquotedKeys adds quotes to unquoted keys in a JSON-like string
func QuoteUnquotedKeys(s string) string {
	re := regexp.MustCompile(`(\{|\,)\s*([a-zA-Z0-9_.]+)\s*:`)
	return re.ReplaceAllString(s, `$1 "$2":`)
}