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
	objId := primitive.NewObjectID()
	dateUtc := "2023-10-05T14:34:24Z"
	fixedDate := time.Date(2023, 10, 5, 14, 34, 24, 0, time.UTC)
	date := primitive.NewDateTimeFromTime(fixedDate)

	testCases := []struct {
		name     string
		input    any
		expected string
	}{
		{"String", "test", "test"},
		{"Int32", int32(56), "56"},
		{"Int64", int64(922337203685477), "922337203685477"},
		{"Float", 3.14, "3.14"},
		{"Bool", true, "true"},
		{"ObjectID", objId, objId.Hex()},
		{"DateTime", date, dateUtc},
		{"Array", primitive.A{"a", "b"}, `["a","b"]`},
		{"Object", primitive.M{"key": "value"}, `{"key":"value"}`},
		{"Null", nil, "null"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := StringifyMongoValueByType(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetMongoType(t *testing.T) {
	testCases := []struct {
		name     string
		input    any
		expected string
	}{
		{"String", "test", TypeString},
		{"Int32", int32(56), TypeInt32},
		{"Int64", int64(922337203685477), TypeInt64},
		{"Float32", float32(3.14), TypeDouble},
		{"Float64", float64(3.14), TypeDouble},
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
