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
	converted, err := ParseRawDocuments([]primitive.M{document})
	if err != nil {
		return "", err
	}
	return converted[0], nil
}

// ParseStringQuery transforms a query string with ObjectId into a filter map compatible with MongoDB's BSON.
// If keys are not quoted, this function will quote them.
func ParseStringQuery(query string) (map[string]interface{}, error) {
	var parseError error
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

	dateRegex := regexp.MustCompile(`\{\"\$date\"\s*:\s*\"(.*?)\"\}`)
	query = dateRegex.ReplaceAllStringFunc(query, func(match string) string {
		dateStr := dateRegex.FindStringSubmatch(match)[1]
		t, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			parseError = err
			return match
		}
		return fmt.Sprintf(`{"$date":{"$numberLong":"%d"}}`, primitive.NewDateTimeFromTime(t).Time().UnixMilli())
	})
	if parseError != nil {
		return nil, fmt.Errorf("error parsing date: %w", parseError)
	}

	re := regexp.MustCompile(`(\{|\,)\s*([a-zA-Z0-9_.]+)\s*:`)
	query = re.ReplaceAllString(query, `$1 "$2":`)

	var filter primitive.M
	parseError = bson.UnmarshalExtJSON([]byte(query), true, &filter)
	if parseError != nil {
		return nil, fmt.Errorf("error parsing query %s: %w", query, parseError)
	}

	return filter, nil
}

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

// ParseRawDocuments converts a slice of documents to a slice of strings with
// mongo compatible JSON
func ParseRawDocuments(documents []primitive.M) ([]string, error) {
	var docs []string
	for _, doc := range documents {
		for key, value := range doc {
			switch v := value.(type) {
			case primitive.ObjectID:
				doc[key] = primitive.M{
					"$oid": v.Hex(),
				}
			case primitive.DateTime:
				doc[key] = primitive.M{
					"$date": v.Time(),
				}
			}
		}
		jsonBytes, err := json.Marshal(doc)
		if err != nil {
			log.Error().Err(err).Msg("Error marshaling JSON")
			continue
		}
		docs = append(docs, string(jsonBytes))
	}

	return docs, nil
}

// IndentJSON indents a JSON string and returns a a buffer
func IndentJSON(jsonString string) (bytes.Buffer, error) {
	var prettyJson bytes.Buffer
	err := json.Indent(&prettyJson, []byte(jsonString), "", "  ")
	if err != nil {
		log.Error().Err(err).Msg("Error marshaling JSON")
		return bytes.Buffer{}, err
	}

	return prettyJson, nil
}

// ParseJSONDocument converts a JSON string to a primitive.M document
func ParseJSONDocument(jsonDoc string) (primitive.M, error) {
	parsedDocs, err := ParseJSONDocuments([]string{jsonDoc})
	if err != nil {
		return primitive.M{}, err
	}
	return parsedDocs[0], nil
}

// ParseJSONDocuments converts a slice of JSON strings to a slice of primitive.M documents
// with MongoDB-compatible types (ObjectId for $oid and DateTime for $date)
func ParseJSONDocuments(jsonDocs []string) ([]primitive.M, error) {
	var documents []primitive.M

	for _, jsonDoc := range jsonDocs {
		var doc map[string]interface{}
		err := json.Unmarshal([]byte(jsonDoc), &doc)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
		}

		convertedDoc := make(primitive.M)
		for key, value := range doc {
			convertedValue, err := convertValue(value)
			if err != nil {
				return nil, fmt.Errorf("error converting value for key %s: %w", key, err)
			}
			convertedDoc[key] = convertedValue
		}

		documents = append(documents, convertedDoc)
	}

	return documents, nil
}

// convertValue converts a value to a compatible MongoDB type
func convertValue(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case map[string]interface{}:
		if oid, ok := v["$oid"]; ok {
			return primitive.ObjectIDFromHex(oid.(string))
		}
		if date, ok := v["$date"]; ok {
			t, err := time.Parse(time.RFC3339, date.(string))
			if err != nil {
				return nil, fmt.Errorf("error parsing date: %w", err)
			}
			return primitive.NewDateTimeFromTime(t), nil
		}
		convertedMap := make(map[string]interface{})
		for k, v := range v {
			convertedValue, err := convertValue(v)
			if err != nil {
				return nil, err
			}
			convertedMap[k] = convertedValue
		}
		return convertedMap, nil
	case []interface{}:
		convertedArray := make([]interface{}, len(v))
		for i, elem := range v {
			convertedElem, err := convertValue(elem)
			if err != nil {
				return nil, err
			}
			convertedArray[i] = convertedElem
		}
		return convertedArray, nil
	default:
		return v, nil
	}
}
