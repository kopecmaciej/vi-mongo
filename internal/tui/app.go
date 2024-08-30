package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/tui/page"
)

type (
	// App extends the core.App struct
	App struct {
		*core.App
		Root           *page.Root
		FullScreenHelp *page.Help
	}
)

func NewApp(appConfig *config.Config) *App {
	coreApp := core.NewApp(appConfig)

	app := &App{
		App:            coreApp,
		Root:           page.NewRoot(),
		FullScreenHelp: page.NewHelp(),
	}

	return app
}

// Init initializes app
func (a *App) Init() error {
	a.Root.App = a.App
	if err := a.Root.Init(); err != nil {
		return err
	}
	a.SetRoot(a.Pages, true).EnableMouse(true)

	err := a.FullScreenHelp.Init(a.App)
	if err != nil {
		return err
	}
	a.setKeybindings()

	return a.Run()
}

func (a *App) setKeybindings() {
	a.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// TODO: This is temporary solution
		switch {
		case a.Keys.Contains(a.Keys.Global.ToggleFullScreenHelp, event.Name()):
			if a.Pages.HasPage(page.HelpPage) {
				a.Pages.RemovePage(page.HelpPage)
				return nil
			}
			err := a.FullScreenHelp.Render()
			if err != nil {
				return event
			}
			a.Pages.AddPage(page.HelpPage, a.FullScreenHelp, true, true)
			return nil
		}
		return event
	})
}
