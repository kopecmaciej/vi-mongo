// MIGRATION: JSON → YAML keybindings (introduced v0.1.36)
// TODO: Remove after v0.2.x
//
// Standalone script for users who want to migrate their keybindings.json
// to keybindings.yaml manually, outside of the automatic in-app migration.
//
// Usage:
//
//	go run ./scripts/migrate_keybindings
//	go run ./scripts/migrate_keybindings --json /custom/path/keybindings.json
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"
)

func main() {
	jsonPath := flag.String("json", defaultJSONPath(), "path to keybindings.json")
	flag.Parse()

	yamlPath := (*jsonPath)[:len(*jsonPath)-len(filepath.Ext(*jsonPath))] + ".yaml"

	if _, err := os.Stat(*jsonPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: %s not found\n", *jsonPath)
		os.Exit(1)
	}

	if _, err := os.Stat(yamlPath); err == nil {
		fmt.Fprintf(os.Stderr, "Error: %s already exists, remove it first if you want to re-migrate\n", yamlPath)
		os.Exit(1)
	}

	data, err := os.ReadFile(*jsonPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", *jsonPath, err)
		os.Exit(1)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	yamlData, err := yaml.Marshal(raw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling YAML: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(yamlPath, yamlData, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", yamlPath, err)
		os.Exit(1)
	}

	if err := os.Rename(*jsonPath, *jsonPath+".bak"); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not rename %s to .bak: %v\n", *jsonPath, err)
	}

	fmt.Printf("Migrated: %s -> %s\n", *jsonPath, yamlPath)
	fmt.Printf("Backup:   %s.bak\n", *jsonPath)
}

func defaultJSONPath() string {
	configDir, err := xdg.ConfigFile("vi-mongo")
	if err != nil {
		return "keybindings.json"
	}
	return filepath.Join(configDir, "keybindings.json")
}
