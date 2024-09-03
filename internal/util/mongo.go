package util

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	TypeString   = "String"
	TypeInt      = "Int"
	TypeDouble   = "Double"
	TypeBool     = "Bool"
	TypeObjectId = "ObjectID"
	TypeDate     = "Date"
	TypeArray    = "Array"
	TypeObject   = "Object"
	TypeMixed    = "Mixed"
	TypeNull     = "Null"
)

func GetSortedKeysWithTypes(documents []primitive.M, typeColor string) []string {
	keys := make(map[string]string)
	for _, doc := range documents {
		for k, v := range doc {
			if _, exists := keys[k]; exists && keys[k] != GetMongoType(v) {
				keys[k] = TypeMixed
			} else {
				keys[k] = GetMongoType(v)
			}
		}
	}

	// Sort the keys for consistent column order
	sortedKeys := make([]string, 0, len(keys))
	for k, t := range keys {
		sortedKeys = append(sortedKeys, fmt.Sprintf("%s [%s]%s", k, typeColor, t))
	}
	sort.Strings(sortedKeys)

	return sortedKeys
}

func GetValueByType(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case int, int32, int64:
		return fmt.Sprintf("%d", t)
	case float32, float64:
		return fmt.Sprintf("%f", t)
	case bool:
		return fmt.Sprintf("%t", t)
	case primitive.ObjectID:
		return t.Hex()
	case primitive.DateTime:
		return t.Time().Format(time.RFC3339)
	case primitive.A:
		b, _ := json.Marshal(t)
		return string(b)
	case primitive.D, primitive.M:
		b, _ := json.Marshal(t)
		return string(b)
	default:
		return "null"
	}
}

// Helper function to determine MongoDB type
func GetMongoType(v interface{}) string {
	switch v.(type) {
	case string:
		return TypeString
	case int, int32, int64:
		return TypeInt
	case float32, float64:
		return TypeDouble
	case bool:
		return TypeBool
	case primitive.ObjectID:
		return TypeObjectId
	case primitive.DateTime:
		return TypeDate
	case primitive.A:
		return TypeArray
	case primitive.D, primitive.M:
		return TypeObject
	default:
		return TypeNull
	}
}
