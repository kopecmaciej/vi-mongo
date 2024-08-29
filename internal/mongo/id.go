package mongo

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetIDFromJSON returns the _id field of a JSON string as a primitive.ObjectID.
func GetIDFromJSON(jsonString string) (interface{}, error) {
	var doc map[string]interface{}
	err := json.Unmarshal([]byte(jsonString), &doc)
	if err != nil {
		log.Error().Err(err).Msg("Error unmarshaling JSON")
		return nil, err
	}

	objectID, err := GetIDFromDocument(doc)
	if err != nil {
		log.Error().Err(err).Msg("Error converting _id to ObjectID")
		return nil, err
	}

	return objectID, nil
}

// GetIDFromDocument returns the _id field of a document as a primitive.ObjectID
func GetIDFromDocument(document map[string]interface{}) (interface{}, error) {
	rawId, ok := document["_id"]
	if !ok {
		return nil, fmt.Errorf("document has no _id")
	}
	var id interface{}
	switch typedId := rawId.(type) {
	case primitive.ObjectID:
		return typedId, nil
	case string:
		id = typedId
	case map[string]interface{}:
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
func StringifyId(id interface{}) string {
	switch v := id.(type) {
	case primitive.ObjectID:
		return v.Hex()
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}
