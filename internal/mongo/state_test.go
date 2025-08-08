package mongo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCollectionState_UpdateFilter(t *testing.T) {
	cs := &CollectionState{Filter: `{"old": "filter"}`, Skip: 5}

	cs.SetFilter(`{"new": "filter"}`)
	assert.Equal(t, `{"new": "filter"}`, cs.Filter)
	assert.Equal(t, int64(0), cs.Skip)

	cs.SetFilter("  ")
	assert.Equal(t, "", cs.Filter)

	cs.SetFilter("{}")
	assert.Equal(t, "", cs.Filter)
}

func TestCollectionState_UpdateSort(t *testing.T) {
	cs := &CollectionState{Sort: `{"old": 1}`}

	cs.SetSort(`{"new": -1}`)
	assert.Equal(t, `{"new": -1}`, cs.Sort)

	cs.SetSort("  ")
	assert.Equal(t, "", cs.Sort)

	cs.SetSort("{}")
	assert.Equal(t, "", cs.Sort)
}

func TestCollectionState_UpdateProjection(t *testing.T) {
	cs := &CollectionState{Projection: `{"old": 1}`}

	cs.SetProjection(`{"new": 1, "field": 0}`)
	assert.Equal(t, `{"new": 1, "field": 0}`, cs.Projection)

	cs.SetProjection("  ")
	assert.Equal(t, "", cs.Projection)

	cs.SetProjection("{}")
	assert.Equal(t, "", cs.Projection)
}

func TestCollectionState_GetDocById(t *testing.T) {
	cs := &CollectionState{
		docs: []primitive.M{
			{"_id": "1", "value": 1},
		},
	}

	doc := cs.GetDocById("1")
	assert.NotNil(t, doc)
	assert.Equal(t, "1", doc["_id"])

	doc = cs.GetDocById("2")
	assert.Nil(t, doc)
}

func TestCollectionState_PopulateDocs(t *testing.T) {
	cs := &CollectionState{}
	docs := []primitive.M{
		{"_id": "1", "value": 1},
		{"_id": "2", "value": 2},
	}

	cs.PopulateDocs(docs)
	assert.Len(t, cs.docs, 2)
	assert.Equal(t, primitive.M{"_id": "1", "value": 1}, cs.docs[0])
	assert.Equal(t, primitive.M{"_id": "2", "value": 2}, cs.docs[1])
}

func TestCollectionState_AppendDoc(t *testing.T) {
	cs := &CollectionState{Count: 1}
	doc := primitive.M{"_id": "1", "value": 1}

	cs.AppendDoc(doc)
	assert.Len(t, cs.docs, 1)
	assert.Equal(t, doc, cs.docs[0])
	assert.Equal(t, int64(2), cs.Count)
}

func TestCollectionState_DeleteDoc(t *testing.T) {
	cs := &CollectionState{
		docs:  []primitive.M{{"_id": "1", "value": 1}},
		Count: 1,
	}

	cs.DeleteDoc("1")
	assert.Len(t, cs.docs, 0)
	assert.Equal(t, int64(0), cs.Count)
}

func TestCollectionState_GetJsonDocById_DoesNotModifyState(t *testing.T) {
	id1 := primitive.NewObjectID()
	id2 := primitive.NewObjectID()
	cs := &CollectionState{
		docs: []primitive.M{
			{"_id": id1, "value": 1},
			{"_id": id2, "value": 2},
		},
	}

	jsonDoc, err := cs.GetJsonDocById(id1)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonDoc)

	assert.Contains(t, jsonDoc, "$oid")
	assert.Contains(t, jsonDoc, id1.Hex())

	assert.Len(t, cs.docs, 2)
	assert.Equal(t, primitive.M{"_id": id1, "value": 1}, cs.docs[0])
	assert.Equal(t, primitive.M{"_id": id2, "value": 2}, cs.docs[1])
}

func TestCollectionState_GetDocById_WithBinaryId(t *testing.T) {
	binaryId := primitive.Binary{Data: []byte{1, 2, 3, 4}, Subtype: 0}
	cs := &CollectionState{
		docs: []primitive.M{
			{"_id": binaryId, "value": "binary_doc"},
		},
	}

	doc := cs.GetDocById(binaryId)
	assert.NotNil(t, doc)
	assert.Equal(t, binaryId, doc["_id"])
	assert.Equal(t, "binary_doc", doc["value"])

	differentBinaryId := primitive.Binary{Data: []byte{5, 6, 7, 8}, Subtype: 0}
	doc = cs.GetDocById(differentBinaryId)
	assert.Nil(t, doc)
}

func TestCollectionState_DeleteDoc_WithBinaryId(t *testing.T) {
	binaryId := primitive.Binary{Data: []byte{1, 2, 3, 4}, Subtype: 0}
	cs := &CollectionState{
		docs: []primitive.M{
			{"_id": binaryId, "value": "binary_doc"},
		},
		Count: 1,
	}

	cs.DeleteDoc(binaryId)
	assert.Len(t, cs.docs, 0)
	assert.Equal(t, int64(0), cs.Count)
}

func TestCollectionState_UpdateRawDoc_WithBinaryId(t *testing.T) {
	binaryId := primitive.Binary{Data: []byte{1, 2, 3, 4}, Subtype: 0}
	cs := &CollectionState{
		docs: []primitive.M{
			{"_id": binaryId, "value": "old_value"},
		},
	}

	updatedDoc := `{"_id": {"$binary": {"base64": "AQIDBA==", "subType": "00"}}, "value": "new_value"}`
	err := cs.UpdateRawDoc(updatedDoc)
	assert.NoError(t, err)
	assert.Len(t, cs.docs, 1)
	assert.Equal(t, "new_value", cs.docs[0]["value"])
}
