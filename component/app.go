package component

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/config"
	"github.com/kopecmaciej/mongui/manager"
	"github.com/kopecmaciej/mongui/mongo"
	"github.com/kopecmaciej/tview"
	"github.com/rs/zerolog/log"
)

type (
	// App is a main application struct
	App struct {
		*tview.Application

		Dao            *mongo.Dao
		Manager        *manager.ComponentManager
		Root           *Root
		FullScreenHelp *Help
		FooterHelp     *Help
		Styles         *config.Styles
		Config         *config.Config
		Keys           *config.KeyBindings
	}
)

func NewApp(appConfig *config.Config) App {
	styles, err := config.LoadStyles()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load styles")
	}
	keyBindings, err := config.LoadKeybindings()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load keybindings")
	}

	app := App{
		Application:    tview.NewApplication(),
		Root:           NewRoot(),
		FullScreenHelp: NewHelp(),
		FooterHelp:     NewHelp(),
		Manager:        manager.NewComponentManager(),
		Styles:         styles,
		Config:         appConfig,
		Keys:           keyBindings,
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

	err := a.FullScreenHelp.Init(a)
	if err != nil {
		return err
	}
	err = a.FooterHelp.Init(a)
	if err != nil {
		return err
	}
	a.setKeybindings()

	return a.Run()
}

func (a *App) setKeybindings() {
	a.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// TODO: This is temporary solution
		if strings.Contains(string(a.Manager.CurrentComponent()), "Input") {
			return event
		}
		switch {
		case a.Keys.Contains(a.Keys.Global.ToggleFullScreenHelp, event.Name()):
			if a.Root.HasPage(HelpComponent) {
				a.Root.RemovePage(HelpComponent)
				return nil
			}
			err := a.FullScreenHelp.Render(true)
			if err != nil {
				return event
			}
			a.Root.AddPage(HelpComponent, a.FullScreenHelp, true, true)
			return nil
		case a.Keys.Contains(a.Keys.Global.ToggleHelpBarFooter, event.Name()):
			if a.Root.innerFlex.HasItem(a.FooterHelp) {
				a.Root.innerFlex.RemoveItem(a.FooterHelp)
				return nil
			}
			err := a.FooterHelp.Render(false)
			if err != nil {
				return event
			}
			a.Root.innerFlex.AddItem(a.FooterHelp, 10, 0, false)
			return nil
		}
		return event
	})
}
