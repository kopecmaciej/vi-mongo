package manager

import (
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
)

const (
	FocusChanged           MessageType = "focus_changed"
	StyleChanged           MessageType = "style_changed"
	UpdateAutocompleteKeys MessageType = "update_autocomplete"
	UpdateQueryBar         MessageType = "update_query_bar"
)

type (
	MessageType string
	// Messages is a list of messages that can be sent to the manager
	Message struct {
		Type MessageType
		Data any
	}

	// EventMsg is a wrapper for tcell.EventKey that also contains
	// the sender of the event
	EventMsg struct {
		*tcell.EventKey
		Sender  tview.Identifier
		Message Message
	}

	// ElementManager is a helper to manage different Elements
	// and their key handlers, so that only the key handlers of the
	// current element are executed
	ElementManager struct {
		mutex     sync.Mutex
		listeners map[tview.Identifier]chan EventMsg
	}
)

// NewElementManager creates a new ElementManager
func NewElementManager() *ElementManager {
	return &ElementManager{
		mutex:     sync.Mutex{},
		listeners: make(map[tview.Identifier]chan EventMsg),
	}
}

// Subscribe subscribes to events from a specific element
func (eh *ElementManager) Subscribe(element tview.Identifier) chan EventMsg {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	listener := make(chan EventMsg, 1)
	eh.listeners[element] = listener
	return listener
}

// Unsubscribe unsubscribes from events from a specific element
func (eh *ElementManager) Unsubscribe(element tview.Identifier, listener chan EventMsg) {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	delete(eh.listeners, element)
}

// Broadcast sends an event to all listeners
func (eh *ElementManager) Broadcast(event EventMsg) {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	for _, listener := range eh.listeners {
		listener <- event
	}
}

// SendTo sends an event to a specific element
func (eh *ElementManager) SendTo(element tview.Identifier, event EventMsg) {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	if listener, exists := eh.listeners[element]; exists {
		listener <- event
	}
}
