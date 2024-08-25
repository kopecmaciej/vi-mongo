package view

import (
	"sync"

	"github.com/kopecmaciej/mongui/internal/manager"
	"github.com/kopecmaciej/mongui/internal/mongo"
)

// BaseView is a base struct for all views.
// It contains all the basic fields and functions that are used by all views.
// It also implements the BaseView interface.
type BaseView struct {
	// enabled is a flag that indicates if the view is enabled.
	enabled bool

	// name is the name of the view.
	// It's used mainly for managing avaliable keybindings.
	identifier manager.ViewIdentifier
	// app is a pointer to the main app.
	// It's used for accessing app focus, root page etc.
	app *App

	// dao is a pointer to the mongo dao.
	dao *mongo.Dao

	// initFunc is a function that is called when the view is initialized.
	// It's main purpose is to run all the initialization functions of the subviews.
	afterInitFunc func() error

	// listener is a channel that is used to receive events from the app.
	listener chan manager.EventMsg

	// mutex is a mutex that is used to synchronize the view.
	mutex sync.Mutex
}

// NewBaseView is a constructor for the BaseView struct.
func NewBaseView(identifier string) *BaseView {
	return &BaseView{
		identifier: manager.ViewIdentifier(identifier),
	}
}

// Init is a function that is called when the view is initialized.
// If custom initialization is needed, this function should be overriden.
func (c *BaseView) Init(app *App) error {
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
func (c *BaseView) Enable() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.enabled = true
	c.app.Manager.PushView(c.identifier)
}

// Disable unsets the enabled flag.
func (c *BaseView) Disable() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.enabled = false
	c.app.Manager.PopView()
}

// Toggle toggles the enabled flag.
func (c *BaseView) Toggle() {
	if c.IsEnabled() {
		c.Disable()
	} else {
		c.Enable()
	}
}

// IsEnabled returns the enabled flag.
func (c *BaseView) IsEnabled() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.enabled
}

// SetAfterInitFunc sets the optional function that will be run at the end of the Init function.
func (c *BaseView) SetAfterInitFunc(afterInitFunc func() error) {
	c.afterInitFunc = afterInitFunc
}

// GetIdentifier returns the identifier of the view.
func (c *BaseView) GetIdentifier() manager.ViewIdentifier {
	return c.identifier
}

// Subscribe subscribes to the view events.
func (c *BaseView) Subscribe() {
	c.listener = c.app.Manager.Subscribe(c.identifier)
}

// Broadcast sends an event to all listeners.
func (c *BaseView) BroadcastEvent(event manager.EventMsg) {
	c.app.Manager.Broadcast(event)
}

// SendToView sends an event to the view.
func (c *BaseView) SendToView(view manager.ViewIdentifier, event manager.EventMsg) {
	c.app.Manager.SendTo(view, event)
}
