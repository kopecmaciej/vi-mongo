package core

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

// NewBaseView is a constructor for the BaseView struct.
func NewBaseView(identifier string) *BaseView {
	return &BaseView{
		identifier: manager.ViewIdentifier(identifier),
	}
}

// Init is a function that is called when the view is initialized.
// If custom initialization is needed, this function should be overriden.
func (c *BaseView) Init(app *App) error {
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
func (c *BaseView) Enable() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.enabled = true
	c.App.GetManager().PushView(c.identifier)
}

// Disable unsets the enabled flag.
func (c *BaseView) Disable() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.enabled = false
	c.App.GetManager().PopView()
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

// GetAfterInitFunc returns the optional function that will be run at the end of the Init function.
func (c *BaseView) GetAfterInitFunc() func() error {
	return c.afterInitFunc
}

// GetIdentifier returns the identifier of the view.
func (c *BaseView) GetIdentifier() manager.ViewIdentifier {
	return c.identifier
}

// Subscribe subscribes to the view events.
func (c *BaseView) Subscribe() {
	c.Listener = c.App.GetManager().Subscribe(c.identifier)
}

// Broadcast sends an event to all listeners.
func (c *BaseView) BroadcastEvent(event manager.EventMsg) {
	c.App.GetManager().Broadcast(event)
}

// SendToView sends an event to the view.
func (c *BaseView) SendToView(view manager.ViewIdentifier, event manager.EventMsg) {
	c.App.GetManager().SendTo(view, event)
}
