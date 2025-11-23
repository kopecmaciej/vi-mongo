package util

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	unquotedKeysRegex   = regexp.MustCompile(`(\{|\,)\s*([a-zA-Z\d()!@#$%&*._]+)\s*:`)
	multipleSpacesRegex = regexp.MustCompile(`\s+`)
	uriPasswordRegex    = regexp.MustCompile(`://([^:]+):([^@]+)(@.*)`)
	hexColorRegex       = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}){1,2}$`)
	dateRegex           = regexp.MustCompile(`\{\s*\"\$date\"\s*:\s*\"(.*?)\"\s*\}`)
	// Matches regex literals in format: /pattern/flags
	// Group 1: everything before the regex (: or [ or ,)
	// Group 2: the flags (optional)
	// Supports regex after: field values, array starts, and array commas
	regexLiteralPattern = regexp.MustCompile(`([:\[,]\s*)/(?:\\.|[^/\\])+/([gimsx]*)`)
	// Mongosh helper function patterns
	isoDatePattern       = regexp.MustCompile(`ISODate\s*\(\s*"([^"]*)"\s*\)`)
	numberIntPattern     = regexp.MustCompile(`NumberInt\s*\(\s*(\d+)\s*\)`)
	numberLongPattern    = regexp.MustCompile(`NumberLong\s*\(\s*(\d+)\s*\)`)
	numberDecimalPattern = regexp.MustCompile(`NumberDecimal\s*\(\s*"([^"]*)"\s*\)`)
)

// IsHexColor checks if a string is a valid hex color
func IsHexColor(s string) bool {
	return hexColorRegex.MatchString(s)
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

// RestorePasswordInUri replaces the masked password in a safe URI with the actual password
func RestorePasswordInUri(safeURI string, password string) string {
	// Define a regex to match the masked password pattern
	maskedPasswordRegex := regexp.MustCompile(`://([^:]+):(\*+)(@.*)`)

	// Replace the masked password with the actual password
	return maskedPasswordRegex.ReplaceAllString(safeURI, fmt.Sprintf("://$1:%s$3", password))
}

// ParseDateToBson parses a date in a JSON string into a BSON date
func ParseDateToBson(s string) (string, error) {
	var parseError error
	query := dateRegex.ReplaceAllStringFunc(s, func(match string) string {
		dateStr := dateRegex.FindStringSubmatch(match)[1]
		t, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			parseError = err
			return match
		}
		return fmt.Sprintf(`{"$date":{"$numberLong":"%d"}}`, primitive.NewDateTimeFromTime(t).Time().UnixMilli())
	})
	if parseError != nil {
		log.Error().Err(parseError).Msg("Failed to parse date string to BSON")
		return s, parseError
	}
	return query, nil
}

// TransformRegexShorthand converts JavaScript-style regex literals to MongoDB $regex syntax
// Example: { email: /example\.com$/ } -> { email: { "$regex": "example\\.com$" } }
// Example: { name: /^john/i } -> { name: { "$regex": "^john", "$options": "i" } }
func TransformRegexShorthand(s string) string {
	return regexLiteralPattern.ReplaceAllStringFunc(s, func(match string) string {
		// Extract the full match details
		submatches := regexLiteralPattern.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}

		prefix := submatches[1] // The part before the regex (e.g., ": ")
		flags := submatches[2]  // The flags (e.g., "i", "gim")

		// Extract the pattern from the match
		// Find the start of the regex pattern (after prefix and first /)
		patternStart := len(prefix) + 1
		// Find the end (before the last / and flags)
		patternEnd := len(match) - len(flags) - 1

		if patternStart >= patternEnd {
			return match
		}

		pattern := match[patternStart:patternEnd]

		// Escape backslashes first (for JSON), then escape double quotes
		pattern = strings.ReplaceAll(pattern, `\`, `\\`)
		pattern = strings.ReplaceAll(pattern, `"`, `\"`)

		// Build the replacement string
		if flags == "" {
			return fmt.Sprintf(`%s{ "$regex": "%s" }`, prefix, pattern)
		}
		return fmt.Sprintf(`%s{ "$regex": "%s", "$options": "%s" }`, prefix, pattern, flags)
	})
}

func TransformISODate(s string) string {
	return isoDatePattern.ReplaceAllStringFunc(s, func(match string) string {
		dateStr := isoDatePattern.FindStringSubmatch(match)[1]
		t, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			// If parsing fails, return original match
			return match
		}
		millis := primitive.NewDateTimeFromTime(t).Time().UnixMilli()
		return fmt.Sprintf(`{"$date":{"$numberLong":"%d"}}`, millis)
	})
}

func TransformNumberInt(s string) string {
	return numberIntPattern.ReplaceAllString(s, `{"$$numberInt": "$1"}`)
}

func TransformNumberLong(s string) string {
	return numberLongPattern.ReplaceAllString(s, `{"$$numberLong": "$1"}`)
}

func TransformNumberDecimal(s string) string {
	return numberDecimalPattern.ReplaceAllString(s, `{"$$numberDecimal": "$1"}`)
}

func TransformMongoshSyntax(s string) string {
	s = TransformRegexShorthand(s)
	s = TransformISODate(s)
	s = TransformNumberInt(s)
	s = TransformNumberLong(s)
	s = TransformNumberDecimal(s)
	return s
}
