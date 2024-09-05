package util

import (
	"regexp"
)

var (
	unquotedKeysRegex       *regexp.Regexp
	multipleSpacesRegex     *regexp.Regexp
	uriPasswordRegex        *regexp.Regexp
	jsonKeyValuePairRegex   *regexp.Regexp
	keyWithIndentationRegex *regexp.Regexp
)

func init() {
	unquotedKeysRegex = regexp.MustCompile(`(\{|\,)\s*([a-zA-Z\d()!@#$%&*._]+)\s*:`)
	multipleSpacesRegex = regexp.MustCompile(`\s+`)
	uriPasswordRegex = regexp.MustCompile(`://([^:]+):([^@]+)(@.*)`)
	jsonKeyValuePairRegex = regexp.MustCompile(`"([^"]+)":(.*)`)
	keyWithIndentationRegex = regexp.MustCompile(`^\s*"([^"]+)":\s*(.*)$`)
}

// QuoteUnquotedKeys adds quotes to unquoted keys in a JSON-like string
func QuoteUnquotedKeys(s string) string {
	return unquotedKeysRegex.ReplaceAllString(s, `$1 "$2":`)
}

// TrimMultipleSpaces trims multiple spaces into a single space
func TrimMultipleSpaces(s string) string {
	// Then, replace multiple spaces with a single space
	return multipleSpacesRegex.ReplaceAllString(s, " ")
}

// HidePasswordInUri redacts the password in a connection string
func HidePasswordInUri(s string) string {
	return uriPasswordRegex.ReplaceAllString(s, "://$1:********$3")
}
