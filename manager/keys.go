package manager

import (
	"sync"

	"github.com/gdamore/tcell/v2"
)

// KeyAction defines a function that will be executed when a key is pressed.
type KeyAction func()

// KeyBinding represents a single key and its action.
type KeyBinding struct {
	Key         tcell.Key // The key itself, e.g., tcell.KeyCtrlD
	Rune        rune      // The rune (character), use 0 if not applicable
	Name        string    // The name of the key
	Description string    // Description of what the key does
	Action      KeyAction // The action to execute when the key or rune is pressed
}

// KeyManager holds the key bindings for each component.
type KeyManager struct {
	mutex    sync.RWMutex
	bindings map[Component][]KeyBinding
}

// NewKeys creates a new KeyManager.
func NewKeyManager() *KeyManager {
	return &KeyManager{
		bindings: make(map[Component][]KeyBinding),
	}
}

// RegisterKeyBinding assigns a key binding to a component.
func (km *KeyManager) RegisterKeyBinding(comp Component, key tcell.Key, rune rune, name, description string, action KeyAction) {
	km.mutex.Lock()
	defer km.mutex.Unlock()

	// Add the key binding to the specific component.
	km.bindings[comp] = append(km.bindings[comp], KeyBinding{
		Key:         key,
		Rune:        rune,
		Name:        name,
		Description: description,
		Action:      action,
	})
}

// This method integrates the key manager with your TUI components.
func (km *KeyManager) SetInputCapture(comp Component) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		km.mutex.RLock()
		defer km.mutex.RUnlock()

		for _, binding := range km.bindings[comp] {
			if (event.Key() == binding.Key || (event.Key() == tcell.KeyRune && event.Rune() == binding.Rune)) && binding.Action != nil {
				binding.Action()
				return nil
			}
		}

		return event
	}
}

// GetKeysForComponent returns all the key bindings for a specific component.
func (km *KeyManager) GetKeysForComponent(comp Component) []KeyBinding {
	km.mutex.RLock()
	defer km.mutex.RUnlock()

	return km.bindings[comp]
}
