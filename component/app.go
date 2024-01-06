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

type App struct {
	*tview.Application

	ComponentManager *manager.ComponentManager
	Root             *Root
	Styles           *config.Styles
}

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
	}

	return app
}

func (a *App) Init() error {
	ctx := LoadApp(context.Background(), a)
	err := a.Root.Init(ctx)
	if err != nil {
		return err
	}
	a.SetRoot(a.Root.Pages, true).EnableMouse(true)
	return a.Run()
}

func GetApp(ctx context.Context) (*App, error) {
	app, ok := ctx.Value(appCtxKey).(*App)
	if !ok {
		log.Error().Msg("error getting app from context")
		return nil, fmt.Errorf("error getting app from context")
	}
	return app, nil
}

func LoadApp(ctx context.Context, app *App) context.Context {
	return context.WithValue(ctx, appCtxKey, app)
}
