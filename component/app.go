package component

import (
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

		Dao     *mongo.Dao
		Manager *manager.ComponentManager
		Root    *Root
		Styles  *config.Styles
		Config  *config.Config
		Keys    *config.KeyBindings
	}
)

func NewApp(appConfig *config.Config) App {
	styles, err := config.LoadStyles()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load styles")
	}
	keyBindings := config.NewKeyBindings()

	app := App{
		Application: tview.NewApplication(),
		Root:        NewRoot(),
		Manager:     manager.NewComponentManager(),
		Styles:      styles,
		Config:      appConfig,
		Keys:        &keyBindings,
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
		switch {
		case a.Keys.Contains(a.Keys.Global.ToggleFullScreenHelp, event.Name()):
			if a.Root.HasPage(string(HelpComponent)) {
				a.Root.RemovePage(HelpComponent)
				return nil
			}
			err := help.Render()
			if err != nil {
				return event
			}
			a.Root.AddPage(HelpComponent, help, true, true)
			return nil
		case a.Keys.Contains(a.Keys.Global.ToggleHelpBar, event.Name()):
			if a.Root.innerFlex.HasItem(help) {
				a.Root.innerFlex.RemoveItem(help)
				return nil
			}
			err := help.Render()
			if err != nil {
				return event
			}
			a.Root.innerFlex.AddItem(help, 10, 0, false)
			return nil
		}
		return event
	})
}
