package view

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/internal/config"
	"github.com/kopecmaciej/mongui/internal/manager"
	"github.com/kopecmaciej/mongui/internal/mongo"
	"github.com/kopecmaciej/mongui/internal/views/core"
	"github.com/kopecmaciej/tview"
	"github.com/rs/zerolog/log"
)

type (
	// TODO: remove from views package
	// App is a main application struct
	App struct {
		*tview.Application

		Pages          *core.Pages
		Dao            *mongo.Dao
		Manager        *manager.ViewManager
		Root           *Root
		FullScreenHelp *Help
		Styles         *config.Styles
		Config         *config.Config
		Keys           *config.KeyBindings
		PreviousFocus  tview.Primitive
	}
)

func NewApp(appConfig *config.Config) *App {
	styles, err := config.LoadStyles()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load styles")
	}
	keyBindings, err := config.LoadKeybindings()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load keybindings")
	}

	app := &App{
		Application:    tview.NewApplication(),
		Root:           NewRoot(),
		FullScreenHelp: NewHelp(),
		Manager:        manager.NewViewManager(),
		Styles:         styles,
		Config:         appConfig,
		Keys:           keyBindings,
	}

	app.Pages = core.NewPages(app.Manager, app)

	return app
}

// Init initializes app
func (a *App) Init() error {
	a.Root.app = a
	if err := a.Root.Init(); err != nil {
		return err
	}
	a.SetRoot(a.Pages, true).EnableMouse(true)

	err := a.FullScreenHelp.Init(a)
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

func (a *App) SetPreviousFocus() {
	a.PreviousFocus = a.GetFocus()
}

func (a *App) SetFocus(p tview.Primitive) {
	a.PreviousFocus = a.GetFocus()
	a.Application.SetFocus(p)
}

func (a *App) GiveBackFocus() {
	if a.PreviousFocus != nil {
		a.SetFocus(a.PreviousFocus)
		a.PreviousFocus = nil
	}
}
