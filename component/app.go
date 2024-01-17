package component

import (
	"context"
	"fmt"

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
	// App is a main application struct
	App struct {
		*tview.Application

		Dao     *mongo.Dao
		Manager *manager.ComponentManager
		Root    *Root
		Styles  *config.Styles
		Config  *config.Config
	}
)

func NewApp(appConfig *config.Config) App {
	styles := config.NewStyles()

	app := App{
		Application: tview.NewApplication(),
		Root:        NewRoot(),
		Manager:     manager.NewComponentManager(),
		Styles:      styles,
		Config:      appConfig,
	}

	return app
}

// Init initializes app
func (a *App) Init() error {
	if err := a.Root.Init(a); err != nil {
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
