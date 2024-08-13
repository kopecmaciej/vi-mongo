package mongo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// StringifyDocument converts a map to a JSON string
func StringifyDocument(document map[string]interface{}) (string, error) {
	// convert id to oid
	converted, err := ConvertSpecialTypes([]primitive.M{document})
	if err != nil {
		return "", err
	}
	return converted[0], nil
}

// ParseStringQuery transforms a query string with ObjectId into a filter map compatible with MongoDB's BSON.
// If keys are not quoted, this function will quote them.
func ParseStringQuery(query string) (map[string]interface{}, error) {
	if query == "" {
		return map[string]interface{}{}, nil
	}
	query = strings.ReplaceAll(query, " ", "")

	if strings.Contains(query, "$") {
		re := regexp.MustCompile(`(\{|\,)(\$[a-zA-Z0-9_.]+)`)
		query = re.ReplaceAllString(query, `$1"$2"`)
	}

	query = strings.ReplaceAll(query, "ObjectId(\"", "{\"$oid\": \"")
	query = strings.ReplaceAll(query, "\")", "\"}")

	re := regexp.MustCompile(`(\{|\,)\s*([a-zA-Z0-9_.]+)\s*:`)
	query = re.ReplaceAllString(query, `$1 "$2":`)

	filter := map[string]interface{}{}

	err := bson.UnmarshalExtJSON([]byte(query), true, &filter)
	if err != nil {
		return nil, fmt.Errorf("error parsing query %s: %w", query, err)
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

// ConvertSpecialTypes converts a slice of documents to a slice of strings with special types converted
func ConvertSpecialTypes(documents []primitive.M) ([]string, error) {
	var docs []string
	for _, document := range documents {
		 convertSpecialTypesInDoc(document)
		jsonBytes, err := json.Marshal(document)
		if err != nil {
			log.Error().Err(err).Msg("Error marshaling JSON")
			continue
		}
		docs = append(docs, string(jsonBytes))
	}

	return docs, nil
}

func convertSpecialTypesInDoc(doc primitive.M) {
	for key, value := range doc {
		switch v := value.(type) {
		case primitive.ObjectID:
			doc[key] = primitive.M{"$oid": v.Hex()}
		case primitive.DateTime:
			doc[key] = primitive.M{"$date": v.Time().Format(time.RFC3339)}
		case primitive.M:
			convertSpecialTypesInDoc(v)
		case []interface{}:
			for i, item := range v {
				if subDoc, ok := item.(primitive.M); ok {
					convertSpecialTypesInDoc(subDoc)
					v[i] = subDoc
				}
			}
		}
	}
}

// IndientJSON indents a JSON string and returns a a buffer
func IndientJSON(jsonString string) (bytes.Buffer, error) {
	var prettyJson bytes.Buffer
	err := json.Indent(&prettyJson, []byte(jsonString), "", "  ")
	if err != nil {
		log.Error().Err(err).Msg("Error marshaling JSON")
		return bytes.Buffer{}, err
	}

	return prettyJson, nil
}

// MongoToJSON converts a MongoDB document to a JSON string
func MongoToJSON(document bson.M) (string, error) {
	convertedDoc := convertMongoToJSON(document)
	jsonBytes, err := json.Marshal(convertedDoc)
	if err != nil {
		return "", fmt.Errorf("error marshaling to JSON: %w", err)
	}
	return string(jsonBytes), nil
}

func convertMongoToJSON(doc bson.M) bson.M {
	for key, value := range doc {
		switch v := value.(type) {
		case primitive.ObjectID:
			doc[key] = v.Hex()
		case primitive.DateTime:
			doc[key] = v.Time().Format(time.RFC3339)
		case bson.M:
			doc[key] = convertMongoToJSON(v)
		case []interface{}:
			for i, item := range v {
				if subDoc, ok := item.(bson.M); ok {
					v[i] = convertMongoToJSON(subDoc)
				}
			}
		}
	}
	return doc
}

// JSONToMongo converts a JSON string to a MongoDB document
func JSONToMongo(jsonString string) (bson.M, error) {
	var doc bson.M
	err := bson.UnmarshalExtJSON([]byte(jsonString), true, &doc)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON to MongoDB document: %w", err)
	}
	return convertJSONToMongo(doc), nil
}

func convertJSONToMongo(doc bson.M) bson.M {
	for key, value := range doc {
		switch v := value.(type) {
		case string:
			if key == "_id" || strings.HasSuffix(key, "Id") {
				if objectID, err := primitive.ObjectIDFromHex(v); err == nil {
					doc[key] = objectID
				}
			} else if t, err := time.Parse(time.RFC3339, v); err == nil {
				doc[key] = primitive.NewDateTimeFromTime(t)
			}
		case bson.M:
			doc[key] = convertJSONToMongo(v)
		case []interface{}:
			for i, item := range v {
				if subDoc, ok := item.(bson.M); ok {
					v[i] = convertJSONToMongo(subDoc)
				}
			}
		}
	}
	return doc
}