package mongo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGetIDFromJSON(t *testing.T) {
	t.Run("Valid JSON with ObjectID", func(t *testing.T) {
		jsonString := `{"_id": {"$oid": "5f8f9e5f1c9d440000d1b3c5"}}`
		id, err := GetIDFromJSON(jsonString)
		assert.NoError(t, err)
		assert.IsType(t, primitive.ObjectID{}, id)
		assert.Equal(t, "5f8f9e5f1c9d440000d1b3c5", id.(primitive.ObjectID).Hex())
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		jsonString := `{"_id": "invalid`
		_, err := GetIDFromJSON(jsonString)
		assert.Error(t, err)
	})
}

func TestGetIDFromDocument(t *testing.T) {
	t.Run("Document with ObjectID", func(t *testing.T) {
		objectID := primitive.NewObjectID()
		doc := map[string]interface{}{"_id": objectID}
		id, err := getIdFromDocument(doc)
		assert.NoError(t, err)
		assert.Equal(t, objectID, id)
	})

	t.Run("Document with string ID", func(t *testing.T) {
		doc := map[string]interface{}{"_id": "123456"}
		id, err := getIdFromDocument(doc)
		assert.NoError(t, err)
		assert.Equal(t, "123456", id)
	})

	t.Run("Document with $oid", func(t *testing.T) {
		doc := map[string]interface{}{"_id": map[string]interface{}{"$oid": "5f8f9e5f1c9d440000d1b3c5"}}
		id, err := getIdFromDocument(doc)
		assert.NoError(t, err)
		assert.IsType(t, primitive.ObjectID{}, id)
		assert.Equal(t, "5f8f9e5f1c9d440000d1b3c5", id.(primitive.ObjectID).Hex())
	})

	t.Run("Document without _id", func(t *testing.T) {
		doc := map[string]interface{}{"name": "John"}
		_, err := getIdFromDocument(doc)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "document has no _id")
	})
}

func TestStringifyId(t *testing.T) {
	t.Run("ObjectID", func(t *testing.T) {
		objectID := primitive.NewObjectID()
		result := StringifyId(objectID)
		assert.Equal(t, objectID.Hex(), result)
	})

	t.Run("String", func(t *testing.T) {
		result := StringifyId("123456")
		assert.Equal(t, "123456", result)
	})

	t.Run("Int", func(t *testing.T) {
		result := StringifyId(123456)
		assert.Equal(t, "123456", result)
	})
}
