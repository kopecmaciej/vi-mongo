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
		dao           *mongo.Dao
		manager       *manager.ElementManager
		styles        *config.Styles
		config        *config.Config
		keys          *config.KeyBindings
		previousFocus tview.Primitive
	}
)

func NewApp(appConfig *config.Config) *App {
	styles, err := config.LoadStyles(appConfig.CurrentStyle)
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
		manager:     manager.NewElementManager(),
		styles:      styles,
		config:      appConfig,
		keys:        keyBindings,
	}

	app.Pages = NewPages(app.manager, app)

	return app
}

func (a *App) NextStyle() error {
	nextStyle, err := a.styles.PickNextStyle(a.config.CurrentStyle)
	if err != nil {
		return err
	}

	a.config.CurrentStyle = nextStyle
	err = a.config.UpdateConfig()
	if err != nil {
		return err
	}

	a.styles, err = config.LoadStyles(a.config.CurrentStyle)
	if err != nil {
		return err
	}
	a.styles.LoadMainStyles()

	// Broadcast style change event
	a.manager.Broadcast(manager.EventMsg{
		Message: manager.Message{
			Type: manager.StyleChanged,
			Data: a.styles,
		},
	})

	return nil
}

func (a *App) SetPreviousFocus() {
	a.previousFocus = a.GetFocus()
}

func (a *App) SetFocus(p tview.Primitive) {
	a.previousFocus = a.GetFocus()
	a.Application.SetFocus(p)
	a.FocusChanged(p)
}

func (a *App) GiveBackFocus() {
	if a.previousFocus != nil {
		a.SetFocus(a.previousFocus)
		a.previousFocus = nil
	}
}

// FocusChanged is a callback that is called when the focus is changed
// it is used to update the keys
func (a *App) FocusChanged(p tview.Primitive) {
	msg := manager.EventMsg{
		Message: manager.Message{
			Type: manager.FocusChanged,
			Data: p.GetIdentifier(),
		},
	}
	a.manager.Broadcast(msg)
}

func (a *App) GetDao() *mongo.Dao {
	return a.dao
}

func (a *App) SetDao(dao *mongo.Dao) {
	a.dao = dao
}

func (a *App) GetManager() *manager.ElementManager {
	return a.manager
}

func (a *App) GetKeys() *config.KeyBindings {
	return a.keys
}

func (a *App) GetStyles() *config.Styles {
	return a.styles
}

func (a *App) GetConfig() *config.Config {
	return a.config
}
