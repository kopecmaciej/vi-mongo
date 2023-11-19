package body

import (
	"context"
	"mongo-ui/mongo"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type flexTuple struct {
	label string
	fixed int
	prop  int
}

type Body struct {
	*tview.Flex

	app     *tview.Application
	sideBar *SideBar
	content *Content
	table   *tview.Table
	mongo   *mongo.Dao
}

var (
	flexTuples = []flexTuple{
		{"sideBar", 0, 1},
		{"content", 0, 4},
	}
)

func NewBody(mongo *mongo.Dao) *Body {
	return &Body{
		Flex:    tview.NewFlex(),
		sideBar: NewSideBar(mongo),
		content: NewContent(mongo),
		table:   tview.NewTable(),
		mongo:   mongo,
	}
}

func (b *Body) Init(ctx context.Context) *tview.Flex {
	b.app = ctx.Value("app").(*tview.Application)

	b.SetStyle()

	b.sideBar.Init(ctx)
	b.Flex.AddItem(b.sideBar, 0, 1, false)

	b.content.Init(ctx)

	b.app.SetFocus(b.sideBar)

	b.sideBar.RenderTree(ctx, b.content.RenderDocuments)
	b.Flex.AddItem(b.content, 0, 4, false)

	b.SetShortcuts()

	return b.Flex
}

func (b *Body) SetStyle() {
	b.Flex.SetBackgroundColor(tcell.ColorDefault)
}

func (b *Body) SetShortcuts() {
	b.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			focus := b.app.GetFocus()
			if focus == b.sideBar.TreeView {
				b.app.SetFocus(b.content.Table)
			} else {
				b.app.SetFocus(b.sideBar.TreeView)
			}
		case tcell.KeyCtrlS:
			if _, ok := b.Flex.GetItem(0).(*SideBar); ok {
				b.Flex.RemoveItem(b.sideBar)
				b.app.SetFocus(b.content.Table)
			} else {
				b.Flex.Clear()
				b.render()
			}
		case tcell.KeyCtrlD:
			if b.Flex.GetItemCount() > 1 && b.Flex.GetItem(1) == b.content {
				b.Flex.RemoveItem(b.content)
				b.app.SetFocus(b.sideBar.TreeView)
			} else {
				b.Flex.Clear()
				b.render()
			}
		}

		return event
	})
}

func (b *Body) render() {
	for _, tuple := range flexTuples {
		b.Flex.AddItem(b.GetPrimitiveByLabel(tuple.label), tuple.fixed, tuple.prop, false)
	}
}

func (b *Body) GetPrimitiveByLabel(label string) tview.Primitive {
	switch label {
	case "sideBar":
		return b.sideBar
	case "content":
		return b.content
	}

	return nil
}
