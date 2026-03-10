// MIGRATION: JSON → YAML keybindings (introduced v2.0.0)
// TODO: Remove this file after v2.x once adoption is sufficient.

package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/kopecmaciej/vi-mongo/internal/util"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// RunKeybindingsMigration converts keybindings.json to keybindings.yaml.
// It is a no-op when the JSON file does not exist or the YAML already exists.
func RunKeybindingsMigration() error {
	configDir, err := util.GetConfigDir()
	if err != nil {
		return err
	}

	jsonPath := configDir + "/keybindings.json"
	yamlPath := configDir + "/keybindings.yaml"

	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		return nil // nothing to migrate
	}
	if _, err := os.Stat(jsonPath + ".bak"); err == nil {
		return nil // already migrated
	}

	if err := migrateJSONKeybindingsToYAML(jsonPath, yamlPath); err != nil {
		return err
	}

	log.Info().Str("from", jsonPath).Str("to", yamlPath).Msg("Keybindings migrated to YAML")
	return nil
}

func migrateJSONKeybindingsToYAML(jsonPath, yamlPath string) error {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", jsonPath, err)
	}

	// Decode into the typed struct so yaml.Node.Encode preserves field
	// declaration order (Global, Help, Welcome, …) instead of Go map
	// iteration order (alphabetical).
	var kb KeyBindings
	if err := json.Unmarshal(data, &kb); err != nil {
		return fmt.Errorf("parse JSON: %w", err)
	}
	normalizeCtrlKeys(reflect.ValueOf(&kb).Elem())

	var node yaml.Node
	if err := node.Encode(&kb); err != nil {
		return fmt.Errorf("encode YAML node: %w", err)
	}
	setSequencesFlowStyle(&node)

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(4)
	if err := enc.Encode(&node); err != nil {
		return fmt.Errorf("marshal YAML: %w", err)
	}

	content := append([]byte(keybindingsFileHeader), buf.Bytes()...)
	if err := os.WriteFile(yamlPath, content, 0600); err != nil {
		return fmt.Errorf("write %s: %w", yamlPath, err)
	}

	if err := os.Rename(jsonPath, jsonPath+".bak"); err != nil {
		log.Warn().Err(err).Str("path", jsonPath).Msg("Could not rename old JSON keybindings to .bak")
	}

	return nil
}

// normalizeCtrlKeys lowercases the letter in Ctrl+X combos stored in Key.Keys
// so the config reflects what physically works on the keyboard.
func normalizeCtrlKeys(val reflect.Value) {
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if field.Type() == reflect.TypeOf(Key{}) {
			key := field.Addr().Interface().(*Key)
			for j, k := range key.Keys {
				if strings.HasPrefix(k, "Ctrl+") && len(k) == 6 {
					key.Keys[j] = "Ctrl+" + strings.ToLower(string(k[5]))
				}
			}
		} else if field.Kind() == reflect.Struct {
			normalizeCtrlKeys(field)
		}
	}
}

func setSequencesFlowStyle(node *yaml.Node) {
	if node.Kind == yaml.SequenceNode {
		node.Style = yaml.FlowStyle
	}
	for _, child := range node.Content {
		setSequencesFlowStyle(child)
	}
}
