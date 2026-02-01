package util

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/kopecmaciej/tview"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	TypeString    = "String"
	TypeInt32     = "Int32"
	TypeInt64     = "Int64"
	TypeDouble    = "Double"
	TypeBool      = "Bool"
	TypeObjectId  = "ObjectID"
	TypeDate      = "Date"
	TypeTimestamp = "Timestamp"
	TypeArray     = "Array"
	TypeObject    = "Object"
	TypeRegex     = "Regex"
	TypeBinary    = "Binary"
	TypeMinKey    = "MinKey"
	TypeMaxKey    = "MaxKey"
	TypeMixed     = "Mixed"
	TypeNull      = "Null"
	TypeUndefined = "Undefined"
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

// StringifyMongoValueByType converts a value to a string
func StringifyMongoValueByType(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case int32, int64:
		return fmt.Sprintf("%d", t)
	case float32, float64:
		return fmt.Sprintf("%g", t)
	case bool:
		return fmt.Sprintf("%t", t)
	case primitive.ObjectID:
		return t.Hex()
	case primitive.DateTime:
		return t.Time().UTC().Format("2006-01-02T15:04:05.000+00:00")
	case primitive.A, primitive.D, primitive.M, map[string]any, []any:
		b, _ := json.Marshal(t)
		// Use tview's Escape function to prevent brackets from being interpreted as color tags
		return tview.Escape(string(b))
	case primitive.E:
		return fmt.Sprintf("%v", t)
	case primitive.Binary:
		return fmt.Sprintf("%v", t)
	case primitive.Regex:
		return fmt.Sprintf("%v", t)
	case primitive.Undefined:
		return fmt.Sprintf("%v", t)
	case primitive.MinKey:
		return fmt.Sprintf("%v", t)
	case primitive.MaxKey:
		return fmt.Sprintf("%v", t)
	default:
		return "null"
	}
}

// Helper function to determine MongoDB type
func GetMongoType(v any) string {
	switch v.(type) {
	case string:
		return TypeString
	case int32:
		return TypeInt32
	case int64:
		return TypeInt64
	case float32, float64:
		return TypeDouble
	case bool:
		return TypeBool
	case primitive.ObjectID:
		return TypeObjectId
	case primitive.DateTime:
		return TypeDate
	case primitive.Timestamp:
		return TypeTimestamp
	case primitive.Regex:
		return TypeRegex
	case primitive.Binary:
		return TypeBinary
	case primitive.MinKey:
		return TypeMinKey
	case primitive.MaxKey:
		return TypeMaxKey
	case primitive.Undefined:
		return TypeUndefined
	case primitive.A:
		return TypeArray
	case primitive.D, primitive.M:
		return TypeObject
	default:
		return TypeNull
	}
}

func DeepCopy(v primitive.M) primitive.M {
	copy := make(primitive.M)
	for k, v := range v {
		copy[k] = v
	}
	return copy
}
