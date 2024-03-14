package manager

import (
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
)

const (
	GlobalComponent = "Global"
)

type (
	// EventMsg is a wrapper for tcell.EventKey that also contains
	// the sender of the event
	EventMsg struct {
		*tcell.EventKey
		Sender tview.Identifier
	}

	// Listener
	Listener struct {
		Component tview.Identifier
		Channel   chan EventMsg
	}

	// ComponentManager is a helper to manage different Components
	// and their key handlers, so that only the key handlers of the
	// current component are executed
	ComponentManager struct {
		// componentStack is responsible for keeping track of the current component
		// that is being rendered
		componentStack []tview.Identifier
		mutex          sync.Mutex
		listeners      map[tview.Identifier]chan EventMsg
	}
)

// NewComponentManager creates a new ComponentManager
func NewComponentManager() *ComponentManager {
	return &ComponentManager{
		componentStack: make([]tview.Identifier, 0),
		mutex:          sync.Mutex{},
		listeners:      make(map[tview.Identifier]chan EventMsg),
	}
}

// PushComponent adds a new component to the component stack
func (eh *ComponentManager) PushComponent(component tview.Identifier, subcomponents ...tview.Identifier) {
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
func (eh *ComponentManager) CurrentComponent() tview.Identifier {
	log.Info().Msgf("Current Component: %v", eh.componentStack)
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	if len(eh.componentStack) == 0 {
		return ""
	}
	return eh.componentStack[len(eh.componentStack)-1]
}

// Subscribe subscribes to events from a specific component
func (eh *ComponentManager) Subscribe(component tview.Identifier) chan EventMsg {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	listener := make(chan EventMsg, 1)
	eh.listeners[component] = listener
	return listener
}

// Unsubscribe unsubscribes from events from a specific component
func (eh *ComponentManager) Unsubscribe(component tview.Identifier, listener chan EventMsg) {
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
func (eh *ComponentManager) SendTo(component tview.Identifier, event EventMsg) {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	if listener, exists := eh.listeners[component]; exists {
		listener <- event
	}
}
