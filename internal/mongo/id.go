package mongo

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetIDFromJSON returns the _id field of a JSON string as a primitive.ObjectID.
func GetIDFromJSON(jsonString string) (any, error) {
	var doc map[string]any
	err := json.Unmarshal([]byte(jsonString), &doc)
	if err != nil {
		log.Error().Err(err).Msg("Error unmarshaling JSON")
		return nil, err
	}

	objectID, err := getIdFromDocument(doc)
	if err != nil {
		log.Error().Err(err).Msg("Error converting _id to ObjectID")
		return nil, err
	}

	return objectID, nil
}

// getIdFromDocument returns the _id field of a document as a primitive.ObjectID
func getIdFromDocument(document map[string]any) (any, error) {
	rawId, ok := document["_id"]
	if !ok {
		return nil, fmt.Errorf("document has no _id")
	}
	var id any
	switch typedId := rawId.(type) {
	case primitive.ObjectID:
		return typedId, nil
	case string:
		id = typedId
	case map[string]any:
		oidString, ok := typedId["$oid"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid $oid field in _id")
		}
		objectId, err := primitive.ObjectIDFromHex(oidString)
		if err != nil {
			return nil, fmt.Errorf("invalid ObjectID: %w", err)
		}
		id = objectId
	default:
		return nil, fmt.Errorf("document _id is not a string or primitive.ObjectID")
	}

	return id, nil
}

// StringifyId converts the _id field of a document to a string
func StringifyId(id any) string {
	switch v := id.(type) {
	case primitive.ObjectID:
		return v.Hex()
	default:
		return fmt.Sprintf("%v", v)
	}
}
