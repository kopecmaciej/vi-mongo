package main

import (
	"context"
	"mongo-ui/mongo"
	"mongo-ui/pages"

	"github.com/rivo/tview"
)

const (
	AppCtxKey  = "app"
	RootCtxKey = "root"
)

type App struct {
	*tview.Application

	Root *pages.Root
}

func NewApp() App {
	mongoConfig := mongo.NewConfig()
	client := mongo.NewClient(mongoConfig)
	client.Connect()
	mongoDao := mongo.NewDao(client.Client)

	app := App{
		Application: tview.NewApplication(),
		Root:        pages.NewRoot(mongoDao),
	}

	return app
}

func (a *App) Init() error {
	root := a.Root.Init(a.setContext(context.Background()))
	focus := a.GetFocus()
	a.SetRoot(root, true).EnableMouse(true)
	a.SetFocus(focus)
	return a.Run()
}

// set app into context and use it in other packages
func (a *App) setContext(ctx context.Context) context.Context {
	ctxWithValue := context.WithValue(ctx, AppCtxKey, a.Application)
	ctxWithValue = context.WithValue(ctxWithValue, RootCtxKey, a.Root.Pages)
	return ctxWithValue
}

