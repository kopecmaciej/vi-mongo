package mongo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/kopecmaciej/vi-mongo/internal/util"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ParseBsonDocument converts a map to a JSON string
func ParseBsonDocument(document map[string]interface{}) (string, error) {
	// convert id to oid
	converted, err := ParseBsonDocuments([]primitive.M{document})
	if err != nil {
		return "", err
	}
	return converted[0], nil
}

// ParseBsonDocuments converts a slice of documents to a slice of strings with
// mongo compatible JSON
func ParseBsonDocuments(documents []primitive.M) ([]string, error) {
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

// ParseStringQuery transforms a query string with ObjectId into a filter map compatible with MongoDB's BSON.
// If keys are not quoted, this function will quote them.
func ParseStringQuery(query string) (map[string]interface{}, error) {
	var parseError error
	if query == "" {
		return map[string]interface{}{}, nil
	}

	if strings.Contains(query, "$") {
		query = util.QuoteUnquotedKeys(query)
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

	query = util.QuoteUnquotedKeys(query)

	var filter primitive.M
	parseError = bson.UnmarshalExtJSON([]byte(query), true, &filter)
	if parseError != nil {
		return nil, fmt.Errorf("error parsing query %s: %w", query, parseError)
	}

	return filter, nil
}

// IndentJson indents a JSON string and returns a a buffer
func IndentJson(jsonString string) (bytes.Buffer, error) {
	var prettyJson bytes.Buffer
	err := json.Indent(&prettyJson, []byte(jsonString), "", "  ")
	if err != nil {
		log.Error().Err(err).Msg("Error marshaling JSON")
		return bytes.Buffer{}, err
	}

	return prettyJson, nil
}

// ParseJsonToBson converts a JSON string to a primitive.M document
func ParseJsonToBson(jsonDoc string) (primitive.M, error) {
	var doc map[string]interface{}
	err := json.Unmarshal([]byte(jsonDoc), &doc)
	if err != nil {
		return primitive.M{}, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	return convertToBson(doc)
}

// convertToBson converts a map[string]interface{} to a primitive.M document
// with MongoDB-compatible types (ObjectId for $oid and DateTime for $date)
func convertToBson(doc map[string]interface{}) (primitive.M, error) {
	convertedDoc := make(primitive.M)
	for key, value := range doc {
		convertedValue, err := convertValue(value)
		if err != nil {
			return nil, fmt.Errorf("error converting value for key %s: %w", key, err)
		}
		convertedDoc[key] = convertedValue
	}
	return convertedDoc, nil
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
