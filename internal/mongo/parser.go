package mongo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kopecmaciej/vi-mongo/internal/util"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

// ParseStringQuery transforms a query string into a filter map compatible with MongoDB's BSON.
// It transforms special Mongodb JS syntax into proper BSON
func ParseStringQuery(query string) (map[string]any, error) {
	if query == "" {
		return map[string]any{}, nil
	}

	query = util.QuoteUnquotedKeys(query)
	var err error
	query, err = util.TransformMongoshSyntax(query)
	if err != nil {
		return nil, err
	}

	query = strings.ReplaceAll(query, "'", "\"")

	var filter primitive.M
	err = bson.UnmarshalExtJSON([]byte(query), false, &filter)
	if err != nil {
		log.Error().Err(err).Msgf("Error parsing query %s", query)
		return nil, fmt.Errorf("error parsing query %s: %w", query, err)
	}

	filter = util.ConvertRegexInArrays(filter)

	return filter, nil
}

// ParseSortOptions parses a sort options string into a BSON-compatible map.
func ParseSortOptions(sortOptions string) (map[string]any, error) {
	if sortOptions == "" {
		return map[string]any{}, nil
	}

	sortOptions = util.QuoteUnquotedKeys(sortOptions)

	var sort primitive.M
	err := bson.UnmarshalExtJSON([]byte(sortOptions), false, &sort)
	if err != nil {
		log.Error().Err(err).Msgf("Error parsing sort options %s", sortOptions)
		return nil, fmt.Errorf("error parsing sort options %s: %w", sortOptions, err)
	}

	return sort, nil
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
		case primitive.DateTime:
			for _, format := range util.MongoDateFormats {
				if parsed, err := time.Parse(format, value); err == nil {
					return primitive.NewDateTimeFromTime(parsed), nil
				}
			}
			return nil, fmt.Errorf("unable to parse datetime value: %s", value)
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

// ParsePipeline parses a slice of stage JSON strings into a mongo.Pipeline.
// Each stage should be a full stage document like {$match: {status: "active"}}.
// Key order within each stage is preserved by unmarshalling directly into bson.D.
func ParsePipeline(stages []string) (mongo.Pipeline, error) {
	pipeline := make(mongo.Pipeline, 0, len(stages))
	for idx, stage := range stages {
		normalized := util.QuoteUnquotedKeys(stage)
		var err error
		normalized, err = util.TransformMongoshSyntax(normalized)
		if err != nil {
			return nil, fmt.Errorf("stage %d: %w", idx, err)
		}
		normalized = strings.ReplaceAll(normalized, "'", "\"")

		var bsonStage bson.D
		if err := bson.UnmarshalExtJSON([]byte(normalized), false, &bsonStage); err != nil {
			return nil, fmt.Errorf("stage %d: %w", idx, err)
		}
		pipeline = append(pipeline, bsonStage)
	}
	return pipeline, nil
}

// ExtractStageOperator returns the top-level key from a stage JSON string (e.g. "$match").
func ExtractStageOperator(stage string) string {
	parsed, err := ParseStringQuery(stage)
	if err != nil || len(parsed) == 0 {
		return stage
	}
	for k := range parsed {
		return k
	}
	return stage
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

// ReconcileDocumentTypes walks updatedDoc and, for each field that also exists
// in origDoc, coerces the updated value back to the original BSON type when it
// can be represented losslessly. This prevents silent type changes (e.g. double
// → int32) when editing documents in an external editor.
// To explicitly change a type, use Extended JSON notation (e.g. {"$numberInt":"8"}).
func ReconcileDocumentTypes(origDoc, updatedDoc primitive.M) primitive.M {
	result := make(primitive.M, len(updatedDoc))
	for key, updatedVal := range updatedDoc {
		if origVal, exists := origDoc[key]; exists {
			result[key] = reconcileFieldType(origVal, updatedVal)
		} else {
			result[key] = updatedVal
		}
	}
	return result
}

func reconcileFieldType(origVal, updatedVal any) any {
	if origVal == nil || updatedVal == nil {
		return updatedVal
	}

	if origMap, ok := origVal.(primitive.M); ok {
		if updatedMap, ok := updatedVal.(primitive.M); ok {
			return ReconcileDocumentTypes(origMap, updatedMap)
		}
		return updatedVal
	}

	if origArr, ok := toAnySlice(origVal); ok {
		if updatedArr, ok := toAnySlice(updatedVal); ok {
			result := make(primitive.A, len(updatedArr))
			for i, updatedElem := range updatedArr {
				if i < len(origArr) {
					result[i] = reconcileFieldType(origArr[i], updatedElem)
				} else {
					result[i] = updatedElem
				}
			}
			return result
		}
		return updatedVal
	}

	return coerceNumericType(origVal, updatedVal)
}

// coerceNumericType tries to represent updatedVal in the same numeric type as
// origVal, but only when it can be done losslessly (no fractional part lost).
func coerceNumericType(origVal, updatedVal any) any {
	updatedF, ok := anyToFloat64(updatedVal)
	if !ok {
		return updatedVal
	}
	switch origVal.(type) {
	case int32:
		if updatedF == math.Trunc(updatedF) {
			return int32(updatedF)
		}
	case int64:
		if updatedF == math.Trunc(updatedF) {
			return int64(updatedF)
		}
	case float32:
		return float32(updatedF)
	case float64:
		return updatedF
	}
	return updatedVal
}

func anyToFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case float32:
		return float64(val), true
	case float64:
		return val, true
	}
	return 0, false
}

func toAnySlice(v any) ([]any, bool) {
	switch val := v.(type) {
	case primitive.A:
		return []any(val), true
	case []any:
		return val, true
	}
	return nil, false
}
