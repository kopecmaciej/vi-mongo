package util

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGetSortedKeysWithTypes(t *testing.T) {
	documents := []primitive.M{
		{"name": "John", "age": 30, "active": true},
		{"name": "Jane", "age": 25.5, "email": "jane@example.com"},
	}

	result := GetSortedKeysWithTypes(documents, tcell.ColorBlue.Name())

	expected := []string{
		"active [blue]Bool",
		"age [blue]Mixed",
		"email [blue]String",
		"name [blue]String",
	}

	assert.Equal(t, expected, result)
}

func TestGetValueByType(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"String", "test", "test"},
		{"Int", 42, "42"},
		{"Float", 3.14, "3.140000"},
		{"Bool", true, "true"},
		{"ObjectID", primitive.NewObjectID(), ""},                   // Hex value will be different each time
		{"DateTime", primitive.NewDateTimeFromTime(time.Now()), ""}, // Formatted time will be different
		{"Array", primitive.A{"a", "b"}, `["a","b"]`},
		{"Object", primitive.M{"key": "value"}, `{"key":"value"}`},
		{"Null", nil, "null"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetValueByType(tc.input)
			if tc.name == "ObjectID" {
				assert.Len(t, result, 24) // ObjectID hex string length
			} else if tc.name == "DateTime" {
				_, err := time.Parse(time.RFC3339, result)
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestGetMongoType(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"String", "test", TypeString},
		{"Int", 42, TypeInt},
		{"Float", 3.14, TypeDouble},
		{"Bool", true, TypeBool},
		{"ObjectID", primitive.NewObjectID(), TypeObjectId},
		{"DateTime", primitive.NewDateTimeFromTime(time.Now()), TypeDate},
		{"Array", primitive.A{"a", "b"}, TypeArray},
		{"Object", primitive.M{"key": "value"}, TypeObject},
		{"Null", nil, TypeNull},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetMongoType(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
