package mongo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// parseQuery transforms a query string with ObjectId into a filter map compatible with MongoDB's BSON.
func ParseStringQuery(query string) (map[string]interface{}, error) {
	if query == "" {
		return map[string]interface{}{}, nil
	}

	query = strings.ReplaceAll(query, "ObjectId(\"", "{\"$oid\": \"")
	query = strings.ReplaceAll(query, "\")", "\"}")

	re := regexp.MustCompile(`(\w+):`)
	quotedKeysQuery := re.ReplaceAllString(query, `"$1":`)

	filter := map[string]interface{}{}

	err := bson.UnmarshalExtJSON([]byte(quotedKeysQuery), true, &filter)
	if err != nil {
		return nil, fmt.Errorf("error parsing query: %v", err)
	}

	return filter, nil
}

// GetIDFromJSON returns the _id field of a JSON string as a primitive.ObjectID.
func GetIDFromJSON(jsonString string) (primitive.ObjectID, error) {
	var doc map[string]interface{}
	err := json.Unmarshal([]byte(jsonString), &doc)
	if err != nil {
		log.Error().Err(err).Msg("Error unmarshaling JSON")
		return primitive.ObjectID{}, err
	}

	objectID, err := GetIDFromDocument(doc)
	if err != nil {
		log.Error().Err(err).Msg("Error converting _id to ObjectID")
		return primitive.ObjectID{}, err
	}

	return objectID, nil
}

// GetIDFromDocument returns the _id field of a document as a primitive.ObjectID
func GetIDFromDocument(document map[string]interface{}) (primitive.ObjectID, error) {
	rawId, ok := document["_id"]
	if !ok {
		return primitive.ObjectID{}, fmt.Errorf("document has no _id")
	}
	var id string
	switch rawId.(type) {
	case primitive.ObjectID:
		return rawId.(primitive.ObjectID), nil
	case string:
		id = rawId.(string)
	case map[string]interface{}:
		id = rawId.(map[string]interface{})["$oid"].(string)
	default:
		return primitive.ObjectID{}, fmt.Errorf("document _id is not a string or primitive.ObjectID")
	}

	return primitive.ObjectIDFromHex(id)
}

// ConvertIDsToOIDs converts a slice of documents to a slice of strings with the _id field converted to an $oid
func ConvertIDsToOIDs(documents []primitive.M) ([]string, error) {
	var docs []string
	for _, document := range documents {
		for key, value := range document {
			if oid, ok := value.(primitive.ObjectID); ok {
				obj := primitive.M{
					"$oid": oid.Hex(),
				}
				document[key] = obj
			}
		}
		jsonBytes, err := json.Marshal(document)
		if err != nil {
			log.Error().Err(err).Msg("Error marshaling JSON")
			continue
		}
		docs = append(docs, string(jsonBytes))
	}

	return docs, nil
}

func IndientJSON(jsonString string) (string, error) {
	var prettyJson bytes.Buffer
	err := json.Indent(&prettyJson, []byte(jsonString), "", "  ")
	if err != nil {
		log.Error().Err(err).Msg("Error marshaling JSON")
		return "", nil
	}
	text := string(prettyJson.Bytes())

  return text, nil
}
