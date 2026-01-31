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

func TestTransformMongoshSyntax(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// ObjectId
		{
			name:  "ObjectID uppercase",
			input: `{ _id: ObjectID("507f1f77bcf86cd799439011") }`,
			want:  `{ _id: {"$oid": "507f1f77bcf86cd799439011"} }`,
		},
		{
			name:  "ObjectId lowercase d",
			input: `{ _id: ObjectId("507f1f77bcf86cd799439011") }`,
			want:  `{ _id: {"$oid": "507f1f77bcf86cd799439011"} }`,
		},
		{
			name:  "ObjectId with spaces",
			input: `{ _id: ObjectId( "507f1f77bcf86cd799439011" ) }`,
			want:  `{ _id: {"$oid": "507f1f77bcf86cd799439011"} }`,
		},
		// ISODate
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
			name:  "ISODate date only",
			input: `{ createdAt: ISODate("2025-11-16") }`,
			want:  `{ createdAt: {"$date":{"$numberLong":"1763251200000"}} }`,
		},
		{
			name:  "ISODate without timezone",
			input: `{ createdAt: ISODate("2024-07-04T05:50:15") }`,
			want:  `{ createdAt: {"$date":{"$numberLong":"1720072215000"}} }`,
		},
		// NumberInt
		{
			name:  "Simple NumberInt",
			input: `{ age: NumberInt(30) }`,
			want:  `{ age: {"$numberInt": "30"} }`,
		},
		{
			name:  "Multiple NumberInt fields",
			input: `{ min: NumberInt(0), max: NumberInt(100) }`,
			want:  `{ min: {"$numberInt": "0"}, max: {"$numberInt": "100"} }`,
		},
		{
			name:  "NumberInt with spaces",
			input: `{ value: NumberInt( 42 ) }`,
			want:  `{ value: {"$numberInt": "42"} }`,
		},
		// NumberLong
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
		// NumberDecimal
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
		// Combined
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
			input: `{ _id: ObjectId("507f1f77bcf86cd799439011"), name: /^john/i, age: NumberInt(25), balance: NumberDecimal("1000.50"), createdAt: ISODate("2024-01-01T00:00:00Z"), views: NumberLong(999999) }`,
			want:  `{ _id: {"$oid": "507f1f77bcf86cd799439011"}, name: { "$regex": "^john", "$options": "i" }, age: {"$numberInt": "25"}, balance: {"$numberDecimal": "1000.50"}, createdAt: {"$date":{"$numberLong":"1704067200000"}}, views: {"$numberLong": "999999"} }`,
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
			got, err := TransformMongoshSyntax(tt.input)
			if err != nil {
				t.Fatalf("TransformMongoshSyntax() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("TransformMongoshSyntax() = %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("Invalid ISODate returns error", func(t *testing.T) {
		_, err := TransformMongoshSyntax(`{ created_at: ISODate("not-a-date") }`)
		if err == nil {
			t.Error("TransformMongoshSyntax() expected error for invalid ISODate, got nil")
		}
	})
}

func TestTransformRegexShorthand(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Simple regex without flags",
			input: `{ email: /website\.com$/ }`,
			want:  `{ email: { "$regex": "website\\.com$" } }`,
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
		{
			name:  "Regex in array - single element",
			input: `{ tags: [/tag1/] }`,
			want:  `{ tags: [{ "$regex": "tag1" }] }`,
		},
		{
			name:  "Regex in array - multiple elements without flags",
			input: `{ tags: [/tag1/, /tag2/] }`,
			want:  `{ tags: [{ "$regex": "tag1" }, { "$regex": "tag2" }] }`,
		},
		{
			name:  "Regex in array - multiple elements with flags",
			input: `{ tags: [/tag1/i, /tag2/gi] }`,
			want:  `{ tags: [{ "$regex": "tag1", "$options": "i" }, { "$regex": "tag2", "$options": "gi" }] }`,
		},
		{
			name:  "Regex in array with mixed flags",
			input: `{ patterns: [/^start/, /end$/i, /middle/gim] }`,
			want:  `{ patterns: [{ "$regex": "^start" }, { "$regex": "end$", "$options": "i" }, { "$regex": "middle", "$options": "gim" }] }`,
		},
		{
			name:  "Regex with $in operator",
			input: `{ email: { $in: [/gmail\.com/, /yahoo\.com/i] } }`,
			want:  `{ email: { $in: [{ "$regex": "gmail\\.com" }, { "$regex": "yahoo\\.com", "$options": "i" }] } }`,
		},
		{
			name:  "Regex in nested array",
			input: `{ filters: [{ name: /john/i }, { email: /gmail/ }] }`,
			want:  `{ filters: [{ name: { "$regex": "john", "$options": "i" } }, { email: { "$regex": "gmail" } }] }`,
		},
		{
			name:  "Array with regex and other values mixed",
			input: `{ tags: [/pattern/i, "literal", /another/] }`,
			want:  `{ tags: [{ "$regex": "pattern", "$options": "i" }, "literal", { "$regex": "another" }] }`,
		},
		{
			name:  "Regex in array with escaped characters",
			input: `{ paths: [/\/api\/v1\//i, /\/api\/v2\//] }`,
			want:  `{ paths: [{ "$regex": "\\/api\\/v1\\/", "$options": "i" }, { "$regex": "\\/api\\/v2\\/" }] }`,
		},
		{
			name:  "Regex in array with complex patterns",
			input: `{ emails: [/^[a-z0-9]+@gmail\.com$/i, /^[a-z0-9]+@yahoo\.com$/i] }`,
			want:  `{ emails: [{ "$regex": "^[a-z0-9]+@gmail\\.com$", "$options": "i" }, { "$regex": "^[a-z0-9]+@yahoo\\.com$", "$options": "i" }] }`,
		},
		{
			name:  "Array with spaces around regex",
			input: `{ tags: [ /tag1/i , /tag2/ ] }`,
			want:  `{ tags: [ { "$regex": "tag1", "$options": "i" } , { "$regex": "tag2" } ] }`,
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
