package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsJsonEmpty(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Empty string", "", true},
		{"Whitespace only", "   ", true},
		{"Empty object", "{}", true},
		{"Whitespace with empty object", "  {}  ", true},
		{"Non-empty object", `{"key": "value"}`, false},
		{"Non-empty array", "[1, 2, 3]", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsJsonEmpty(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCleanJsonWhitespaces(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Remove newlines and extra spaces",
			input:    "{\n  \"key1\": \"value1\",\n  \"key2\": \"value2\"\n}",
			expected: `{ "key1": "value1", "key2": "value2" }`,
		},
		{
			name:     "Remove trailing comma",
			input:    `{"key": "value"},`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "Preserve spaces in quotes",
			input:    `{"key": "value with spaces"}`,
			expected: `{"key": "value with spaces"}`,
		},
		{
			name: "Complex JSON",
			input: `{
				"key1": "value1",
				"key2": [1, 2, 3],
				"key3": {
					"nested": "object"
				},
				"key4": "value with \\"quotes\\""
			}`,
			expected: `{ "key1": "value1", "key2": [1, 2, 3], "key3": { "nested": "object" }, "key4": "value with \\"quotes\\"" }`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CleanJsonWhitespaces(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
