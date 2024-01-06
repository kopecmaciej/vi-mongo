package component

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/config"
	"github.com/kopecmaciej/mongui/manager"
	"github.com/kopecmaciej/mongui/mongo"
	"github.com/rivo/tview"
)

const (
	RootComponent manager.Component = "Root"
)

type Root struct {
	*tview.Pages
	*tview.Flex

	mongoDao *mongo.Dao
	app      *App
  style    *config.Root
	header   *Header
	sideBar  *SideBar
	content  *Content
	manager  *manager.ComponentManager
}

func NewRoot(mongoDao *mongo.Dao) *Root {
	root := &Root{
		Pages:    tview.NewPages(),
		Flex:     tview.NewFlex(),
		mongoDao: mongoDao,
		header:   NewHeader(mongoDao),
		sideBar:  NewSideBar(mongoDao),
		content:  NewContent(mongoDao),
	}

	return root
}
func (r *Root) Init(ctx context.Context) error {
	app, err := GetApp(ctx)
	if err != nil {
		return err
	}
	r.app = app
	r.manager = r.app.ComponentManager

  r.style = &r.app.Styles.Root

	r.Pages.SetBackgroundColor(r.style.BackgroundColor.Get())
	r.Flex.SetBackgroundColor(r.style.BackgroundColor.Get())

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

	r.sideBar.DBTree.NodeSelectFunc = r.content.RenderContent

	r.render(ctx)
	r.registerKeyHandlers(ctx)
	r.setShortcuts(ctx)

	r.AddPage(RootComponent, r.Flex, true, true)

	return nil
}

func (r *Root) render(ctx context.Context) error {
	body := tview.NewFlex()
	body.SetBackgroundColor(r.style.BackgroundColor.Get())
	body.SetDirection(tview.FlexRow)

	r.Flex.AddItem(r.sideBar, 30, 0, false)
	r.Flex.AddItem(body, 0, 7, true)
	body.AddItem(r.header.Table, 0, 1, false)
	body.AddItem(r.content.Flex, 0, 7, true)

	r.app.SetFocus(r.sideBar.Flex)

	return nil
}

func (r *Root) registerKeyHandlers(ctx context.Context) {
	rootManager := r.manager.SetKeyHandlerForComponent(RootComponent)
	rootManager(tcell.KeyCtrlS, func() {
		if _, ok := r.Flex.GetItem(0).(*SideBar); ok {
			r.Flex.RemoveItem(r.sideBar)
			r.app.SetFocus(r.content.Table)
		} else {
			r.Flex.Clear()
			r.render(ctx)
		}
	})
	rootManager(tcell.KeyTab, func() {
		focus := r.app.GetFocus()
		if focus == r.sideBar.DBTree {
			r.app.SetFocus(r.content.Table)
		} else {
			r.app.SetFocus(r.sideBar.DBTree)
		}
	})
}

func (r *Root) setShortcuts(ctx context.Context) {
	r.app.Root.Pages.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		r.manager.HandleKey(event.Key())
		return event
	})
}

func (r *Root) AddPage(component manager.Component, page tview.Primitive, resize, visable bool) {
	r.Pages.AddPage(string(component), page, resize, visable)
	r.manager.PushComponent(component)
}

func (r *Root) RemovePage(component manager.Component) {
	r.Pages.RemovePage(string(component))
	r.manager.PopComponent()
}
