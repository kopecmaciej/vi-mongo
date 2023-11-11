package main

import (
	"context"
	"mongo-ui/mongo"
	"mongo-ui/pages"

	"github.com/rivo/tview"
)

const (
	AppCtxKey = "app"
)

type App struct {
	*tview.Application

	root *pages.Root
}

func NewApp() App {
	mongoConfig := mongo.NewConfig()
	client := mongo.NewClient(mongoConfig)
	client.Connect()
	mongoDao := mongo.NewDao(client.Client)

	app := App{
		Application: tview.NewApplication(),
		root:        pages.NewRoot(mongoDao),
	}

	return app
}

func (a *App) Init() error {
  root := a.root.Init(a.setContext(context.Background()))
  focus := a.GetFocus()
  a.SetRoot(root, true).EnableMouse(true)
  a.SetFocus(focus)
	return a.Run()
}

// set app into context and use it in other packages
func (a *App) setContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, AppCtxKey, a.Application)
}
