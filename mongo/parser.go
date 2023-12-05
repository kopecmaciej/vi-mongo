package mongo

import (
	"fmt"
	"regexp"
	"strings"

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

func GetIDFromDocument(document map[string]interface{}) (primitive.ObjectID, error) {
	oid, ok := document["_id"].(map[string]interface{})
	if !ok {
		return primitive.ObjectID{}, fmt.Errorf("document has no _id")
	}
	id, ok := oid["$oid"].(string)
	if !ok {
		return primitive.ObjectID{}, fmt.Errorf("document has no $oid")
	}
	return primitive.ObjectIDFromHex(id)
}
