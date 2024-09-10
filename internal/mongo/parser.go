package mongo

import (
	"bytes"
	"encoding/json"
	"fmt"
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
			doc[key] = ParseBsonValue(value)
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

func ParseBsonValue(value interface{}) interface{} {
	var parsed interface{}
	switch v := value.(type) {
	case primitive.ObjectID:
		parsed = primitive.M{
			"$oid": v.Hex(),
		}
	case primitive.DateTime:
		parsed = primitive.M{
			"$date": v.Time(),
		}
	}

	if parsed == nil {
		return value
	}

	return parsed
}

// ParseStringQuery transforms a query string with ObjectID into a filter map compatible with MongoDB's BSON.
// If keys are not quoted, this function will quote them.
func ParseStringQuery(query string) (map[string]interface{}, error) {
	if query == "" {
		return map[string]interface{}{}, nil
	}

	query = util.QuoteUnquotedKeys(query)

	query = strings.ReplaceAll(query, "ObjectID(\"", "{\"$oid\": \"")
	query = strings.ReplaceAll(query, "\")", "\"}")

	query, err := util.ParseDate(query)
	if err != nil {
		return nil, fmt.Errorf("error parsing date: %w", err)
	}

	var filter primitive.M
	err = bson.UnmarshalExtJSON([]byte(query), true, &filter)
	if err != nil {
		return nil, fmt.Errorf("error parsing query %s: %w", query, err)
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
// with MongoDB-compatible types (ObjectID for $oid and DateTime for $date)
func convertToBson(doc map[string]interface{}) (primitive.M, error) {
	convertedDoc := make(primitive.M)
	for key, value := range doc {
		convertedValue, err := ParseJsonValue(value)
		if err != nil {
			return nil, fmt.Errorf("error converting value for key %s: %w", key, err)
		}
		convertedDoc[key] = convertedValue
	}
	return convertedDoc, nil
}

// ParseJsonValue converts a value to a compatible MongoDB type
func ParseJsonValue(value interface{}) (interface{}, error) {
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
			convertedValue, err := ParseJsonValue(v)
			if err != nil {
				return nil, err
			}
			convertedMap[k] = convertedValue
		}
		return convertedMap, nil
	case []interface{}:
		convertedArray := make([]interface{}, len(v))
		for i, elem := range v {
			convertedElem, err := ParseJsonValue(elem)
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
