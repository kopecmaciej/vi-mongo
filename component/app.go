package component

import (
	"context"
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/config"
	"github.com/kopecmaciej/mongui/manager"
	"github.com/kopecmaciej/mongui/mongo"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
)

const (
	appCtxKey = "app"
)

type (
	// EventMsg is a wrapper for tcell.EventKey that also contains
	// the sender of the event
	EventMsg struct {
		*tcell.EventKey
		Sender manager.Component
	}

	// Broadcaster is a struct that contains listeners for events
	Broadcaster struct {
		listeners map[manager.Component]chan EventMsg
		mutex     sync.Mutex
	}

	// Listener
	Listener struct {
		Component manager.Component
		Channel   chan EventMsg
	}

	// App is a main application struct
	App struct {
		*tview.Application

		ComponentManager *manager.ComponentManager
		Root             *Root
		Styles           *config.Styles
		Broadcaster      *Broadcaster
	}
)

func NewApp(appConfig *config.MonguiConfig) App {
	client := mongo.NewClient(&appConfig.Mongo)
	client.Connect()
	mongoDao := mongo.NewDao(client.Client, client.Config)

	styles := config.NewStyles()

	app := App{
		Application:      tview.NewApplication(),
		Root:             NewRoot(mongoDao),
		ComponentManager: manager.NewComponentManager(),
		Styles:           styles,
		Broadcaster:      &Broadcaster{listeners: make(map[manager.Component]chan EventMsg)},
	}

	return app
}

// Init initializes app
func (a *App) Init() error {
	ctx := LoadApp(context.Background(), a)
	err := a.Root.Init(ctx)
	if err != nil {
		return err
	}
	a.SetRoot(a.Root.Pages, true).EnableMouse(true)
	return a.Run()
}

// GetApp gets app from context
func GetApp(ctx context.Context) (*App, error) {
	app, ok := ctx.Value(appCtxKey).(*App)
	if !ok {
		log.Error().Msg("error getting app from context")
		return nil, fmt.Errorf("error getting app from context")
	}
	return app, nil
}

// LoadApp loads app into context
func LoadApp(ctx context.Context, app *App) context.Context {
	return context.WithValue(ctx, appCtxKey, app)
}

// Subscribe subscribes to events from a specific component
func (b *Broadcaster) Subscribe(component manager.Component) chan EventMsg {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	listener := make(chan EventMsg)
	b.listeners[component] = listener
	return listener
}

// Unsubscribe unsubscribes from events from a specific component
func (b *Broadcaster) Unsubscribe(component manager.Component, listener chan EventMsg) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	delete(b.listeners, component)
}

// Broadcast sends an event to a specific component
func (b *Broadcaster) Broadcast(event EventMsg) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	for _, listener := range b.listeners {
		listener <- event
	}
}
