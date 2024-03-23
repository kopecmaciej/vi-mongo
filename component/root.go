package component

import (
	"github.com/rs/zerolog/log"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/config"
	"github.com/kopecmaciej/mongui/manager"
	"github.com/kopecmaciej/mongui/mongo"
	"github.com/kopecmaciej/tview"
)

const (
	RootComponent manager.Component = "Root"
)

// Root is a component that manages visaibility of other components
type Root struct {
	*Component
	*tview.Pages

	mainFlex  *tview.Flex
	innerFlex *tview.Flex
	style     *config.RootStyle
	connector *Connector
	header    *Header
	databases *Databases
	content   *Content
}

func NewRoot() *Root {
	r := &Root{
		Component: NewComponent(RootComponent),
		Pages:     tview.NewPages(),
		mainFlex:  tview.NewFlex(),
		innerFlex: tview.NewFlex(),
		connector: NewConnector(),
		header:    NewHeader(),
		databases: NewDatabases(),
		content:   NewContent(),
	}

	return r
}

// Init initializes root component and
// initializes all subcomponents asynchronically
func (r *Root) Init() error {
	r.setStyles()
	r.setKeybindings()

	if err := r.connector.Init(r.app); err != nil {
		return err
	}
	if r.app.Config.ShowWelcomePage {
		if err := r.renderWelcome(); err != nil {
			return err
		}
	} else {
		currConn := r.app.Config.GetCurrentConnection()
		if currConn == nil || r.app.Config.ShowConnectionPage {
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
	currConn := r.app.Config.GetCurrentConnection()
	if r.app.Dao != nil && *r.app.Dao.Config == *currConn {
		return nil
	} else {
		// TODO: find the correct way to refresh those components
		r.content = NewContent()
		r.databases = NewDatabases()
		r.header = NewHeader()
		client := mongo.NewClient(currConn)
		err := client.Connect()
		if err != nil {
			return err
		}
		err = client.Ping()
		if err != nil {
			return err
		}

		r.app.Dao = mongo.NewDao(client.Client, client.Config)

		if err := r.initSubcomponents(); err != nil {
			return err
		}

		r.databases.dbTree.NodeSelectFunc = r.content.RenderContent

		r.render()

		return nil
	}
}

func (r *Root) initSubcomponents() error {
	runWithDraw := func(f func(app *App) error) {
		r.app.QueueUpdateDraw(func() {
			if err := f(r.app); err != nil {
				log.Error().Err(err).Msg("Error initializing components")
			}
		})
	}

	go runWithDraw(r.header.Init)
	go runWithDraw(r.databases.Init)
	go runWithDraw(r.content.Init)

	return nil
}

func (r *Root) setStyles() {
	r.style = &r.app.Styles.Root
	r.Pages.SetBackgroundColor(r.style.BackgroundColor.Color())
	r.mainFlex.SetBackgroundColor(r.style.BackgroundColor.Color())
	r.innerFlex.SetBackgroundColor(r.style.BackgroundColor.Color())
}

// setKeybindings sets a key binding for the root Component
func (r *Root) setKeybindings() {
	k := r.app.Keys
	r.app.Root.mainFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case k.Contains(k.Root.FocusNext, event.Name()):
			focus := r.app.GetFocus()
			if focus == r.databases.dbTree {
				r.app.SetFocus(r.content.Table)
			} else {
				r.app.SetFocus(r.databases.dbTree)
			}
			return nil
		case k.Contains(k.Root.HideDatabases, event.Name()):
			if _, ok := r.mainFlex.GetItem(0).(*Databases); ok {
				r.mainFlex.RemoveItem(r.databases)
				r.app.SetFocus(r.content.Table)
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

// render renders the root component and all subcomponents
func (r *Root) render() error {
	r.mainFlex.Clear()
	r.innerFlex.Clear()

	r.innerFlex.SetBackgroundColor(r.style.BackgroundColor.Color())
	r.innerFlex.SetDirection(tview.FlexRow)

	r.mainFlex.AddItem(r.databases, 30, 0, true)
	r.mainFlex.AddItem(r.innerFlex, 0, 7, false)
	r.innerFlex.AddItem(r.header, 4, 0, false)
	r.innerFlex.AddItem(r.content, 0, 7, true)

	r.AddPage(r.GetIdentifier(), r.mainFlex, true, true)
	r.app.SetFocus(r.mainFlex)

	return nil
}

// renderWelcome renders welcome component
func (r *Root) renderWelcome() error {
	welcome := NewWelcome()

	if err := welcome.Init(r.app); err != nil {
		return err
	}
	welcome.SetOnSubmitFunc(func() {
		r.RemovePage(welcome.GetIdentifier())
		err := r.renderConnector()
		if err != nil {
			r.AddPage(welcome.GetIdentifier(), welcome, true, true)
			ShowErrorModal(r, "Error while connecting to the database", err)
			return
		}
	})
	r.AddPage(welcome.GetIdentifier(), welcome, true, true)
	return nil
}

// renderConnector renders connector component
func (r *Root) renderConnector() error {
	r.connector.SetOnSubmitFunc(func() {
		r.RemovePage(r.connector.GetIdentifier())
		err := r.renderMainView()
		if err != nil {
			r.AddPage(r.connector.GetIdentifier(), r.connector, true, true)
			ShowErrorModal(r, "Error while connecting to the database", err)
		}
	})

	r.AddPage(r.connector.GetIdentifier(), r.connector, true, true)
	return nil
}

// AddPage is a wrapper for tview.Pages.AddPage
func (r *Root) AddPage(component manager.Component, page tview.Primitive, resize, visable bool) *tview.Pages {
	if r.Pages.HasPage(string(component)) && r.app.Manager.CurrentComponent() == component {
		return r.Pages
	}
	r.app.Manager.PushComponent(component)
	return r.Pages.AddPage(string(component), page, resize, visable)
}

// RemovePage is a wrapper for tview.Pages.RemovePage
func (r *Root) RemovePage(component manager.Component) *tview.Pages {
	r.app.Manager.PopComponent()
	return r.Pages.RemovePage(string(component))
}

// HasPage is a wrapper for tview.Pages.HasPage
func (r *Root) HasPage(component manager.Component) bool {
	return r.Pages.HasPage(string(component))
}
