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
