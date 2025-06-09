package util

import (
	"encoding/json"
	"strings"
	"unicode"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IsJsonEmpty checks if a JSON string is empty or only contains whitespace
func IsJsonEmpty(s string) bool {
	s = strings.ReplaceAll(s, " ", "")
	return s == "" || s == "{}"
}

// CleanJsonWhitespaces removes new lines and redundant spaces from a JSON string
// and also removes comma from the end of the string
func CleanJsonWhitespaces(s string) string {
	s = strings.TrimSuffix(s, ",")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\t", "")

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
		} else if unicode.IsSpace(char) && prevChar != ' ' {
			result.WriteRune(char)
		}

		prevChar = char
	}

	return result.String()
}

// CleanAllWhitespaces removes all whitespaces from a string
func CleanAllWhitespaces(s string) string {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\t", "")
	return s
}

func ParseJsonArray(value string) (any, error) {
	var jsonArray []any
	if err := json.Unmarshal([]byte(value), &jsonArray); err != nil {
		return value, nil
	}

	bsonArray := make(primitive.A, len(jsonArray))
	copy(bsonArray, jsonArray)

	return bsonArray, nil
}
