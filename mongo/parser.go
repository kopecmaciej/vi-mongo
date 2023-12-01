package mongo

import (
	"fmt"
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
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
