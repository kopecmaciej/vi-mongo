package util

import (
	"regexp"
	"strings"
	"unicode"
)

// TrimJson removes whitespace from a JSON string, except within quotes
// and also removes comma from the end of the string
func TrimJson(s string) string {
	s = strings.TrimSuffix(s, ",")

	var result strings.Builder
	inQuotes := false
	prevChar := ' '

	// remove whitespace from a JSON string, except within quotes
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

// TrimMultipleSpaces trims multiple spaces into a single space
func TrimMultipleSpaces(s string) string {
	// Then, replace multiple spaces with a single space
	space := regexp.MustCompile(`\s+`)
	return space.ReplaceAllString(s, " ")
}

// HidePasswordInUri redacts the password in a connection string
func HidePasswordInUri(s string) string {
	re := regexp.MustCompile(`://([^:]+):([^@]+)(@.*)`)

	return re.ReplaceAllString(s, "://$1:********$3")
}
