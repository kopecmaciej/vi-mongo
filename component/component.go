package component

import (
	"context"

	"github.com/kopecmaciej/mongui/manager"
	"github.com/kopecmaciej/mongui/mongo"
)

// every component should implement this interface
// it's used for managing components in the app
type ComponentRenderer interface {
	// Render is a function that renders the component.
	Render(ctx context.Context) error
}

// Component is a base struct for all components.
// It contains all the basic fields and functions that are used by all components.
// It also implements the Component interface.
type Component struct {
	// app is a pointer to the main app.
	// It's used for accessing app focus, root page etc.
	app *App

	// dao is a pointer to the mongo dao.
	dao *mongo.Dao

	// name is the name of the component.
	// It's used mainly for managing avaliable keybindings.
	identifier manager.Component

	// initFunc is a function that is called when the component is initialized.
	// It's main purpose is to run all the initialization functions of the subcomponents.
	afterInitFunc func() error
}

// NewComponent is a constructor for the Component struct.
func NewComponent(identifier string) *Component {
	return &Component{
		identifier: manager.Component(identifier),
	}
}

// Init is a function that is called when the component is initialized.
// If custom initialization is needed, this function should be overriden.
func (c *Component) Init(app *App) error {
	c.app = app
	if app.Dao != nil {
		c.dao = app.Dao
	}

	if c.afterInitFunc != nil {
		return c.afterInitFunc()
	}

	return nil
}

// SetAfterInitFunc sets the optional function that will be run at the end of the Init function.
func (c *Component) SetAfterInitFunc(afterInitFunc func() error) {
	c.afterInitFunc = afterInitFunc
}

// GetComponent returns the component.
func (c *Component) GetComponent() *Component {
	return c
}

// GetIdentifier returns the identifier of the component.
func (c *Component) GetIdentifier() manager.Component {
	return c.identifier
}

// Subscribe subscribes the component to the global events.
func (c *Component) Subscribe() {
	c.app.Manager.Subscribe(c.identifier)
}

// SendEvent sends an event to the app.
func (c *Component) BroadcastEvent(event manager.EventMsg) {
	c.app.Manager.Broadcast(event)
}

// SendToComponent sends an event to the component.
func (c *Component) SendToComponent(component manager.Component, event manager.EventMsg) {
	c.app.Manager.SendTo(component, event)
}
