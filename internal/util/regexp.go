package util

import (
	"fmt"
	"regexp"
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
