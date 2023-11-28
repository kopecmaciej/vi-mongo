package component

import (
	"context"
	"mongo-ui/mongo"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	appCtxKey = "app"
)

type App struct {
	*tview.Application

	Root *Root
}

func NewApp() App {
	client := mongo.NewClient()
	client.Connect()
	mongoDao := mongo.NewDao(client.Client, client.Config)

	loadStyles()

	app := App{
		Application: tview.NewApplication(),
		Root:        NewRoot(mongoDao),
	}

	return app
}

func (a *App) Init() error {
	ctx := LoadApp(context.Background(), a)
	err := a.Root.Init(ctx)
	if err != nil {
		return err
	}
	focus := a.GetFocus()
	a.SetRoot(a.Root.Pages, true).EnableMouse(true)
	a.SetFocus(focus)
	return a.Run()
}

func loadStyles() {
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
	tview.Styles.ContrastBackgroundColor = tcell.ColorDefault
	tview.Styles.MoreContrastBackgroundColor = tcell.ColorDefault
	tview.Styles.PrimaryTextColor = tcell.ColorDefault
	tview.Styles.SecondaryTextColor = tcell.ColorYellow
	tview.Styles.TertiaryTextColor = tcell.ColorBlue
	tview.Styles.InverseTextColor = tcell.ColorBlue
	tview.Styles.ContrastSecondaryTextColor = tcell.ColorYellow
	tview.Styles.BorderColor = tcell.ColorBlue
	tview.Styles.TitleColor = tcell.ColorDefault
	tview.Styles.GraphicsColor = tcell.ColorBlue
}

func GetApp(ctx context.Context) *App {
	app, ok := ctx.Value(appCtxKey).(*App)
	if !ok {
		panic("App not found in context")
	}
	return app
}

func LoadApp(ctx context.Context, app *App) context.Context {
	return context.WithValue(ctx, appCtxKey, app)
}
