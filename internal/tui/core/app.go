package core

import (
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
	"github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/rs/zerolog/log"
)

type (
	// App is a main application struct
	App struct {
		*tview.Application

		Pages         *Pages
		Dao           *mongo.Dao
		Manager       *manager.ElementManager
		Styles        *config.Styles
		Config        *config.Config
		Keys          *config.KeyBindings
		PreviousFocus tview.Primitive
	}
)

func NewApp(appConfig *config.Config) *App {
	styles, err := config.LoadStyles()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load styles")
	}
	styles.LoadMainStyles()
	keyBindings, err := config.LoadKeybindings()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load keybindings")
	}

	app := &App{
		Application: tview.NewApplication(),
		Manager:     manager.NewElementManager(),
		Styles:      styles,
		Config:      appConfig,
		Keys:        keyBindings,
	}

	app.Pages = NewPages(app.Manager, app)

	return app
}

func (a *App) SetPreviousFocus() {
	a.PreviousFocus = a.GetFocus()
}

func (a *App) SetFocus(p tview.Primitive) {
	a.PreviousFocus = a.GetFocus()
	a.Application.SetFocus(p)
	a.FocusChanged(p)
}

func (a *App) GiveBackFocus() {
	if a.PreviousFocus != nil {
		a.SetFocus(a.PreviousFocus)
		a.PreviousFocus = nil
	}
}

// FocusChanged is a callback that is called when the focus is changed
// it is used to update the keys
func (a *App) FocusChanged(p tview.Primitive) {
	a.Manager.Broadcast(manager.EventMsg{
		Message: manager.Message{
			Type: manager.FocusChanged,
			Data: p.GetIdentifier(),
		},
	})
}

// GetDao implements model.AppInterface
func (a *App) GetDao() *mongo.Dao {
	return a.Dao
}

// GetManager implements model.AppInterface
func (a *App) GetManager() *manager.ElementManager {
	return a.Manager
}

// GetKeys implements models.App
func (a *App) GetKeys() *config.KeyBindings {
	return a.Keys
}

// GetStyles implements models.App
func (a *App) GetStyles() *config.Styles {
	return a.Styles
}

// GetConfig implements models.App
func (a *App) GetConfig() *config.Config {
	return a.Config
}

// GetPages implements models.App
func (a *App) GetPages() *tview.Pages {
	return a.Pages.Pages
}
