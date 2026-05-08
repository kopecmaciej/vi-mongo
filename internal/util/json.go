package util

import (
	"bytes"
	"encoding/json"
	"strings"
	"unicode"
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

// InlineJson formats JSON as a single line with spaces inside delimiters:
// {"a":{"b":1}} → { "a": { "b": 1 } }
func InlineJson(s string) (string, error) {
	var compacted bytes.Buffer
	if err := json.Compact(&compacted, []byte(s)); err != nil {
		return "", err
	}

	src := compacted.Bytes()
	var result strings.Builder
	result.Grow(len(src) * 2)

	inQuotes := false
	escaped := false
	var last byte

	write := func(b byte) {
		result.WriteByte(b)
		last = b
	}

	for i := range src {
		ch := src[i]

		if escaped {
			write(ch)
			escaped = false
			continue
		}
		if ch == '\\' && inQuotes {
			write(ch)
			escaped = true
			continue
		}
		if ch == '"' {
			inQuotes = !inQuotes
			write(ch)
			continue
		}
		if inQuotes {
			write(ch)
			continue
		}

		switch ch {
		case '{', '[':
			write(ch)
			if i+1 < len(src) && src[i+1] != '}' && src[i+1] != ']' {
				write(' ')
			}
		case '}', ']':
			if last != '{' && last != '[' {
				write(' ')
			}
			write(ch)
		case ':':
			write(ch)
			write(' ')
		case ',':
			write(ch)
			write(' ')
		default:
			write(ch)
		}
	}

	return result.String(), nil
}

// CleanAllWhitespaces removes all whitespaces from a string
func CleanAllWhitespaces(s string) string {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\t", "")
	return s
}
