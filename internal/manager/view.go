package manager

import (
	"sync"

	"github.com/gdamore/tcell/v2"
)

type (
	ViewIdentifier string
	// EventMsg is a wrapper for tcell.EventKey that also contains
	// the sender of the event
	EventMsg struct {
		*tcell.EventKey
		Sender  ViewIdentifier
		Message interface{}
	}

	// ViewManager is a helper to manage different Views
	// and their key handlers, so that only the key handlers of the
	// current view are executed
	ViewManager struct {
		// viewStack is responsible for keeping track of the current view
		// that is being rendered
		viewStack []ViewIdentifier
		mutex     sync.Mutex
		listeners map[ViewIdentifier]chan EventMsg
	}
)

// NewViewManager creates a new ViewManager
func NewViewManager() *ViewManager {
	return &ViewManager{
		viewStack: make([]ViewIdentifier, 0),
		mutex:     sync.Mutex{},
		listeners: make(map[ViewIdentifier]chan EventMsg),
	}
}

// PushView adds a new view to the view stack
func (eh *ViewManager) PushView(view ViewIdentifier, subviews ...ViewIdentifier) {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	if len(eh.viewStack) > 0 && eh.viewStack[len(eh.viewStack)-1] == view {
		return
	}
	eh.viewStack = append(eh.viewStack, view)
}

// PopView removes the latest view from the view stack
func (eh *ViewManager) PopView() {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	if len(eh.viewStack) > 0 {
		eh.viewStack = eh.viewStack[:len(eh.viewStack)-1]
	}
}

// CurrentView returns the current view
func (eh *ViewManager) CurrentView() ViewIdentifier {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	if len(eh.viewStack) == 0 {
		return ""
	}
	return eh.viewStack[len(eh.viewStack)-1]
}

// Subscribe subscribes to events from a specific view
func (eh *ViewManager) Subscribe(view ViewIdentifier) chan EventMsg {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	listener := make(chan EventMsg, 1)
	eh.listeners[view] = listener
	return listener
}

// Unsubscribe unsubscribes from events from a specific view
func (eh *ViewManager) Unsubscribe(view ViewIdentifier, listener chan EventMsg) {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	delete(eh.listeners, view)
}

// Broadcast sends an event to all listeners
func (eh *ViewManager) Broadcast(event EventMsg) {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	for _, listener := range eh.listeners {
		listener <- event
	}
}

// SendTo sends an event to a specific view
func (eh *ViewManager) SendTo(view ViewIdentifier, event EventMsg) {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	if listener, exists := eh.listeners[view]; exists {
		listener <- event
	}
}
