package component

import (
	"github.com/rs/zerolog/log"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/config"
	"github.com/kopecmaciej/mongui/manager"
	"github.com/kopecmaciej/mongui/mongo"
	"github.com/rivo/tview"
)

// Root is a component that manages visaibility of other components
type Root struct {
	*Component
	*tview.Pages

	flex      *tview.Flex
	style     *config.RootStyle
	connector *Connector
	header    *Header
	sideBar   *SideBar
	content   *Content
}

func NewRoot() *Root {
	r := &Root{
		Component: NewComponent("Root"),
		Pages:     tview.NewPages(),
		flex:      tview.NewFlex(),
		connector: NewConnector(),
		header:    NewHeader(),
		sideBar:   NewSideBar(),
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

	return nil
}

func (r *Root) renderMainView() error {
	currConn := r.app.Config.GetCurrentConnection()
	client := mongo.NewClient(currConn)
	err := client.Connect()
	if err != nil {
		ShowErrorModal(r, "Error connecting to database", err)
		return err
	}
	err = client.Ping()
	if err != nil {
		ShowErrorModal(r, "Error pinging database", err)
		return err
	}

	r.app.Dao = mongo.NewDao(client.Client, client.Config)

	if err := r.initSubcomponents(); err != nil {
		return err
	}

	r.sideBar.dbTree.NodeSelectFunc = r.content.RenderContent

	r.render()

	return nil
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
	go runWithDraw(r.sideBar.Init)
	go runWithDraw(r.content.Init)

	return nil
}

func (r *Root) setStyles() {
	r.style = &r.app.Styles.Root
	r.Pages.SetBackgroundColor(r.style.BackgroundColor.Color())
	r.flex.SetBackgroundColor(r.style.BackgroundColor.Color())
}

// setKeybindings sets a key binding for the root Component
func (r *Root) setKeybindings() {
	k := r.app.Keys
	r.app.Root.flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case k.Contains(k.Root.FocusNext, event.Name()):
			focus := r.app.GetFocus()
			if focus == r.sideBar.dbTree {
				r.app.SetFocus(r.content.Table)
			} else {
				r.app.SetFocus(r.sideBar.dbTree)
			}
			return nil
		case k.Contains(k.Root.HideSidebar, event.Name()):
			if _, ok := r.flex.GetItem(0).(*SideBar); ok {
				r.flex.RemoveItem(r.sideBar)
				r.app.SetFocus(r.content.Table)
			} else {
				r.flex.Clear()
				r.render()
			}
			return nil
		case k.Contains(k.Root.OpenConnector, event.Name()):
			r.flex.Clear()
			r.renderConnector()
			return nil
		}
		return event
	})
}

// render renders the root component and all subcomponents
func (r *Root) render() error {
	body := tview.NewFlex()
	body.SetBackgroundColor(r.style.BackgroundColor.Color())
	body.SetDirection(tview.FlexRow)

	r.flex.AddItem(r.sideBar, 30, 0, true)
	r.flex.AddItem(body, 0, 7, false)
	body.AddItem(r.header, 0, 1, false)
	body.AddItem(r.content, 0, 7, true)

	r.AddPage(r.GetIdentifier(), r.flex, true, true)

	return nil
}

// renderConnector renders connector component
func (r *Root) renderConnector() error {
	r.connector.SetCallback(func() {
		err := r.renderMainView()
		if err != nil {
			ShowErrorModal(r, "Error connecting to database", err)
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
