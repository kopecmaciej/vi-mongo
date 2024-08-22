package component

import (
	"sync"

	"github.com/kopecmaciej/mongui/internal/manager"
	"github.com/kopecmaciej/mongui/internal/mongo"
)

// BaseComponent is a base struct for all components.
// It contains all the basic fields and functions that are used by all components.
// It also implements the BaseComponent interface.
type BaseComponent struct {
	// enabled is a flag that indicates if the component is enabled.
	enabled bool

	// name is the name of the component.
	// It's used mainly for managing avaliable keybindings.
	identifier manager.Component
	// app is a pointer to the main app.
	// It's used for accessing app focus, root page etc.
	app *App

	// dao is a pointer to the mongo dao.
	dao *mongo.Dao

	// initFunc is a function that is called when the component is initialized.
	// It's main purpose is to run all the initialization functions of the subcomponents.
	afterInitFunc func() error

	// listener is a channel that is used to receive events from the app.
	listener chan manager.EventMsg

	// mutex is a mutex that is used to synchronize the component.
	mutex sync.Mutex
}

// NewBaseComponent is a constructor for the BaseComponent struct.
func NewBaseComponent(identifier string) *BaseComponent {
	return &BaseComponent{
		identifier: manager.Component(identifier),
	}
}

// Init is a function that is called when the component is initialized.
// If custom initialization is needed, this function should be overriden.
func (c *BaseComponent) Init(app *App) error {
	if c.app != nil && c.identifier != "" {
		return nil
	}

	c.app = app
	if app.Dao != nil {
		c.dao = app.Dao
	}

	if c.afterInitFunc != nil {
		return c.afterInitFunc()
	}

	return nil
}

// Enable sets the enabled flag.
func (c *BaseComponent) Enable() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.enabled = true
	c.app.Manager.PushComponent(c.identifier)
}

// Disable unsets the enabled flag.
func (c *BaseComponent) Disable() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.enabled = false
	c.app.Manager.PopComponent()
}

// Toggle toggles the enabled flag.
func (c *BaseComponent) Toggle() {
	if c.IsEnabled() {
		c.Disable()
	} else {
		c.Enable()
	}
}

// IsEnabled returns the enabled flag.
func (c *BaseComponent) IsEnabled() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.enabled
}

// SetAfterInitFunc sets the optional function that will be run at the end of the Init function.
func (c *BaseComponent) SetAfterInitFunc(afterInitFunc func() error) {
	c.afterInitFunc = afterInitFunc
}

// GetComponent returns the component.
func (c *BaseComponent) GetComponent() *BaseComponent {
	return c
}

// GetIdentifier returns the identifier of the component.
func (c *BaseComponent) GetIdentifier() manager.Component {
	return c.identifier
}

// Subscribe subscribes to the component events.
func (c *BaseComponent) Subscribe() {
	c.listener = c.app.Manager.Subscribe(c.identifier)
}

// Broadcast sends an event to all listeners.
func (c *BaseComponent) BroadcastEvent(event manager.EventMsg) {
	c.app.Manager.Broadcast(event)
}

// SendToComponent sends an event to the component.
func (c *BaseComponent) SendToComponent(component manager.Component, event manager.EventMsg) {
	c.app.Manager.SendTo(component, event)
}
