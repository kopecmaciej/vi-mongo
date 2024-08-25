package tui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/internal/config"
	"github.com/kopecmaciej/mongui/internal/tui/core"
)

type (
	// App extends the core.App struct
	App struct {
		*core.App
		Root           *Root
		FullScreenHelp *Help
	}
)

func NewApp(appConfig *config.Config) *App {
	coreApp := core.NewApp(appConfig)

	app := &App{
		App:            coreApp,
		Root:           NewRoot(),
		FullScreenHelp: NewHelp(),
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
			if a.Pages.HasPage(HelpView) {
				a.Pages.RemovePage(HelpView)
				return nil
			}
			err := a.FullScreenHelp.Render(true)
			if err != nil {
				return event
			}
			a.Pages.AddPage(HelpView, a.FullScreenHelp, true, true)
			return nil
		case a.Keys.Contains(a.Keys.Global.ToggleHelpBarFooter, event.Name()):

			if strings.Contains(string(a.Manager.CurrentView()), "Input") {
				return event
			}
			return nil
		}
		return event
	})
}
