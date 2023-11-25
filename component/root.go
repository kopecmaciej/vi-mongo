package component

import (
	"context"
	"mongo-ui/mongo"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type flexStack struct {
	label string
	fixed int
	prop  int
}

var (
	rootStack = []flexStack{
		{"header", 0, 1},
		{"searchBar", 1, 0},
		{"sideBar", 30, 0},
		{"content", 0, 1},
	}
)

type Root struct {
	*tview.Pages
	*tview.Flex

	mongoDao *mongo.Dao
	app      *App
	header   *Header
	sideBar  *SideBar
	content  *Content
}

func NewRoot(mongoDao *mongo.Dao) *Root {
	return &Root{
		Pages:    tview.NewPages(),
		Flex:     tview.NewFlex(),
		header:   NewHeader(mongoDao),
		sideBar:  NewSideBar(mongoDao),
		content:  NewContent(mongoDao),
		mongoDao: mongoDao,
	}
}
func (r *Root) Init(ctx context.Context) error {
	r.app = GetApp(ctx)

	r.Pages.SetBackgroundColor(tcell.ColorDefault)
	r.Flex.SetBackgroundColor(tcell.ColorDefault)

	if err := r.header.Init(); err != nil {
		return err
	}
	if err := r.sideBar.Init(ctx); err != nil {
		return err
	}
	if err := r.content.Init(ctx); err != nil {
		return err
	}

	r.render(ctx)
  r.SetShortcuts(ctx)

	r.AddPage("main", r.Flex, true, true)

	return nil
}

func (r *Root) render(ctx context.Context) error {
	body := tview.NewFlex()
	body.SetBackgroundColor(tcell.ColorDefault)
	body.SetDirection(tview.FlexRow)

	r.Flex.AddItem(r.sideBar.TreeView, 30, 0, false)
	r.Flex.AddItem(body, 0, 7, true)
	body.AddItem(r.header.Table, 0, 1, false)
	body.AddItem(r.content.Table, 0, 7, true)

	err := r.sideBar.RenderTree(ctx, r.content.RenderContent)
	if err != nil {
		return err
	}

	r.app.SetFocus(r.sideBar.TreeView)

	return nil
}

func (r *Root) SetShortcuts(ctx context.Context) {
	r.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			focus := r.app.GetFocus()
			if focus == r.sideBar.TreeView {
				r.app.SetFocus(r.content.Table)
			} else {
				r.app.SetFocus(r.sideBar.TreeView)
			}
		case tcell.KeyCtrlS:
			if _, ok := r.Flex.GetItem(0).(*SideBar); ok {
				r.Flex.RemoveItem(r.sideBar)
				r.app.SetFocus(r.content.Table)
			} else {
				r.Flex.Clear()
				r.render(ctx)
			}
		case tcell.KeyCtrlD:
			if r.Flex.GetItemCount() > 1 && r.Flex.GetItem(1) == r.content {
				r.Flex.RemoveItem(r.content)
				r.app.SetFocus(r.sideBar.TreeView)
			} else {
				r.Flex.Clear()
				r.render(ctx)
			}
		}

		return event
	})
}
