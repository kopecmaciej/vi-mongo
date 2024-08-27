package manager

import (
	"sync"

	"github.com/gdamore/tcell/v2"
)

type (
	ElementId string
	// EventMsg is a wrapper for tcell.EventKey that also contains
	// the sender of the event
	EventMsg struct {
		*tcell.EventKey
		Sender  ElementId
		Message interface{}
	}

	// ElementManager is a helper to manage different Views
	// and their key handlers, so that only the key handlers of the
	// current view are executed
	ElementManager struct {
		// elementStack is responsible for keeping track of the current view
		// that is being rendered
		elementStack []ElementId
		mutex        sync.Mutex
		listeners    map[ElementId]chan EventMsg
	}
)

// NewElementManager creates a new ElementManager
func NewElementManager() *ElementManager {
	return &ElementManager{
		elementStack: make([]ElementId, 0),
		mutex:        sync.Mutex{},
		listeners:    make(map[ElementId]chan EventMsg),
	}
}

// PushElement adds a new element to the element stack
func (eh *ElementManager) PushElement(element ElementId, subelements ...ElementId) {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	if len(eh.elementStack) > 0 && eh.elementStack[len(eh.elementStack)-1] == element {
		return
	}
	eh.elementStack = append(eh.elementStack, element)
}

// PopElement removes the latest element from the element stack
func (eh *ElementManager) PopElement() {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	if len(eh.elementStack) > 0 {
		eh.elementStack = eh.elementStack[:len(eh.elementStack)-1]
	}
}

// CurrentElement returns the current element
func (eh *ElementManager) CurrentElement() ElementId {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	if len(eh.elementStack) == 0 {
		return ""
	}
	return eh.elementStack[len(eh.elementStack)-1]
}

// Subscribe subscribes to events from a specific element
func (eh *ElementManager) Subscribe(element ElementId) chan EventMsg {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	listener := make(chan EventMsg, 1)
	eh.listeners[element] = listener
	return listener
}

// Unsubscribe unsubscribes from events from a specific element
func (eh *ElementManager) Unsubscribe(element ElementId, listener chan EventMsg) {
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
func (eh *ElementManager) SendTo(element ElementId, event EventMsg) {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	if listener, exists := eh.listeners[element]; exists {
		listener <- event
	}
}
