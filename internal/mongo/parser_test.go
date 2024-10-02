package mongo

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestParseQueryEmptyInput(t *testing.T) {
	result, err := ParseStringQuery("")

	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{}, result)
}

func TestParseQueryValidInput(t *testing.T) {
	objectID := primitive.NewObjectID()
	query := fmt.Sprintf(`{_id: ObjectID("%s")}`, objectID.Hex())
	expected := map[string]interface{}{"_id": objectID}

	result, err := ParseStringQuery(query)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestParseQueryWithMultipleFields(t *testing.T) {
	objectID := primitive.NewObjectID()
	query := fmt.Sprintf(`{_id: ObjectID("%s"), name: "John"}`, objectID.Hex())
	expected := map[string]interface{}{"_id": objectID, "name": "John"}

	result, err := ParseStringQuery(query)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestParseQueryWithMultipleFieldsAndSpaces(t *testing.T) {
	objectID := primitive.NewObjectID()
	query := fmt.Sprintf(`{ _id: ObjectID("%s"), name: "John" }`, objectID.Hex())
	expected := map[string]interface{}{"_id": objectID, "name": "John"}

	result, err := ParseStringQuery(query)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestParseQueryWithMultipleNestedFieldsAndSpaces(t *testing.T) {
	objectID := primitive.NewObjectID()
	query := fmt.Sprintf(`{ _id: ObjectID("%s"), name.first: "John" }`, objectID.Hex())
	expected := map[string]interface{}{"_id": objectID, "name.first": "John"}

	result, err := ParseStringQuery(query)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestParseQueryWithOperator(t *testing.T) {
	objectID := primitive.NewObjectID()
	query := fmt.Sprintf(`{ _id: ObjectID("%s"), name: { $exists: true } }`, objectID.Hex())
	expected := map[string]interface{}{"_id": objectID, "name": primitive.M{"$exists": true}}

	result, err := ParseStringQuery(query)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestParseQueryInvalidInput(t *testing.T) {
	query := `{"_id": ObjectID("123")}`

	_, err := ParseStringQuery(query)

	expected := fmt.Errorf("error parsing query")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expected.Error())
}

func TestParseQueryWithNestedDocument(t *testing.T) {
	query := `{ user: { name: "Mike", weight: 75.7 } }`
	expected := map[string]interface{}{"user": primitive.M{"name": "Mike", "weight": 75.7}}

	result, err := ParseStringQuery(query)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestParseQueryWithArray(t *testing.T) {
	query := `{ tags: ["mongodb", "database"] }`
	expected := map[string]interface{}{"tags": primitive.A{"mongodb", "database"}}

	result, err := ParseStringQuery(query)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestParseQueryWithDate(t *testing.T) {
	query := `{ createdAt: { $date: "2023-04-15T12:00:00Z" } }`
	expectedDate := time.Date(2023, 4, 15, 12, 0, 0, 0, time.UTC)
	expected := map[string]interface{}{"createdAt": primitive.NewDateTimeFromTime(expectedDate)}

	result, err := ParseStringQuery(query)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestParseQueryWithMultipleOperators(t *testing.T) {
	query := `{ age: { $gte: 18, $lte: 65 } }`
	expected := map[string]interface{}{"age": primitive.M{"$gte": int32(18), "$lte": int32(65)}}

	result, err := ParseStringQuery(query)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestParseQueryWithRegex(t *testing.T) {
	query := `{ name: { $regex: "^J", $options: "i" } }`
	expected := map[string]interface{}{"name": primitive.M{"$regex": "^J", "$options": "i"}}

	result, err := ParseStringQuery(query)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}
