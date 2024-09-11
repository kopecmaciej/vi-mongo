package util

import (
	"fmt"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	unquotedKeysRegex   = regexp.MustCompile(`(\{|\,)\s*([a-zA-Z\d()!@#$%&*._]+)\s*:`)
	multipleSpacesRegex = regexp.MustCompile(`\s+`)
	uriPasswordRegex    = regexp.MustCompile(`://([^:]+):([^@]+)(@.*)`)
	hexColorRegex       = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}){1,2}$`)
	dateRegex           = regexp.MustCompile(`\{\s*\"\$date\"\s*:\s*\"(.*?)\"\s*\}`)
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
		return s, parseError
	}
	return query, nil
}
