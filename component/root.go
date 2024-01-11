package component

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/config"
	"github.com/kopecmaciej/mongui/manager"
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

	r.SetAfterInitFunc(r.init)

	return r
}

// Init initializes root component and
// initializes all subcomponents asynchronically
func (r *Root) init(ctx context.Context) error {
	r.setStyles()

	var e error

	go func() {
		r.app.QueueUpdateDraw(func() {
			if err := r.header.Init(ctx); err != nil {
				e = err
				return
			}
		})
	}()
	go func() {
		r.app.QueueUpdateDraw(func() {
			if err := r.sideBar.Init(ctx); err != nil {
				e = err
				return
			}
		})
	}()
	go func() {
		r.app.QueueUpdateDraw(func() {
			if err := r.content.Init(ctx); err != nil {
				e = err
				return
			}
		})
	}()

	if e != nil {
		log.Error().Err(e).Msg("Error initializing root")
		return e
	}

	r.sideBar.dbTree.NodeSelectFunc = r.content.RenderContent

	r.Render(ctx)
	r.registerKeyHandlers(ctx)
	r.shortcuts(ctx)

	r.AddPage(r.GetIdentifier(), r.flex, true, true)

	return nil
}

func (r *Root) setStyles() {
	r.style = &r.app.Styles.Root
	r.Pages.SetBackgroundColor(r.style.BackgroundColor.Color())
	r.flex.SetBackgroundColor(r.style.BackgroundColor.Color())
}

func (r *Root) shortcuts(ctx context.Context) {
	r.app.Root.Pages.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		r.manager.HandleKey(event.Key())
		return event
	})
}

// Render renders the root component and all subcomponents
func (r *Root) Render(ctx context.Context) error {
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

// registerKeyHandlers registers global key handlers
// for every component
func (r *Root) registerKeyHandlers(ctx context.Context) {
	rootManager := r.manager.SetKeyHandlerForComponent(r.GetIdentifier())
	rootManager(tcell.KeyCtrlS, func() {
		if _, ok := r.flex.GetItem(0).(*SideBar); ok {
			r.flex.RemoveItem(r.sideBar)
			r.app.SetFocus(r.content.Table)
		} else {
			r.flex.Clear()
			r.Render(ctx)
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
	r.manager.PushComponent(component)
	return r.Pages.AddPage(string(component), page, resize, visable)
}

// RemovePage is a wrapper for tview.Pages.RemovePage
func (r *Root) RemovePage(component manager.Component) *tview.Pages {
	r.manager.PopComponent()
	return r.Pages.RemovePage(string(component))
}
