package manager

import (
	"sync"

	"github.com/gdamore/tcell/v2"
)

const (
	// GlobalComponent is a special component that is used for key bindings that
	// are not specific to any component.
	GlobalComponent Component = "Global"
)

var ()

// KeyAction defines a function that will be executed when a key is pressed.
type KeyAction func(e *tcell.EventKey) *tcell.EventKey

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

// GetKeysForComponent returns all the key bindings for a specific component.
func (km *KeyManager) GetKeysForComponent(comp Component) []KeyBinding {
	km.mutex.RLock()
	defer km.mutex.RUnlock()

	return km.bindings[comp]
}

// GetGlobalKeys returns all the key bindings that are not specific to any component.
func (km *KeyManager) GetGlobalKeys() []KeyBinding {
	km.mutex.RLock()
	defer km.mutex.RUnlock()

	return km.bindings[GlobalComponent]
}
