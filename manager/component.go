package manager

import (
	"sync"

	"github.com/gdamore/tcell/v2"
)

type (
	// Component is a type for component names
	Component string

	// EventMsg is a wrapper for tcell.EventKey that also contains
	// the sender of the event
	EventMsg struct {
		*tcell.EventKey
		Sender Component
	}

	// Listener
	Listener struct {
		Component Component
		Channel   chan EventMsg
	}

	// ComponentManager is a helper to manage different Components
	// and their key handlers, so that only the key handlers of the
	// current component are executed
	ComponentManager struct {
		// componentStack is responsible for keeping track of the current component
		// that is being rendered
		componentStack []Component
		// subcomponentsMap is a map that contains all the subcomponents of a component
		// it's used mainly for managing keybindings
		subcomponentsMap map[Component][]Component
		mutex            sync.Mutex
		listeners        map[Component]chan EventMsg
		KeyManager       *KeyManager
	}
)

// NewComponentManager creates a new ComponentManager
func NewComponentManager() *ComponentManager {
	return &ComponentManager{
		componentStack:   make([]Component, 0),
		subcomponentsMap: make(map[Component][]Component),
		mutex:            sync.Mutex{},
		listeners:        make(map[Component]chan EventMsg),
		KeyManager:       NewKeyManager(),
	}
}

// PushComponent adds a new component to the component stack
func (eh *ComponentManager) PushComponent(component Component, subcomponents ...Component) {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	if len(eh.componentStack) > 0 && eh.componentStack[len(eh.componentStack)-1] == component {
		return
	}
	eh.componentStack = append(eh.componentStack, component)
}

// PopComponent removes the latest component from the component stack
func (eh *ComponentManager) PopComponent() {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	if len(eh.componentStack) > 0 {
		eh.componentStack = eh.componentStack[:len(eh.componentStack)-1]
	}
}

// CurrentComponent returns the current component
func (eh *ComponentManager) CurrentComponent() Component {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	if len(eh.componentStack) == 0 {
		return ""
	}
	return eh.componentStack[len(eh.componentStack)-1]
}

// AddSubcomponents adds subcomponents to a component
func (eh *ComponentManager) AddSubcomponents(component Component, subcomponents []Component) {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()

	eh.subcomponentsMap[component] = append(eh.subcomponentsMap[component], subcomponents...)
}

// GetSubcomponents returns the subcomponents of a component
func (eh *ComponentManager) GetSubcomponents(component Component) ([]Component, bool) {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()

	if subcomponents, exists := eh.subcomponentsMap[component]; exists {
		return subcomponents, true
	}
	return nil, false
}

// SetKeyHandlerForComponent is a helper function to set a key handler for a specific component
func (eh *ComponentManager) SetKeyHandlerForComponent(component Component) func(key tcell.Key, r rune, description string, action KeyAction) {
	return func(key tcell.Key, r rune, description string, action KeyAction) {
		name := ""
		if r != 0 {
			name = string(r)
		} else {
			name = tcell.KeyNames[key]
		}
		eh.KeyManager.RegisterKeyBinding(component, key, r, name, description, action)
	}
}

// HandleKeyEvent handles a key event based on the current component
func (eh *ComponentManager) HandleKeyEvent(e *tcell.EventKey) *tcell.EventKey {
	component := eh.CurrentComponent()
	keys := eh.KeyManager.GetKeysForComponent(component)
	for _, k := range keys {
		if (e.Key() == tcell.KeyRune && k.Rune == e.Rune()) || (k.Key == e.Key() && k.Rune == 0) {
			return k.Action(e)
		}
	}

	// handle subcomponents
	subcomponents, _ := eh.GetSubcomponents(component)
	for _, subcomponent := range subcomponents {
		subKeys := eh.KeyManager.GetKeysForComponent(subcomponent)
		for _, k := range subKeys {
			if (e.Key() == tcell.KeyRune && k.Rune == e.Rune()) || (k.Key == e.Key() && k.Rune == 0) {
				return k.Action(e)
			}
		}
	}

	// If no key was found for the current component, check the global keys
	globalKeys := eh.KeyManager.GetGlobalKeys()
	for _, k := range globalKeys {
		if (e.Key() == tcell.KeyRune && k.Rune == e.Rune()) || (k.Key == e.Key() && k.Rune == 0) {
			return k.Action(e)
		}
	}

	return e
}

// Subscribe subscribes to events from a specific component
func (eh *ComponentManager) Subscribe(component Component) chan EventMsg {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	listener := make(chan EventMsg, 1)
	eh.listeners[component] = listener
	return listener
}

// Unsubscribe unsubscribes from events from a specific component
func (eh *ComponentManager) Unsubscribe(component Component, listener chan EventMsg) {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	delete(eh.listeners, component)
}

// Broadcast sends an event to a specific component
func (eh *ComponentManager) Broadcast(event EventMsg) {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	for _, listener := range eh.listeners {
		listener <- event
	}
}

// BroadcastTo sends an event to a specific component
func (eh *ComponentManager) SendTo(component Component, event EventMsg) {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	if listener, exists := eh.listeners[component]; exists {
		listener <- event
	}
}
