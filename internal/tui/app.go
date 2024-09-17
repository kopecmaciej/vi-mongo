package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/tui/modal"
	"github.com/kopecmaciej/vi-mongo/internal/tui/page"
)

type (
	// App extends the core.App struct
	App struct {
		*core.App

		// initial pages
		connector *page.Connector
		main      *page.Main
		help      *page.Help
	}
)

func NewApp(appConfig *config.Config) *App {
	coreApp := core.NewApp(appConfig)

	app := &App{
		App: coreApp,

		connector: page.NewConnector(),
		main:      page.NewMain(),
		help:      page.NewHelp(),
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

	if err := a.connector.Init(a.App); err != nil {
		return err
	}

	return nil
}

func (a *App) Run() error {
	return a.Application.Run()
}

func (a *App) setKeybindings() {
	a.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// TODO: This is temporary solution
		switch {
		case a.GetKeys().Contains(a.GetKeys().Global.OpenConnector, event.Name()):
			a.renderConnector()
			return nil
		case a.GetKeys().Contains(a.GetKeys().Global.NextStyle, event.Name()):
			err := a.NextStyle()
			if err != nil {
				modal.ShowError(a.Pages, "Failed to load next style", err)
			}
			return nil
		case a.GetKeys().Contains(a.GetKeys().Global.ToggleFullScreenHelp, event.Name()):
			if a.Pages.HasPage(page.HelpPage) {
				a.Pages.RemovePage(page.HelpPage)
				return nil
			}
			err := a.help.Render()
			if err != nil {
				return event
			}
			a.Pages.AddPage(page.HelpPage, a.help, true, true)
			return nil
		}
		return event
	})
}

func (a *App) connectToMongo() error {
	currConn := a.App.GetConfig().GetCurrentConnection()
	if a.GetDao() != nil && *a.GetDao().Config == *currConn {
		return nil
	}

	client := mongo.NewClient(currConn)
	if err := client.Connect(); err != nil {
		return err
	}
	if err := client.Ping(); err != nil {
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
		if err := a.renderWelcome(); err != nil {
			modal.ShowError(a.Pages, "Error while rendering welcome page", err)
		}
	case a.App.GetConfig().GetCurrentConnection() == nil, a.App.GetConfig().ShowConnectionPage:
		if err := a.renderConnector(); err != nil {
			modal.ShowError(a.Pages, "Error while rendering connector", err)
		}
	default:
		// we need to init main view after connection is established
		// as it depends on the dao
		if err := a.initAndRenderMain(); err != nil {
			modal.ShowError(a.Pages, "Error while initializing main view", err)
			return
		}
	}
}

// initAndRenderMain initializes and renders the main page
// methods are combined as we need to establish connection first
func (a *App) initAndRenderMain() error {
	if err := a.connectToMongo(); err != nil {
		return err
	}

	// if main view is already initialized, we just update dao
	if a.main.App != nil || a.main.Dao != nil {
		a.main.UpdateDao(a.GetDao())
	} else {
		if err := a.main.Init(a.App); err != nil {
			return err
		}
	}

	a.main.Render()
	a.Pages.AddPage(a.main.GetIdentifier(), a.main, true, true)
	return nil
}

// renderConnector renders the connector page
func (a *App) renderConnector() error {
	a.connector.SetOnSubmitFunc(func() {
		a.Pages.RemovePage(a.connector.GetIdentifier())
		err := a.initAndRenderMain()
		if err != nil {
			a.Pages.AddPage(a.connector.GetIdentifier(), a.connector, true, true)
			modal.ShowError(a.App.Pages, "Error while connecting to the database", err)
		}
	})

	a.connector.Render()
	a.Pages.AddPage(a.connector.GetIdentifier(), a.connector, true, true)
	return nil
}

// renderWelcome renders the welcome page
// it's initialized inside render function
// as it's probalby won't be used very often
func (a *App) renderWelcome() error {
	welcome := page.NewWelcome()
	if err := welcome.Init(a.App); err != nil {
		return err
	}
	welcome.SetOnSubmitFunc(func() {
		a.Pages.RemovePage(welcome.GetIdentifier())
		err := a.renderConnector()
		if err != nil {
			a.Pages.AddPage(welcome.GetIdentifier(), welcome, true, true)
			modal.ShowError(a.Pages, "Error while rendering connector page", err)
		}
	})
	welcome.Render()
	a.Pages.AddPage(welcome.GetIdentifier(), welcome, true, true)
	return nil
}
