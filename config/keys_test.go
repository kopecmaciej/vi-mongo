package config

import (
	"reflect"
	"testing"
)

func TestGetKeysForComponent(t *testing.T) {
	// Initialize a KeyBindings instance (you can use NewKeyBindings or define directly)
	kb := NewKeyBindings()

	// Define test cases
	tests := []struct {
		component      string
		expectedKeys   []Key
		expectingError bool
	}{
		{
			component: "ConnectorForm",
			expectedKeys: []Key{
				{
					Keys:        []string{"Up"},
					Description: "Move form focus up",
				},
				{
					Keys:        []string{"Down"},
					Description: "Move form focus down",
				},
				{
					Keys:        []string{"Ctrl+S"},
					Description: "Save connection",
				},
				{
					Keys:        []string{"Esc"},
					Description: "Focus Connection List",
				},
			},
		},
	}

	for _, test := range tests {
		keys, err := kb.GetKeysForComponent(test.component)
		if test.expectingError {
			if err == nil {
				t.Errorf("Expected error for component %s, but got none", test.component)
			}
			continue // Skip further checks if an error was expected
		}
		if err != nil {
			t.Errorf("Did not expect an error for component %s, but got one: %v", test.component, err)
			continue
		}
		if !reflect.DeepEqual(keys, test.expectedKeys) {
			t.Errorf("Expected keys %+v for component %s, but got %+v", test.expectedKeys, test.component, keys)
		}
	}
}
