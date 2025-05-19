package util

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/adrg/xdg"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

const (
	ConfigDir = "vi-mongo"
)

// MergeConfigs merges the loaded config with the default config
func MergeConfigs(loaded, defaultConfig any) {
	loadedVal := reflect.ValueOf(loaded).Elem()
	defaultVal := reflect.ValueOf(defaultConfig).Elem()
	mergeConfigsRecursive(loadedVal, defaultVal)
}

// mergeConfigsRecursive recursively merges nested structs.
// This may be a bit complicated for such a simple merge, but it allows for
// more flexibility in the future if we want to add more complex merging logic
// TODO: probably merging keybindings and config should be split into two functions
func mergeConfigsRecursive(loaded, defaultValue reflect.Value) {
	for i := 0; i < loaded.NumField(); i++ {
		field := loaded.Field(i)
		defaultField := defaultValue.Field(i)

		// Special handling for Key structs
		if field.Type().Name() == "Key" {
			// If any field in the Key struct is set, keep the entire struct as-is
			if !isEmptyKey(field) {
				continue
			}
			// If the Key struct is completely empty, use the default
			field.Set(defaultField)
			continue
		}

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

// isEmptyKey checks if a Key struct is completely empty
func isEmptyKey(keyValue reflect.Value) bool {
	for i := 0; i < keyValue.NumField(); i++ {
		field := keyValue.Field(i)
		switch field.Kind() {
		case reflect.String:
			if field.String() != "" {
				return false
			}
		case reflect.Slice:
			if field.Len() > 0 {
				return false
			}
		}
	}
	return true
}

// LoadConfigFile loads a configuration file, merges it with defaults, and returns the result
func LoadConfigFile[T any](defaultConfig *T, configPath string) (*T, error) {
	err := ensureConfigDirExist()
	if err != nil {
		log.Error().Err(err).Msg("Failed to ensure config directory exists")
		return nil, fmt.Errorf("failed to ensure config directory exists: %w", err)
	}

	bytes, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			bytes, err = marshalConfig(defaultConfig, configPath)
			if err != nil {
				log.Error().Err(err).Str("path", configPath).Msg("Failed to marshal default config")
				return nil, fmt.Errorf("failed to marshal default config: %w", err)
			}
			err = os.WriteFile(configPath, bytes, 0644)
			if err != nil {
				log.Error().Err(err).Str("path", configPath).Msg("Failed to write default config file")
				return nil, fmt.Errorf("failed to write default config file: %w", err)
			}
			return defaultConfig, nil
		}
		log.Error().Err(err).Str("path", configPath).Msg("Failed to read config file")
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := new(T)
	err = unmarshalConfig(bytes, configPath, config)
	if err != nil {
		log.Error().Err(err).Str("path", configPath).Msg("Failed to unmarshal config file")
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	MergeConfigs(config, defaultConfig)
	return config, nil
}

// marshalConfig marshals the config based on the file extension
func marshalConfig[T any](config *T, configPath string) ([]byte, error) {
	switch filepath.Ext(configPath) {
	case ".json":
		return json.MarshalIndent(config, "", "    ")
	case ".yaml", ".yml":
		return yaml.Marshal(config)
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", configPath)
	}
}

// unmarshalConfig unmarshals the config based on the file extension
func unmarshalConfig[T any](data []byte, configPath string, config *T) error {
	switch filepath.Ext(configPath) {
	case ".json":
		return json.Unmarshal(data, config)
	case ".yaml", ".yml":
		return yaml.Unmarshal(data, config)
	default:
		return fmt.Errorf("unsupported file extension: %s", configPath)
	}
}

// ensureConfigDirExist ensures the config directory exists
// If it does not exist, it will be created
func ensureConfigDirExist() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return os.MkdirAll(configDir, 0755)
	}
	return nil
}

// GetConfigDir returns the path to the config directory
func GetConfigDir() (string, error) {
	configPath, err := xdg.ConfigFile(ConfigDir)
	if err != nil {
		log.Error().Err(err).Msg("Error while getting config path directory")
		return "", err
	}
	return configPath, nil
}
