package component

import (
	"context"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/config"
	"github.com/kopecmaciej/mongui/manager"
	"github.com/kopecmaciej/mongui/mongo"
	"github.com/rivo/tview"
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
	ctx := context.Background()
	a.Root.app = a
	if err := a.Root.Init(); err != nil {
		return err
	}
	a.SetRoot(a.Root.Pages, true).EnableMouse(true)

	help := NewHelp()
	err := help.Init(a)
	if err != nil {
		return err
	}
	a.setKeybindings(ctx, help)

	return a.Run()
}

func (a *App) setKeybindings(ctx context.Context, help *Help) {
	m := a.Manager.SetKeyHandlerForComponent(manager.GlobalComponent)
	m(tcell.KeyRune, '?', "Toggle help", func(e *tcell.EventKey) *tcell.EventKey {
		if a.Root.HasPage(string(HelpComponent)) {
			a.Root.RemovePage(HelpComponent)
			return nil
		}
		help.Render()
		a.Root.AddPage(HelpComponent, help, true, true)
		return nil
	})
	m(tcell.KeyCtrlC, 0, "Quit the application", func(e *tcell.EventKey) *tcell.EventKey {
		if a.Dao != nil {
			a.Dao.ForceClose(ctx)
		}
		a.Stop()
		return nil
	})

	a.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return a.Manager.HandleKeyEvent(event, manager.GlobalComponent)
	})
}
