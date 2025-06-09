package mongo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/kopecmaciej/vi-mongo/internal/util"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ParseBsonDocument converts a map to a JSON string
func ParseBsonDocument(document map[string]any) (string, error) {
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
		sortedDoc := sortDocumentKeys(doc)
		jsonBytes, err := bson.MarshalExtJSON(sortedDoc, false, false)
		if err != nil {
			log.Error().Err(err).Msg("Error marshaling JSON")
			continue
		}
		docs = append(docs, string(jsonBytes))
	}

	return docs, nil
}

// TODO: Remove this and convert everything to primitive.D
func sortDocumentKeys(doc primitive.M) primitive.D {
	keys := make([]string, 0, len(doc))
	for key := range doc {
		keys = append(keys, key)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return strings.Compare(keys[i], keys[j]) < 0
	})

	sortedDoc := primitive.D{}
	for _, key := range keys {
		value := doc[key]
		sortedValue := sortValue(value)
		sortedDoc = append(sortedDoc, bson.E{Key: key, Value: sortedValue})
	}

	return sortedDoc
}

func sortValue(value any) any {
	switch v := value.(type) {
	case primitive.M:
		return sortDocumentKeys(v)
	case []any:
		return sortArray(v)
	case primitive.A:
		return sortArray([]any(v))
	default:
		return value
	}
}

func sortArray(arr []any) primitive.A {
	sorted := make(primitive.A, len(arr))
	for i, v := range arr {
		sorted[i] = sortValue(v)
	}
	return sorted
}

// ParseStringQuery transforms a query string with ObjectID into a filter map compatible with MongoDB's BSON.
// If keys are not quoted, this function will quote them.
func ParseStringQuery(query string) (map[string]any, error) {
	if query == "" {
		return map[string]any{}, nil
	}

	query = util.QuoteUnquotedKeys(query)

	query = strings.ReplaceAll(query, "ObjectID(\"", "{\"$oid\": \"")
	query = strings.ReplaceAll(query, "\")", "\"}")

	var filter primitive.M
	err := bson.UnmarshalExtJSON([]byte(query), false, &filter)
	if err != nil {
		log.Error().Err(err).Msgf("Error parsing query %s", query)
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
	var doc primitive.M
	err := bson.UnmarshalExtJSON([]byte(jsonDoc), false, &doc)
	if err != nil {
		log.Error().Err(err).Msg("Error unmarshaling JSON")
		return primitive.M{}, fmt.Errorf("error unmarshaling JSON: %w", err)
	}
	return doc, nil
}

func ParseValueByType(value string, originalValue any) (any, error) {
	if originalValue != nil {
		switch originalValue.(type) {
		case primitive.M, map[string]interface{}:
			if strings.HasPrefix(strings.TrimSpace(value), "{") && strings.HasSuffix(strings.TrimSpace(value), "}") {
				if parsed, err := ParseJsonToBson(value); err == nil {
					return parsed, nil
				}
			}
		case primitive.A, []any:
			if strings.HasPrefix(strings.TrimSpace(value), "[") && strings.HasSuffix(strings.TrimSpace(value), "]") {
				return ParseJsonArray(value)
			}
		case int, int32, int64:
			return stringToInt(value)
		case float32, float64:
			return stringToFloat(value)
		case bool:
			return stringToBool(value)
		}
	}

	if strings.HasPrefix(strings.TrimSpace(value), "{") && strings.HasSuffix(strings.TrimSpace(value), "}") {
		if parsed, err := ParseJsonToBson(value); err == nil {
			return parsed, nil
		}
	}

	if strings.HasPrefix(strings.TrimSpace(value), "[") && strings.HasSuffix(strings.TrimSpace(value), "]") {
		return ParseJsonArray(value)
	}

	if value == "true" || value == "false" {
		return stringToBool(value)
	}

	if intVal, err := stringToInt(value); err == nil {
		return intVal, nil
	}

	if floatVal, err := stringToFloat(value); err == nil {
		return floatVal, nil
	}

	return value, nil
}

func stringToInt(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func stringToFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func stringToBool(s string) (bool, error) {
	return strconv.ParseBool(s)
}

func ParseJsonArray(value string) (any, error) {
	var jsonArray []any
	if err := json.Unmarshal([]byte(value), &jsonArray); err != nil {
		return value, nil
	}

	bsonArray := make(primitive.A, len(jsonArray))
	copy(bsonArray, jsonArray)

	return bsonArray, nil
}
