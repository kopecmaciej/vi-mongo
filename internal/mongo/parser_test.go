package mongo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// mustParseDecimal is a helper function for tests that need decimal128 values
func mustParseDecimal(s string) primitive.Decimal128 {
	d, err := primitive.ParseDecimal128(s)
	if err != nil {
		panic(fmt.Sprintf("failed to parse decimal: %v", err))
	}
	return d
}

func TestParseStringQuery(t *testing.T) {
	objectID, err := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	assert.NoError(t, err, "Failed to create ObjectID for testing")

	cases := []struct {
		name     string
		input    string
		expected map[string]any
		hasError bool
	}{
		{
			name:     "Empty input",
			input:    "",
			expected: map[string]any{},
			hasError: false,
		},
		{
			name:     "Valid input with ObjectID",
			input:    `{_id: ObjectID("507f1f77bcf86cd799439011")}`,
			expected: map[string]any{"_id": objectID},
			hasError: false,
		},
		{
			name:     "Multiple fields with nested document",
			input:    `{ _id: ObjectID("507f1f77bcf86cd799439011"), user: { name: "John", age: 30 } }`,
			expected: map[string]any{"_id": objectID, "user": primitive.M{"name": "John", "age": int32(30)}},
			hasError: false,
		},
		{
			name:     "Array and date",
			input:    `{ tags: ["mongodb", "database"], createdAt: { $date: "2023-04-15T12:00:00Z" } }`,
			expected: map[string]any{"tags": primitive.A{"mongodb", "database"}, "createdAt": primitive.NewDateTimeFromTime(time.Date(2023, 4, 15, 12, 0, 0, 0, time.UTC))},
			hasError: false,
		},
		{
			name:     "Invalid ObjectID",
			input:    `{"_id": ObjectID("invalid")}`,
			expected: nil,
			hasError: true,
		},
		{
			name:     "Regex shorthand without flags",
			input:    `{ email: /example\.com$/ }`,
			expected: map[string]any{"email": primitive.M{"$regex": "example\\.com$"}},
			hasError: false,
		},
		{
			name:     "Regex shorthand with case-insensitive flag",
			input:    `{ name: /^john/i }`,
			expected: map[string]any{"name": primitive.M{"$regex": "^john", "$options": "i"}},
			hasError: false,
		},
		{
			name:     "Regex shorthand with multiple flags",
			input:    `{ text: /pattern/gim }`,
			expected: map[string]any{"text": primitive.M{"$regex": "pattern", "$options": "gim"}},
			hasError: false,
		},
		{
			name:     "Multiple regex patterns in one query",
			input:    `{ name: /john/, email: /gmail\.com/i }`,
			expected: map[string]any{"name": primitive.M{"$regex": "john"}, "email": primitive.M{"$regex": "gmail\\.com", "$options": "i"}},
			hasError: false,
		},
		{
			name:     "Regex with ObjectID in same query",
			input:    `{ _id: ObjectID("507f1f77bcf86cd799439011"), email: /test@example\.com/ }`,
			expected: map[string]any{"_id": objectID, "email": primitive.M{"$regex": "test@example\\.com"}},
			hasError: false,
		},
		{
			name:     "Nested object with regex",
			input:    `{ user: { email: /admin@/i } }`,
			expected: map[string]any{"user": primitive.M{"email": primitive.M{"$regex": "admin@", "$options": "i"}}},
			hasError: false,
		},
		{
			name:     "Complex regex pattern",
			input:    `{ email: /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/i }`,
			expected: map[string]any{"email": primitive.M{"$regex": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", "$options": "i"}},
			hasError: false,
		},
		{
			name:     "ISODate syntax",
			input:    `{ createdAt: ISODate("2024-01-01T00:00:00Z") }`,
			expected: map[string]any{"createdAt": primitive.NewDateTimeFromTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))},
			hasError: false,
		},
		{
			name:     "NumberInt syntax",
			input:    `{ age: NumberInt(30) }`,
			expected: map[string]any{"age": int32(30)},
			hasError: false,
		},
		{
			name:     "NumberLong syntax",
			input:    `{ userId: NumberLong(123456789) }`,
			expected: map[string]any{"userId": int64(123456789)},
			hasError: false,
		},
		{
			name:     "NumberDecimal syntax",
			input:    `{ price: NumberDecimal("19.99") }`,
			expected: map[string]any{"price": mustParseDecimal("19.99")},
			hasError: false,
		},
		{
			name:     "Combined mongosh syntax",
			input:    `{ name: /^john/i, age: NumberInt(25), createdAt: ISODate("2024-01-01T00:00:00Z") }`,
			expected: map[string]any{"name": primitive.M{"$regex": "^john", "$options": "i"}, "age": int32(25), "createdAt": primitive.NewDateTimeFromTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))},
			hasError: false,
		},
		{
			name:     "Multiple ISODate fields",
			input:    `{ start: ISODate("2024-01-01T00:00:00Z"), end: ISODate("2024-12-31T23:59:59Z") }`,
			expected: map[string]any{"start": primitive.NewDateTimeFromTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)), "end": primitive.NewDateTimeFromTime(time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC))},
			hasError: false,
		},
		{
			name:     "ObjectId lowercase variant",
			input:    `{ _id: ObjectId("507f1f77bcf86cd799439011") }`,
			expected: map[string]any{"_id": objectID},
			hasError: false,
		},
		{
			name:     "Regex in array - single element",
			input:    `{ tags: [/mongodb/] }`,
			expected: map[string]any{"tags": primitive.A{primitive.M{"$regex": "mongodb"}}},
			hasError: false,
		},
		{
			name:     "Regex in array - multiple elements",
			input:    `{ tags: [/mongodb/i, /database/] }`,
			expected: map[string]any{"tags": primitive.A{primitive.M{"$regex": "mongodb", "$options": "i"}, primitive.M{"$regex": "database"}}},
			hasError: false,
		},
		{
			name:     "Regex with $in operator",
			input:    `{ email: { $in: [/gmail\.com/i, /yahoo\.com/] } }`,
			expected: map[string]any{"email": primitive.M{"$in": primitive.A{primitive.Regex{Pattern: "gmail\\.com", Options: "i"}, primitive.Regex{Pattern: "yahoo\\.com", Options: ""}}}},
			hasError: false,
		},
		{
			name:     "Array with mixed types including regex",
			input:    `{ status: ["active", /pending/i, "inactive"] }`,
			expected: map[string]any{"status": primitive.A{"active", primitive.M{"$regex": "pending", "$options": "i"}, "inactive"}},
			hasError: false,
		},
		{
			name:     "Nested array with regex and other mongosh syntax",
			input:    `{ filters: [{ name: /john/i, age: NumberInt(30) }] }`,
			expected: map[string]any{"filters": primitive.A{primitive.M{"name": primitive.M{"$regex": "john", "$options": "i"}, "age": int32(30)}}},
			hasError: false,
		},
		{
			name:     "Complex query with $or and array regex",
			input:    `{ $or: [{ email: /gmail/i }, { tags: [/mongo/i, /db/] }] }`,
			expected: map[string]any{"$or": primitive.A{primitive.M{"email": primitive.M{"$regex": "gmail", "$options": "i"}}, primitive.M{"tags": primitive.A{primitive.M{"$regex": "mongo", "$options": "i"}, primitive.M{"$regex": "db"}}}}},
			hasError: false,
		},
		{
			name:     "MongoDB docs example: $in with multiple regex patterns",
			input:    `{ name: { $in: [ /^acme/i, /^ack/ ] } }`,
			expected: map[string]any{"name": primitive.M{"$in": primitive.A{primitive.Regex{Pattern: "^acme", Options: "i"}, primitive.Regex{Pattern: "^ack", Options: ""}}}},
			hasError: false,
		},
		{
			name:     "$in with regex and literal strings mixed",
			input:    `{ status: { $in: [ "active", /pend.*ing/i, "closed" ] } }`,
			expected: map[string]any{"status": primitive.M{"$in": primitive.A{"active", primitive.Regex{Pattern: "pend.*ing", Options: "i"}, "closed"}}},
			hasError: false,
		},
		{
			name:     "$and with $in containing regex patterns",
			input:    `{ $and: [ { name: { $in: [ /acme.*corp/i ] } }, { status: "active" } ] }`,
			expected: map[string]any{"$and": primitive.A{primitive.M{"name": primitive.M{"$in": primitive.A{primitive.Regex{Pattern: "acme.*corp", Options: "i"}}}}, primitive.M{"status": "active"}}},
			hasError: false,
		},
		{
			name:     "ISODate date only format",
			input:    `{ created_at: { $gt: ISODate("2025-11-16") } }`,
			expected: map[string]any{"created_at": primitive.M{"$gt": primitive.NewDateTimeFromTime(time.Date(2025, 11, 16, 0, 0, 0, 0, time.UTC))}},
			hasError: false,
		},
		{
			name:     "Invalid ISODate format",
			input:    `{ created_at: ISODate("not-a-date") }`,
			expected: nil,
			hasError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseStringQuery(tc.input)
			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestParseJsonToBson(t *testing.T) {
	objectID, err := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	assert.NoError(t, err, "Failed to create ObjectID for testing")

	cases := []struct {
		name     string
		input    string
		expected primitive.M
		hasError bool
	}{
		{
			name:     "Valid JSON with ObjectID",
			input:    `{"_id": {"$oid": "507f1f77bcf86cd799439011"}, "name": "John"}`,
			expected: primitive.M{"_id": objectID, "name": "John"},
			hasError: false,
		},
		{
			name:     "Valid JSON with Date",
			input:    `{"createdAt": {"$date": "2024-01-01T00:00:00Z"}}`,
			expected: primitive.M{"createdAt": primitive.NewDateTimeFromTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))},
			hasError: false,
		},
		{
			name:     "Invalid JSON",
			input:    `{"invalid": json}`,
			expected: nil,
			hasError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseJsonToBson(tc.input)
			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestParseBsonDocument(t *testing.T) {
	objectID, err := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	assert.NoError(t, err, "Failed to create ObjectID for testing")

	input := map[string]any{
		"_id":  objectID,
		"name": "Mark Twain",
		"age":  60,
		"tags": []string{"mongodb", "database"},
	}

	expected := fmt.Sprintf(`{"_id":{"$oid":"%s"},"age":60,"name":"Mark Twain","tags":["mongodb","database"]}`, objectID.Hex())

	result, err := ParseBsonDocument(input)
	assert.NoError(t, err)
	assert.Equal(t, result, expected)
}

func TestSortDocumentKeys(t *testing.T) {
	cases := []struct {
		name     string
		input    primitive.M
		expected string
		hasError bool
	}{
		{
			name: "Nested documents and arrays",
			input: primitive.M{
				"name": "John Doe",
				"contacts": primitive.A{
					primitive.M{"phone": "123-456-789", "email": "john@example.com"},
					primitive.M{"website": "john.com", "social": "twitter.com/john"},
				},
				"address": primitive.M{
					"zip":    "12345",
					"street": "Main St",
					"city":   "New York",
				},
				"metadata": primitive.A{
					primitive.M{
						"settings": primitive.M{
							"theme":  "dark",
							"active": true,
						},
					},
				},
			},
			expected: `{
  "address": {
    "city": "New York",
    "street": "Main St",
    "zip": "12345"
  },
  "contacts": [
    {
      "email": "john@example.com",
      "phone": "123-456-789"
    },
    {
      "social": "twitter.com/john",
      "website": "john.com"
    }
  ],
  "metadata": [
    {
      "settings": {
        "active": true,
        "theme": "dark"
      }
    }
  ],
  "name": "John Doe"
}`,
			hasError: false,
		},
		{
			name: "Deep nested arrays with user data",
			input: primitive.M{
				"users": primitive.A{
					primitive.M{
						"permissions": primitive.A{
							primitive.M{"role": "admin", "level": 3, "access": true},
							primitive.M{"role": "user", "level": 1, "access": true},
						},
					},
				},
				"version": "1.0.0",
			},
			expected: `{
  "users": [
    {
      "permissions": [
        {
          "access": true,
          "level": 3,
          "role": "admin"
        },
        {
          "access": true,
          "level": 1,
          "role": "user"
        }
      ]
    }
  ],
  "version": "1.0.0"
}`,
			hasError: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseBsonDocument(tc.input)
			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				var prettyResult bytes.Buffer
				err = json.Indent(&prettyResult, []byte(result), "", "  ")
				assert.NoError(t, err)

				assert.Equal(t, tc.expected, prettyResult.String())
			}
		})
	}
}

func TestParseValueByType(t *testing.T) {
	cases := []struct {
		name          string
		value         string
		originalValue any
		expected      any
		hasError      bool
	}{
		// Test cases with originalValue
		{
			name:          "Original value as map - valid JSON",
			value:         `{"key": "value"}`,
			originalValue: map[string]any{},
			expected:      primitive.M{"key": "value"},
			hasError:      false,
		},
		{
			name:          "Original value as primitive.M - valid JSON",
			value:         `{"key": "value"}`,
			originalValue: primitive.M{},
			expected:      primitive.M{"key": "value"},
			hasError:      false,
		},
		{
			name:          "Original value as array - valid JSON array",
			value:         `["item1", "item2"]`,
			originalValue: []any{},
			expected:      primitive.A{"item1", "item2"},
			hasError:      false,
		},
		{
			name:          "Original value as primitive.A - valid JSON array",
			value:         `["item1", "item2"]`,
			originalValue: primitive.A{},
			expected:      primitive.A{"item1", "item2"},
			hasError:      false,
		},
		{
			name:          "Original value as primitive.A - valid array with nested object",
			value:         `[{"key": "value"}, {"key2": "value2"}]`,
			originalValue: []any{},
			expected:      primitive.A{map[string]any{"key": "value"}, map[string]any{"key2": "value2"}},
			hasError:      false,
		},
		{
			name:          "Original value as int - valid int string",
			value:         "42",
			originalValue: int(0),
			expected:      int64(42),
			hasError:      false,
		},
		{
			name:          "Original value as int64 - valid int string",
			value:         "42",
			originalValue: int64(0),
			expected:      int64(42),
			hasError:      false,
		},
		{
			name:          "Original value as float64 - valid float string",
			value:         "3.14",
			originalValue: float64(0),
			expected:      float64(3.14),
			hasError:      false,
		},
		{
			name:          "Original value as bool - valid bool string",
			value:         "true",
			originalValue: bool(false),
			expected:      true,
			hasError:      false,
		},

		// Test cases without originalValue
		{
			name:          "JSON object without originalValue",
			value:         `{"key": "value"}`,
			originalValue: nil,
			expected:      primitive.M{"key": "value"},
			hasError:      false,
		},
		{
			name:          "JSON array without originalValue",
			value:         `["item1", "item2"]`,
			originalValue: nil,
			expected:      primitive.A{"item1", "item2"},
			hasError:      false,
		},
		{
			name:          "Boolean string without originalValue",
			value:         "true",
			originalValue: nil,
			expected:      true,
			hasError:      false,
		},
		{
			name:          "Integer string without originalValue",
			value:         "42",
			originalValue: nil,
			expected:      int64(42),
			hasError:      false,
		},
		{
			name:          "Float string without originalValue",
			value:         "3.14",
			originalValue: nil,
			expected:      float64(3.14),
			hasError:      false,
		},
		{
			name:          "Plain string value",
			value:         "hello",
			originalValue: nil,
			expected:      "hello",
			hasError:      false,
		},
		{
			name:          "DateTime - RFC3339 format",
			value:         "2024-07-04T05:50:15Z",
			originalValue: primitive.NewDateTimeFromTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			expected:      primitive.NewDateTimeFromTime(time.Date(2024, 7, 4, 5, 50, 15, 0, time.UTC)),
			hasError:      false,
		},
		{
			name:          "DateTime - date only format",
			value:         "2024-07-04",
			originalValue: primitive.NewDateTimeFromTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			expected:      primitive.NewDateTimeFromTime(time.Date(2024, 7, 4, 0, 0, 0, 0, time.UTC)),
			hasError:      false,
		},
		{
			name:          "DateTime - without timezone",
			value:         "2024-07-04T05:50:15",
			originalValue: primitive.NewDateTimeFromTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			expected:      primitive.NewDateTimeFromTime(time.Date(2024, 7, 4, 5, 50, 15, 0, time.UTC)),
			hasError:      false,
		},
		{
			name:          "DateTime - invalid value should error",
			value:         "204r-07-04 05:50:15",
			originalValue: primitive.NewDateTimeFromTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			expected:      nil,
			hasError:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseValueByType(tc.value, tc.originalValue)

			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}
