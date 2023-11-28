package mongo

import (
	"fmt"
	"testing"

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
  query := fmt.Sprintf(`{_id: ObjectId("%s")}`, objectID.Hex())
	expected := map[string]interface{}{"_id": objectID}

	result, err := ParseStringQuery(query)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestParseQueryWithMultipleFields(t *testing.T) {
  objectID := primitive.NewObjectID()
	query := fmt.Sprintf(`{_id: ObjectId("%s"), name: "John"}`, objectID.Hex())
	expected := map[string]interface{}{"_id": objectID, "name": "John"}

	result, err := ParseStringQuery(query)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestParseQueryWithMultipleFieldsAndSpaces(t *testing.T) {
  objectID := primitive.NewObjectID()
  query := fmt.Sprintf(`{ _id: ObjectId("%s"), name: "John" }`, objectID.Hex())
  expected := map[string]interface{}{"_id": objectID, "name": "John"}

  result, err := ParseStringQuery(query)

  assert.NoError(t, err)
  assert.Equal(t, expected, result)
}

func TestParseQueryInvalidInput(t *testing.T) {
  query := `{"_id": ObjectId("123")}`

  _, err := ParseStringQuery(query)

  expected := fmt.Errorf("error parsing query")
  assert.Error(t, err)
  assert.Contains(t, err.Error(), expected.Error())
}
