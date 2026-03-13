package tui

import (
	"github.com/kopecmaciej/vi-mongo/assets"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/tui/modal"
	"gopkg.in/yaml.v3"
)

type changelogYAML struct {
	Entries []changelogEntryYAML `yaml:"entries"`
}

type changelogEntryYAML struct {
	Version   string   `yaml:"version"`
	Breaking  bool     `yaml:"breaking"`
	Title     string   `yaml:"title"`
	Changes   []string `yaml:"changes"`
	Migration string   `yaml:"migration"`
}

var migrationRegistry = map[string]func() error{
	"keybindings_json_to_yaml": config.RunKeybindingsMigration,
}

func loadChangelog() []modal.ChangelogEntry {
	var raw changelogYAML
	if err := yaml.Unmarshal(assets.Changelog, &raw); err != nil {
		return nil
	}

	entries := make([]modal.ChangelogEntry, 0, len(raw.Entries))
	for _, e := range raw.Entries {
		entry := modal.ChangelogEntry{
			Version:  e.Version,
			Breaking: e.Breaking,
			Title:    e.Title,
			Changes:  e.Changes,
		}
		if fn, ok := migrationRegistry[e.Migration]; ok {
			entry.MigrationFn = fn
		}
		entries = append(entries, entry)
	}
	return entries
}
