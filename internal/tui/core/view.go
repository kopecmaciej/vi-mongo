package core

import (
	"sync"

	"github.com/kopecmaciej/vi-mongo/internal/manager"
	"github.com/kopecmaciej/vi-mongo/internal/mongo"
)

// BaseElement is a base struct for all visable elements.
// It contains all the basic fields and functions that are used by all visable elements.
type BaseElement struct {
	// enabled is a flag that indicates if the view is enabled.
	enabled bool

	// name is the name of the view.
	// It's used mainly for managing avaliable keybindings.
	identifier manager.ElementId
	// App is a pointer to the main App.
	// It's used for accessing App focus, root page etc.
	App *App

	// dao is a pointer to the mongo dao.
	Dao *mongo.Dao

	// afterInitFunc is a function that is called when the view is initialized.
	// It's main purpose is to run all the initialization functions of the subviews.
	afterInitFunc func() error

	// Listener is a channel that is used to receive events from the app.
	Listener chan manager.EventMsg

	// mutex is a mutex that is used to synchronize the view.
	mutex sync.Mutex
}

// NewBaseElement is a constructor for the BaseElement struct.
func NewBaseElement(identifier string) *BaseElement {
	return &BaseElement{
		identifier: manager.ElementId(identifier),
	}
}

// Init is a function that is called when the view is initialized.
// If custom initialization is needed, this function should be overriden.
func (c *BaseElement) Init(app *App) error {
	if c.App != nil && c.identifier != "" {
		return nil
	}

	c.App = app
	if app.GetDao() != nil {
		c.Dao = app.GetDao()
	}

	if c.afterInitFunc != nil {
		return c.afterInitFunc()
	}

	return nil
}

// Enable sets the enabled flag.
func (c *BaseElement) Enable() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.enabled = true
	c.App.GetManager().PushElement(c.identifier)
}

// Disable unsets the enabled flag.
func (c *BaseElement) Disable() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.enabled = false
	c.App.GetManager().PopElement()
}

// Toggle toggles the enabled flag.
func (c *BaseElement) Toggle() {
	if c.IsEnabled() {
		c.Disable()
	} else {
		c.Enable()
	}
}

// IsEnabled returns the enabled flag.
func (c *BaseElement) IsEnabled() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.enabled
}

// SetAfterInitFunc sets the optional function that will be run at the end of the Init function.
func (c *BaseElement) SetAfterInitFunc(afterInitFunc func() error) {
	c.afterInitFunc = afterInitFunc
}

// GetAfterInitFunc returns the optional function that will be run at the end of the Init function.
func (c *BaseElement) GetAfterInitFunc() func() error {
	return c.afterInitFunc
}

// GetIdentifier returns the identifier of the view.
func (c *BaseElement) GetIdentifier() manager.ElementId {
	return c.identifier
}

// Subscribe subscribes to the view events.
func (c *BaseElement) Subscribe() {
	c.Listener = c.App.GetManager().Subscribe(c.identifier)
}

// Broadcast sends an event to all listeners.
func (c *BaseElement) BroadcastEvent(event manager.EventMsg) {
	c.App.GetManager().Broadcast(event)
}

// SendToView sends an event to the view.
func (c *BaseElement) SendToView(view manager.ElementId, event manager.EventMsg) {
	c.App.GetManager().SendTo(view, event)
}
