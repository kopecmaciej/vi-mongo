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

func TestInlineJson(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Flat object",
			input:    `{"$match":{"interval":60}}`,
			expected: `{ "$match": { "interval": 60 } }`,
		},
		{
			name:     "Already pretty printed",
			input:    "{\n  \"$match\": {\n    \"interval\": 60\n  }\n}",
			expected: `{ "$match": { "interval": 60 } }`,
		},
		{
			name:     "Empty object",
			input:    `{}`,
			expected: `{}`,
		},
		{
			name:     "Empty array",
			input:    `[]`,
			expected: `[]`,
		},
		{
			name:     "Array of values",
			input:    `[1,2,3]`,
			expected: `[ 1, 2, 3 ]`,
		},
		{
			name:     "Nested empty",
			input:    `{"a":{}}`,
			expected: `{ "a": {} }`,
		},
		{
			name:     "Spaces in string values preserved",
			input:    `{"key":"value with spaces"}`,
			expected: `{ "key": "value with spaces" }`,
		},
		{
			name:     "Escaped quotes in strings",
			input:    `{"key":"say \"hi\""}`,
			expected: `{ "key": "say \"hi\"" }`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := InlineJson(tc.input)
			assert.NoError(t, err)
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
			input:    "{\n\"key1\": \"value1\",\n  \"key2\": \"value2\"\n}",
			expected: `{"key1": "value1", "key2": "value2"}`,
		},
		{
			name:     "Remove trailing comma",
			input:    `{"key": "value"},`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "Preserve spaces in quotes",
			input:    `{"key 1": "value with spaces"}`,
			expected: `{"key 1": "value with spaces"}`,
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
			expected: `{"key1": "value1","key2": [1, 2, 3],"key3": {"nested": "object"},"key4": "value with \\"quotes\\""}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CleanJsonWhitespaces(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
