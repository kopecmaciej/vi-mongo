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

	flex    *tview.Flex
	style   *config.Root
	header  *Header
	sideBar *SideBar
	content *Content
}

func NewRoot() *Root {
	r := &Root{
		Component: NewComponent("Root"),
		Pages:     tview.NewPages(),
		flex:      tview.NewFlex(),
		header:    NewHeader(),
		sideBar:   NewSideBar(),
		content:   NewContent(),
	}

	return r
}

// Init initializes root component and
// initializes all subcomponents asynchronically
func (r *Root) Init(a *App) error {
	r.app = a
	r.setStyles()

	currConn := a.Config.GetCurrentConnection()
	if currConn == nil {
		err := r.renderConnector()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Root) callback() {
	r.RemovePage("Connector")
	currConn := r.app.Config.GetCurrentConnection()
	client := mongo.NewClient(currConn)
	client.Connect()
	r.app.Dao = mongo.NewDao(client.Client, client.Config)

	if err := r.initSubcomponents(); err != nil {
		log.Error().Err(err).Msg("Error initializing root")
	}

	r.sideBar.dbTree.NodeSelectFunc = r.content.RenderContent

	r.render()
	r.registerKeyHandlers()
	r.shortcuts()

	r.AddPage(r.GetIdentifier(), r.flex, true, true)
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

func (r *Root) shortcuts() {
	r.app.Root.Pages.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		r.app.Manager.HandleKey(event.Key())
		return event
	})
}

// render renders the root component and all subcomponents
func (r *Root) render() error {
	body := tview.NewFlex()
	body.SetBackgroundColor(r.style.BackgroundColor.Color())
	body.SetDirection(tview.FlexRow)

	r.flex.AddItem(r.sideBar, 30, 0, false)
	r.flex.AddItem(body, 0, 7, true)
	body.AddItem(r.header, 0, 1, false)
	body.AddItem(r.content, 0, 7, true)

	r.app.SetFocus(r.sideBar)

	return nil
}

// renderConnector renders connector component
func (r *Root) renderConnector() error {
	connector := NewConnector()
	if err := connector.Init(r.app); err != nil {
		log.Error().Err(err).Msg("Error initializing root")
		return err
	}
	connector.SetCallback(r.callback)
	r.AddPage(connector.GetIdentifier(), connector, true, true)
	return nil
}

// registerKeyHandlers registers global key handlers
// for every component
func (r *Root) registerKeyHandlers() {
	rootManager := r.app.Manager.SetKeyHandlerForComponent(r.GetIdentifier())
	rootManager(tcell.KeyCtrlS, func() {
		if _, ok := r.flex.GetItem(0).(*SideBar); ok {
			r.flex.RemoveItem(r.sideBar)
			r.app.SetFocus(r.content.Table)
		} else {
			r.flex.Clear()
			r.render()
		}
	})
	rootManager(tcell.KeyTab, func() {
		focus := r.app.GetFocus()
		if focus == r.sideBar.dbTree {
			r.app.SetFocus(r.content.Table)
		} else {
			r.app.SetFocus(r.sideBar.dbTree)
		}
	})
}

// AddPage is a wrapper for tview.Pages.AddPage
func (r *Root) AddPage(component manager.Component, page tview.Primitive, resize, visable bool) *tview.Pages {
	r.app.Manager.PushComponent(component)
	return r.Pages.AddPage(string(component), page, resize, visable)
}

// RemovePage is a wrapper for tview.Pages.RemovePage
func (r *Root) RemovePage(component manager.Component) *tview.Pages {
	r.app.Manager.PopComponent()
	return r.Pages.RemovePage(string(component))
}
