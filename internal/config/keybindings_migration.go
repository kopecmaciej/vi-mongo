// MIGRATION: JSON → YAML keybindings (introduced v2.0.0)
// TODO: Remove this file after v2.x once adoption is sufficient.
//
// This file is intentionally isolated so it can be deleted as a unit.
// The migration is triggered by the changelog prompt shown on first startup
// after upgrading.

package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

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
	// Check if migration already ran by looking for the backup file.
	// We cannot rely on keybindings.yaml existing because LoadKeybindings()
	// creates it eagerly with defaults before the changelog prompt is shown.
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

	// Parse as a raw map to preserve all key names without relying on struct tags.
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("parse JSON: %w", err)
	}

	var node yaml.Node
	if err := node.Encode(raw); err != nil {
		return fmt.Errorf("encode YAML node: %w", err)
	}
	setSequencesFlowStyle(&node)

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(4)
	if err := enc.Encode(&node); err != nil {
		return fmt.Errorf("marshal YAML: %w", err)
	}

	if err := os.WriteFile(yamlPath, buf.Bytes(), 0600); err != nil {
		return fmt.Errorf("write %s: %w", yamlPath, err)
	}

	if err := os.Rename(jsonPath, jsonPath+".bak"); err != nil {
		log.Warn().Err(err).Str("path", jsonPath).Msg("Could not rename old JSON keybindings to .bak")
	}

	return nil
}

func setSequencesFlowStyle(node *yaml.Node) {
	if node.Kind == yaml.SequenceNode {
		node.Style = yaml.FlowStyle
	}
	for _, child := range node.Content {
		setSequencesFlowStyle(child)
	}
}
