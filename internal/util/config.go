package util

import (
	"reflect"
)

// MergeConfigs merges the loaded config with the default config
func MergeConfigs(loaded, defaultConfig interface{}) {
	mergeConfigsRecursive(reflect.ValueOf(loaded).Elem(), reflect.ValueOf(defaultConfig).Elem())
}

// mergeConfigsRecursive recursively merges nested structs
func mergeConfigsRecursive(loaded, defaultValue reflect.Value) {
	for i := 0; i < loaded.NumField(); i++ {
		field := loaded.Field(i)
		defaultField := defaultValue.Field(i)

		switch field.Kind() {
		case reflect.String:
			if field.String() == "" {
				field.Set(defaultField)
			}
		case reflect.Slice:
			if field.Len() == 0 {
				field.Set(defaultField)
			}
		case reflect.Struct:
			mergeConfigsRecursive(field, defaultField)
		}
	}
}
