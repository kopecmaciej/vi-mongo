package mongo

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestParseStringQuery(t *testing.T) {
	objectID, err := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	assert.NoError(t, err, "Failed to create ObjectID for testing")

	cases := []struct {
		name     string
		input    string
		expected map[string]interface{}
		hasError bool
	}{
		{
			name:     "Empty input",
			input:    "",
			expected: map[string]interface{}{},
			hasError: false,
		},
		{
			name:     "Valid input with ObjectID",
			input:    `{_id: ObjectID("507f1f77bcf86cd799439011")}`,
			expected: map[string]interface{}{"_id": objectID},
			hasError: false,
		},
		{
			name:     "Multiple fields with nested document",
			input:    `{ _id: ObjectID("507f1f77bcf86cd799439011"), user: { name: "John", age: 30 } }`,
			expected: map[string]interface{}{"_id": objectID, "user": primitive.M{"name": "John", "age": int32(30)}},
			hasError: false,
		},
		{
			name:     "Array and date",
			input:    `{ tags: ["mongodb", "database"], createdAt: { $date: "2023-04-15T12:00:00Z" } }`,
			expected: map[string]interface{}{"tags": primitive.A{"mongodb", "database"}, "createdAt": primitive.NewDateTimeFromTime(time.Date(2023, 4, 15, 12, 0, 0, 0, time.UTC))},
			hasError: false,
		},
		{
			name:     "Invalid ObjectID",
			input:    `{"_id": ObjectID("invalid")}`,
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

	input := map[string]interface{}{
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

func TestParseBsonValue(t *testing.T) {
	objectID, err := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	assert.NoError(t, err, "Failed to create ObjectID for testing")

	dateTime := primitive.NewDateTimeFromTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))

	cases := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "ObjectID",
			input:    primitive.NewObjectID(),
			expected: primitive.M{"$oid": objectID.Hex()},
		},
		{
			name:     "DateTime",
			input:    primitive.NewDateTimeFromTime(time.Now()),
			expected: primitive.M{"$date": dateTime.Time()},
		},
		{
			name:     "String",
			input:    "test",
			expected: "test",
		},
		{
			name:     "Int64",
			input:    int64(123),
			expected: int64(123),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := ParseBsonValue(tc.input)
			assert.IsType(t, tc.expected, result)
		})
	}
}
