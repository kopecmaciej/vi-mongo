package component

import (
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
	a.setKeybindings(help)

	return a.Run()
}

func (a *App) setKeybindings(help *Help) {
	a.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == '?' {
			help.Render()
			a.Root.Pages.AddPage("help", help, true, true)
			return nil
		}

		return event
	})
}
