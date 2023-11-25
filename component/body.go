package component

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

	app     *App
	sideBar *SideBar
	content *Content
	mongo   *mongo.Dao
}

var (
	flexTuples = []flexTuple{
		{"sideBar", 30, 0},
		{"content", 0, 1},
	}
)

func NewBody(mongo *mongo.Dao) *Body {
	return &Body{
		Flex:    tview.NewFlex(),
		sideBar: NewSideBar(mongo),
		content: NewContent(mongo),
		mongo:   mongo,
	}
}

func (b *Body) Init(ctx context.Context) error {
	b.app = GetApp(ctx)

	b.SetStyle()

	err := b.sideBar.Init(ctx)
	if err != nil {
		return err
	}
	err = b.content.Init(ctx)
	if err != nil {
		return err
	}

	b.render()

	b.app.SetFocus(b.sideBar)
	err = b.sideBar.RenderTree(ctx, b.content.RenderContent)
	if err != nil {
		return err
	}

	b.SetShortcuts()

	return nil
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
		b.Flex.AddItem(b.getPrimitiveByLabel(tuple.label), tuple.fixed, tuple.prop, false)
	}
}

func (b *Body) getPrimitiveByLabel(label string) tview.Primitive {
	switch label {
	case "sideBar":
		return b.sideBar
	case "content":
		return b.content
	}

	return nil
}
