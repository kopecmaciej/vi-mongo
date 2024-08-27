package page

import (
	"github.com/rs/zerolog/log"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/kopecmaciej/vi-mongo/internal/tui/component"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/tui/modal"

	"github.com/kopecmaciej/tview"
)

const (
	RootView = "Root"
)

// Root is a view that manages visaibility of other views
type Root struct {
	*core.BaseElement

	mainFlex  *tview.Flex
	innerFlex *tview.Flex
	style     *config.RootStyle
	connector *Connector
	header    *component.Header
	databases *component.Databases
	content   *component.Content
}

func NewRoot() *Root {
	r := &Root{
		BaseElement: core.NewBaseElement(RootView),
		mainFlex:    tview.NewFlex(),
		innerFlex:   tview.NewFlex(),
		connector:   NewConnector(),
		header:      component.NewHeader(),
		databases:   component.NewDatabases(),
		content:     component.NewContent(),
	}

	return r
}

// Init initializes root view and
// initializes all subviews asynchronically
func (r *Root) Init() error {
	r.setStyles()
	r.setKeybindings()

	if err := r.connector.Init(r.App); err != nil {
		return err
	}
	if r.App.GetConfig().ShowWelcomePage {
		if err := r.renderWelcome(); err != nil {
			return err
		}
	} else {
		currConn := r.App.GetConfig().GetCurrentConnection()
		if currConn == nil || r.App.GetConfig().ShowConnectionPage {
			if err := r.renderConnector(); err != nil {
				return err
			}
		} else {
			if err := r.renderMainView(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Root) renderMainView() error {
	currConn := r.App.GetConfig().GetCurrentConnection()
	if r.App.Dao != nil && *r.App.Dao.Config == *currConn {
		return nil
	} else {
		// TODO: find the correct way to refresh those views
		r.content = component.NewContent()
		r.databases = component.NewDatabases()
		r.header = component.NewHeader()
		client := mongo.NewClient(currConn)
		err := client.Connect()
		if err != nil {
			return err
		}
		err = client.Ping()
		if err != nil {
			return err
		}

		r.App.Dao = mongo.NewDao(client.Client, client.Config)

		if err := r.initSubviews(); err != nil {
			return err
		}

		r.databases.SetSelectFunc(r.content.HandleDatabaseSelection)

		r.render()

		return nil
	}
}

func (r *Root) initSubviews() error {
	runWithDraw := func(f func(app *core.App) error) {
		r.App.QueueUpdateDraw(func() {
			if err := f(r.App); err != nil {
				log.Error().Err(err).Msg("Error initializing views")
			}
		})
	}

	go runWithDraw(r.header.Init)
	go runWithDraw(r.databases.Init)
	go runWithDraw(r.content.Init)

	return nil
}

func (r *Root) setStyles() {
	r.style = &r.App.GetStyles().Root
	r.App.Pages.SetBackgroundColor(r.style.BackgroundColor.Color())
	r.mainFlex.SetBackgroundColor(r.style.BackgroundColor.Color())
	r.innerFlex.SetBackgroundColor(r.style.BackgroundColor.Color())
}

// setKeybindings sets a key binding for the root View
func (r *Root) setKeybindings() {
	k := r.App.GetKeys()
	r.mainFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case k.Contains(k.Root.FocusNext, event.Name()):
			r.databases.ToggleFocus()
			return nil
		case k.Contains(k.Root.HideDatabases, event.Name()):
			if _, ok := r.mainFlex.GetItem(0).(*component.Databases); ok {
				r.mainFlex.RemoveItem(r.databases)
				r.App.SetFocus(r.content.Table)
			} else {
				r.mainFlex.Clear()
				r.render()
			}
			return nil
		case k.Contains(k.Root.OpenConnector, event.Name()):
			r.renderConnector()
			return nil
		}
		return event
	})
}

// render renders the root view and all subviews
func (r *Root) render() error {
	r.mainFlex.Clear()
	r.innerFlex.Clear()

	r.innerFlex.SetBackgroundColor(r.style.BackgroundColor.Color())
	r.innerFlex.SetDirection(tview.FlexRow)

	r.mainFlex.AddItem(r.databases, 30, 0, true)
	r.mainFlex.AddItem(r.innerFlex, 0, 7, false)
	r.innerFlex.AddItem(r.header, 4, 0, false)
	r.innerFlex.AddItem(r.content, 0, 7, true)

	r.App.Pages.AddPage(r.GetIdentifier(), r.mainFlex, true, true)
	r.App.SetFocus(r.mainFlex)

	return nil
}

// renderWelcome renders welcome view
func (r *Root) renderWelcome() error {
	welcome := NewWelcome()

	if err := welcome.Init(r.App); err != nil {
		return err
	}
	welcome.SetOnSubmitFunc(func() {
		r.App.Pages.RemovePage(welcome.GetIdentifier())
		err := r.renderConnector()
		if err != nil {
			r.App.Pages.AddPage(welcome.GetIdentifier(), welcome, true, true)
			modal.ShowError(r.App.Pages, "Error while connecting to the database", err)
			return
		}
	})
	r.App.Pages.AddPage(welcome.GetIdentifier(), welcome, true, true)
	return nil
}

// renderConnector renders connector view
func (r *Root) renderConnector() error {
	r.connector.SetOnSubmitFunc(func() {
		r.App.Pages.RemovePage(r.connector.GetIdentifier())
		err := r.renderMainView()
		if err != nil {
			r.App.Pages.AddPage(r.connector.GetIdentifier(), r.connector, true, true)
			modal.ShowError(r.App.Pages, "Error while connecting to the database", err)
		}
	})

	r.App.Pages.AddPage(r.connector.GetIdentifier(), r.connector, true, true)
	return nil
}
