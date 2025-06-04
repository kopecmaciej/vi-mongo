package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/tui/modal"
	"github.com/kopecmaciej/vi-mongo/internal/tui/page"
	"github.com/kopecmaciej/vi-mongo/internal/util"
	"github.com/rs/zerolog/log"
)

type (
	// App extends the core.App struct
	App struct {
		*core.App

		// initial pages
		connection *page.Connection
		main       *page.Main
		help       *page.Help
	}
)

func NewApp(appConfig *config.Config) *App {
	coreApp := core.NewApp(appConfig)

	app := &App{
		App: coreApp,

		connection: page.NewConnection(),
		main:       page.NewMain(),
		help:       page.NewHelp(),
	}

	return app
}

// Init initializes app
func (a *App) Init() error {
	a.SetRoot(a.Pages, true).EnableMouse(true)

	err := a.help.Init(a.App)
	if err != nil {
		return err
	}
	a.setKeybindings()

	a.connection.Init(a.App)
	return nil
}

func (a *App) Run() error {
	return a.Application.Run()
}

func (a *App) setKeybindings() {
	a.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if a.shouldHandleRune(event) {
			return event
		}

		switch {
		case a.GetKeys().Contains(a.GetKeys().Global.CloseApp, event.Name()):
			a.Stop()
			return nil
		case a.GetKeys().Contains(a.GetKeys().Global.OpenConnection, event.Name()):
			a.renderConnection()
			return nil
		case a.GetKeys().Contains(a.GetKeys().Global.ShowStyleModal, event.Name()):
			a.ShowStyleChangeModal()
			return nil
		case a.GetKeys().Contains(a.GetKeys().Global.ToggleFullScreenHelp, event.Name()):
			if a.Pages.HasPage(page.HelpPageId) {
				a.Pages.RemovePage(page.HelpPageId)
				return nil
			}
			a.help.Render()
			a.Pages.AddPage(page.HelpPageId, a.help, true, true)
			return nil
		}
		return event
	})
}

// shouldHandleRune checks if the rune event should be passed through to input fields
// it's needed if we want to assign rune to global keybindings and this rune should be captured
// by input fields
func (a *App) shouldHandleRune(event *tcell.EventKey) bool {
	if !strings.HasPrefix(event.Name(), "Rune") {
		return false
	}

	focus := a.GetFocus()
	identifier := string(focus.GetIdentifier())

	if strings.Contains(identifier, "Bar") || strings.Contains(identifier, "Input") {
		return true
	}

	_, isInputField := focus.(*tview.InputField)
	_, isCustomInputField := focus.(*core.InputField)
	_, isFormItem := focus.(tview.FormItem)

	return isInputField || isCustomInputField || isFormItem
}

func (a *App) connectToMongo() error {
	currConn := a.App.GetConfig().GetCurrentConnection()
	if a.GetDao() != nil && *a.GetDao().Config == *currConn {
		return nil
	}

	client := mongo.NewClient(currConn)
	if err := client.Connect(); err != nil {
		log.Error().Err(err).Msg("Failed to connect to mongo")
		return err
	}
	if err := client.Ping(); err != nil {
		log.Error().Err(err).Msg("Failed to ping to mongo")
		return err
	}
	a.SetDao(mongo.NewDao(client.Client, client.Config))
	return nil
}

// Render is the main render function
// it renders the page based on the config
func (a *App) Render() {
	switch {
	case a.App.GetConfig().ShowWelcomePage:
		a.renderWelcome()
	case a.App.GetConfig().GetCurrentConnection() == nil, a.App.GetConfig().ShowConnectionPage:
		a.renderConnection()
	default:
		// we need to init main view after connection is established
		// as it depends on the dao
		a.initAndRenderMain()
	}
}

// initAndRenderMain initializes and renders the main page
// methods are combined as we need to establish connection first
func (a *App) initAndRenderMain() {
	if err := a.connectToMongo(); err != nil {
		a.renderConnection()
		if _, ok := err.(*util.EncryptionError); ok {
			modal.ShowError(a.Pages, "Encryption error occurred", err)
		} else {
			modal.ShowError(a.Pages, "Error while connecting to mongodb", err)
		}
		return
	}

	// if main view is already initialized, we just update dao
	if a.main.App != nil || a.main.Dao != nil {
		a.main.UpdateDao(a.GetDao())
	} else {
		if err := a.main.Init(a.App); err != nil {
			log.Fatal().Err(err).Msg("Error while initializing main view")
			os.Exit(1)
		}
	}

	a.main.Render()
	a.Pages.AddPage(a.main.GetIdentifier(), a.main, true, true)

	if jumpInto := a.GetConfig().JumpInto; jumpInto != "" {
		if err := a.handleDirectNavigation(jumpInto); err != nil {
			modal.ShowError(a.Pages, "Direct navigation failed", err)
		}
	}
}

// renderConnection renders the connection page
func (a *App) renderConnection() {
	a.connection.SetOnSubmitFunc(func() {
		a.Pages.RemovePage(a.connection.GetIdentifier())
		a.initAndRenderMain()
	})

	a.Pages.AddPage(a.connection.GetIdentifier(), a.connection, true, true)
	a.connection.Render()
}

// renderWelcome renders the welcome page
// it's initialized inside render function
// as it's probalby won't be used very often
func (a *App) renderWelcome() {
	welcome := page.NewWelcome()
	if err := welcome.Init(a.App); err != nil {
		a.Pages.AddPage(welcome.GetIdentifier(), welcome, true, true)
		modal.ShowError(a.Pages, "Error while rendering welcome page", err)
		return
	}
	welcome.SetOnSubmitFunc(func() {
		a.Pages.RemovePage(welcome.GetIdentifier())
		a.renderConnection()
	})
	a.Pages.AddPage(welcome.GetIdentifier(), welcome, true, true)
	welcome.Render()
}

func (a *App) ShowStyleChangeModal() {
	styleChangeModal := modal.NewStyleChangeModal()
	if err := styleChangeModal.Init(a.App); err != nil {
		modal.ShowError(a.Pages, "Error while initializing style change modal", err)
	}
	styleChangeModal.Render()
	styleChangeModal.SetApplyStyle(func(styleName string) error {
		return a.SetStyle(styleName)
	})
}

func (a *App) handleDirectNavigation(directNav string) error {
	parts := strings.Split(directNav, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid format: expected db-name/collection-name, got %s", directNav)
	}

	dbName := strings.TrimSpace(parts[0])
	collName := strings.TrimSpace(parts[1])

	if dbName == "" || collName == "" {
		return fmt.Errorf("database name and collection name cannot be empty")
	}

	return a.main.NavigateToDbCollection(dbName, collName)
}
