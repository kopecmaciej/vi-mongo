package util

import (
	"testing"
)

func TestIsHexColor(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"#fff", true},
		{"#000000", true},
		{"#abc123", true},
		{"#gggggg", false},
		{"#12345", false},
		{"ffffff", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := IsHexColor(tt.input); got != tt.want {
				t.Errorf("IsHexColor(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestQuoteUnquotedKeys(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`{key: "value"}`, `{ "key": "value"}`},
		{`{key1: "value1", key2: "value2"}`, `{ "key1": "value1", "key2": "value2"}`},
		{`{"key": "value"}`, `{"key": "value"}`},
		{`{key_123: "value", another_key: 42}`, `{ "key_123": "value", "another_key": 42}`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := QuoteUnquotedKeys(tt.input); got != tt.want {
				t.Errorf("QuoteUnquotedKeys(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTrimMultipleSpaces(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello   world", "hello world"},
		{"  leading and trailing  ", " leading and trailing "},
		{"no  multiple   spaces", "no multiple spaces"},
		{"single space", "single space"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := TrimMultipleSpaces(tt.input); got != tt.want {
				t.Errorf("TrimMultipleSpaces(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestHidePasswordInUri(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"mongodb://user:password@localhost:27017", "mongodb://user:********@localhost:27017"},
		{"mongodb://localhost:27017", "mongodb://localhost:27017"},
		{"http://example.com", "http://example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := HidePasswordInUri(tt.input); got != tt.want {
				t.Errorf("HidePasswordInUri(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseDateToBson(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{`{"date": {"$date": "2023-04-15T12:30:45Z"}}`, `{"date": {"$date":{"$numberLong":"1681561845000"}}}`, false},
		{`{"invalid": {"$date": "not-a-date"}}`, `{"invalid": {"$date": "not-a-date"}}`, true},
		{`{"normal": "field"}`, `{"normal": "field"}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDateToBson(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDateToBson(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseDateToBson(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTransformISODate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Simple ISODate",
			input: `{ createdAt: ISODate("2024-01-01T00:00:00Z") }`,
			want:  `{ createdAt: {"$date":{"$numberLong":"1704067200000"}} }`,
		},
		{
			name:  "ISODate with milliseconds",
			input: `{ timestamp: ISODate("2024-12-25T23:59:59.999Z") }`,
			want:  `{ timestamp: {"$date":{"$numberLong":"1735171199999"}} }`,
		},
		{
			name:  "Multiple ISODate fields",
			input: `{ start: ISODate("2024-01-01T00:00:00Z"), end: ISODate("2024-12-31T23:59:59Z") }`,
			want:  `{ start: {"$date":{"$numberLong":"1704067200000"}}, end: {"$date":{"$numberLong":"1735689599000"}} }`,
		},
		{
			name:  "ISODate with spaces",
			input: `{ date: ISODate( "2024-06-15T12:30:45Z" ) }`,
			want:  `{ date: {"$date":{"$numberLong":"1718454645000"}} }`,
		},
		{
			name:  "No ISODate - should remain unchanged",
			input: `{ date: "2024-01-01" }`,
			want:  `{ date: "2024-01-01" }`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TransformISODate(tt.input)
			if got != tt.want {
				t.Errorf("TransformISODate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTransformNumberInt(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Simple NumberInt",
			input: `{ age: NumberInt(30) }`,
			want:  `{ age: {"$numberInt": "30"} }`,
		},
		{
			name:  "NumberInt with zero",
			input: `{ count: NumberInt(0) }`,
			want:  `{ count: {"$numberInt": "0"} }`,
		},
		{
			name:  "Multiple NumberInt fields",
			input: `{ min: NumberInt(1), max: NumberInt(100) }`,
			want:  `{ min: {"$numberInt": "1"}, max: {"$numberInt": "100"} }`,
		},
		{
			name:  "NumberInt with spaces",
			input: `{ value: NumberInt( 42 ) }`,
			want:  `{ value: {"$numberInt": "42"} }`,
		},
		{
			name:  "No NumberInt - should remain unchanged",
			input: `{ age: 30 }`,
			want:  `{ age: 30 }`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TransformNumberInt(tt.input)
			if got != tt.want {
				t.Errorf("TransformNumberInt() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTransformNumberLong(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Simple NumberLong",
			input: `{ bigNumber: NumberLong(9223372036854775807) }`,
			want:  `{ bigNumber: {"$numberLong": "9223372036854775807"} }`,
		},
		{
			name:  "NumberLong with smaller value",
			input: `{ id: NumberLong(12345) }`,
			want:  `{ id: {"$numberLong": "12345"} }`,
		},
		{
			name:  "Multiple NumberLong fields",
			input: `{ start: NumberLong(1000), end: NumberLong(2000) }`,
			want:  `{ start: {"$numberLong": "1000"}, end: {"$numberLong": "2000"} }`,
		},
		{
			name:  "NumberLong with spaces",
			input: `{ value: NumberLong( 999999 ) }`,
			want:  `{ value: {"$numberLong": "999999"} }`,
		},
		{
			name:  "No NumberLong - should remain unchanged",
			input: `{ count: 100 }`,
			want:  `{ count: 100 }`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TransformNumberLong(tt.input)
			if got != tt.want {
				t.Errorf("TransformNumberLong() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTransformNumberDecimal(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Simple NumberDecimal",
			input: `{ price: NumberDecimal("19.99") }`,
			want:  `{ price: {"$numberDecimal": "19.99"} }`,
		},
		{
			name:  "NumberDecimal with many decimal places",
			input: `{ pi: NumberDecimal("3.14159265358979323846") }`,
			want:  `{ pi: {"$numberDecimal": "3.14159265358979323846"} }`,
		},
		{
			name:  "Multiple NumberDecimal fields",
			input: `{ price: NumberDecimal("10.50"), tax: NumberDecimal("1.05") }`,
			want:  `{ price: {"$numberDecimal": "10.50"}, tax: {"$numberDecimal": "1.05"} }`,
		},
		{
			name:  "NumberDecimal with spaces",
			input: `{ amount: NumberDecimal( "123.45" ) }`,
			want:  `{ amount: {"$numberDecimal": "123.45"} }`,
		},
		{
			name:  "NumberDecimal with integer value",
			input: `{ whole: NumberDecimal("100") }`,
			want:  `{ whole: {"$numberDecimal": "100"} }`,
		},
		{
			name:  "No NumberDecimal - should remain unchanged",
			input: `{ price: 19.99 }`,
			want:  `{ price: 19.99 }`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TransformNumberDecimal(tt.input)
			if got != tt.want {
				t.Errorf("TransformNumberDecimal() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTransformMongoshSyntax(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Combined: regex and ISODate",
			input: `{ email: /test@example\.com/i, createdAt: ISODate("2024-01-01T00:00:00Z") }`,
			want:  `{ email: { "$regex": "test@example\\.com", "$options": "i" }, createdAt: {"$date":{"$numberLong":"1704067200000"}} }`,
		},
		{
			name:  "Combined: NumberInt and NumberLong",
			input: `{ age: NumberInt(30), userId: NumberLong(123456789) }`,
			want:  `{ age: {"$numberInt": "30"}, userId: {"$numberLong": "123456789"} }`,
		},
		{
			name:  "Combined: all transformations",
			input: `{ name: /^john/i, age: NumberInt(25), balance: NumberDecimal("1000.50"), createdAt: ISODate("2024-01-01T00:00:00Z"), views: NumberLong(999999) }`,
			want:  `{ name: { "$regex": "^john", "$options": "i" }, age: {"$numberInt": "25"}, balance: {"$numberDecimal": "1000.50"}, createdAt: {"$date":{"$numberLong":"1704067200000"}}, views: {"$numberLong": "999999"} }`,
		},
		{
			name:  "No transformations needed",
			input: `{ name: "John", age: 30, active: true }`,
			want:  `{ name: "John", age: 30, active: true }`,
		},
		{
			name:  "Nested objects with transformations",
			input: `{ user: { name: /admin/i, created: ISODate("2024-01-01T00:00:00Z") } }`,
			want:  `{ user: { name: { "$regex": "admin", "$options": "i" }, created: {"$date":{"$numberLong":"1704067200000"}} } }`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TransformMongoshSyntax(tt.input)
			if got != tt.want {
				t.Errorf("TransformMongoshSyntax() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTransformRegexShorthand(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Simple regex without flags",
			input: `{ email: /example\.com$/ }`,
			want:  `{ email: { "$regex": "example\\.com$" } }`,
		},
		{
			name:  "Regex with case-insensitive flag",
			input: `{ name: /^john/i }`,
			want:  `{ name: { "$regex": "^john", "$options": "i" } }`,
		},
		{
			name:  "Regex with multiple flags",
			input: `{ text: /pattern/gim }`,
			want:  `{ text: { "$regex": "pattern", "$options": "gim" } }`,
		},
		{
			name:  "Regex with escaped forward slashes",
			input: `{ path: /\/api\/v1\// }`,
			want:  `{ path: { "$regex": "\\/api\\/v1\\/" } }`,
		},
		{
			name:  "Multiple regex patterns in one query",
			input: `{ name: /john/, email: /gmail\.com/ }`,
			want:  `{ name: { "$regex": "john" }, email: { "$regex": "gmail\\.com" } }`,
		},
		{
			name:  "Regex in nested object",
			input: `{ user: { name: /test/i } }`,
			want:  `{ user: { name: { "$regex": "test", "$options": "i" } } }`,
		},
		{
			name:  "Complex regex pattern",
			input: `{ email: /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/i }`,
			want:  `{ email: { "$regex": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", "$options": "i" } }`,
		},
		{
			name:  "Regex with special characters",
			input: `{ field: /\d{3}-\d{3}-\d{4}/ }`,
			want:  `{ field: { "$regex": "\\d{3}-\\d{3}-\\d{4}" } }`,
		},
		{
			name:  "Mixed query with regex and regular fields",
			input: `{ email: /example\.com$/, status: "active", age: 25 }`,
			want:  `{ email: { "$regex": "example\\.com$" }, status: "active", age: 25 }`,
		},
		{
			name:  "No regex pattern - should remain unchanged",
			input: `{ name: "john", email: "test@example.com" }`,
			want:  `{ name: "john", email: "test@example.com" }`,
		},
		{
			name:  "Regex with dot and star quantifier",
			input: `{ description: /.*important.*/ }`,
			want:  `{ description: { "$regex": ".*important.*" } }`,
		},
		{
			name:  "Regex with word boundaries",
			input: `{ word: /\btest\b/i }`,
			want:  `{ word: { "$regex": "\\btest\\b", "$options": "i" } }`,
		},
		{
			name:  "Regex with character classes",
			input: `{ code: /[A-Z]{2}[0-9]{4}/ }`,
			want:  `{ code: { "$regex": "[A-Z]{2}[0-9]{4}" } }`,
		},
		{
			name:  "Quoted keys with regex",
			input: `{ "email": /test@/i }`,
			want:  `{ "email": { "$regex": "test@", "$options": "i" } }`,
		},
		{
			name:  "Regex with parentheses for grouping",
			input: `{ url: /(http|https):\/\/[^\s]+/ }`,
			want:  `{ url: { "$regex": "(http|https):\\/\\/[^\\s]+" } }`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TransformRegexShorthand(tt.input)
			if got != tt.want {
				t.Errorf("TransformRegexShorthand() = %q, want %q", got, tt.want)
			}
		})
	}
}
