package util

import (
	"reflect"
	"testing"
)

type TestConfig struct {
	KeyField Key
	Nested   NestedConfig
}

type NestedConfig struct {
	KeyField Key
}

type Key struct {
	Keys        []string
	Runes       []string
	Description string
}

func TestMergeKeybindings(t *testing.T) {
	tests := []struct {
		name     string
		loaded   TestConfig
		default_ TestConfig
		expected TestConfig
	}{
		{
			name: "empty loaded config should use defaults",
			default_: TestConfig{
				KeyField: Key{
					Keys:        []string{"Ctrl+C", "Ctrl+X"},
					Runes:       []string{"q"},
					Description: "Close application",
				},
			},
			loaded: TestConfig{},
			expected: TestConfig{
				KeyField: Key{
					Keys:        []string{"Ctrl+C", "Ctrl+X"},
					Runes:       []string{"q"},
					Description: "Close application",
				},
			},
		},
		{
			name:     "empty default config should use loaded",
			default_: TestConfig{},
			loaded: TestConfig{
				KeyField: Key{
					Keys:        []string{"Ctrl+V"},
					Runes:       []string{"q"},
					Description: "Paste",
				},
			},
			expected: TestConfig{
				KeyField: Key{
					Keys:        []string{"Ctrl+V"},
					Runes:       []string{"q"},
					Description: "Paste",
				},
			},
		},
		{
			name: "loaded values with only keys should override defaults",
			default_: TestConfig{
				KeyField: Key{
					Keys:        []string{"Ctrl+O"},
					Runes:       []string{"q"},
					Description: "Default action",
				},
			},
			loaded: TestConfig{
				KeyField: Key{
					Keys:        []string{"Ctrl+N", "Ctrl+O"},
					Description: "Custom action",
				},
			},
			expected: TestConfig{
				KeyField: Key{
					Keys:        []string{"Ctrl+N", "Ctrl+O"},
					Runes:       nil,
					Description: "Custom action",
				},
			},
		},
		{
			name: "loaded values with only runes should override defaults",
			default_: TestConfig{
				KeyField: Key{
					Runes: []string{"p"},
				},
			},
			loaded: TestConfig{
				KeyField: Key{
					Runes: []string{"q", "w"},
				},
			},
			expected: TestConfig{
				KeyField: Key{
					Runes: []string{"q", "w"},
				},
			},
		},
		{
			name: "nested default values should be merged and overridden by loaded, missing fields should be set to defaults",
			default_: TestConfig{
				Nested: NestedConfig{
					KeyField: Key{
						Keys:        []string{"Ctrl+X"},
						Runes:       []string{"q"},
						Description: "Default nested",
					},
				},
				KeyField: Key{
					Runes: []string{"P"},
				},
			},
			loaded: TestConfig{
				Nested: NestedConfig{
					KeyField: Key{
						Runes: []string{"x"},
					},
				},
			},
			expected: TestConfig{
				Nested: NestedConfig{
					KeyField: Key{
						Keys:        nil,
						Runes:       []string{"x"},
						Description: "",
					},
				},
				KeyField: Key{
					Runes: []string{"P"},
				},
			},
		},
		{
			name: "nested loaded values should override defaults",
			default_: TestConfig{
				Nested: NestedConfig{
					KeyField: Key{
						Keys:        []string{"Ctrl+X"},
						Description: "Nuke my system",
					},
				},
			},
			loaded: TestConfig{
				Nested: NestedConfig{
					KeyField: Key{
						Keys:        []string{"Ctrl+Y"},
						Runes:       []string{"y"},
						Description: "Protect my system",
					},
				},
				KeyField: Key{
					Runes:       []string{"P"},
					Description: "I'm just a random key",
				},
			},
			expected: TestConfig{
				Nested: NestedConfig{
					KeyField: Key{
						Keys:        []string{"Ctrl+Y"},
						Runes:       []string{"y"},
						Description: "Protect my system",
					},
				},
				KeyField: Key{
					Runes:       []string{"P"},
					Description: "I'm just a random key",
				},
			},
		},
		{
			name:     "loaded values with empty key struct should be merged with defaults",
			default_: TestConfig{},
			loaded: TestConfig{
				KeyField: Key{},
			},
			expected: TestConfig{
				KeyField: Key{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loaded := &tt.loaded
			MergeConfigs(loaded, &tt.default_)

			if !reflect.DeepEqual(*loaded, tt.expected) {
				t.Errorf("MergeConfigs() = %+v, want %+v", *loaded, tt.expected)
			}
		})
	}
}
