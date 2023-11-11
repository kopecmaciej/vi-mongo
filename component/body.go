package component

import (
	"context"
	"mongo-ui/mongo"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Body struct {
	*tview.Flex

	app     *tview.Application
	sideBar *SideBar
	content *Content
	table   *tview.Table
	mongo   *mongo.Dao
}

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

	b.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			focus := b.app.GetFocus()
			if focus == b.sideBar.TreeView {
				b.app.SetFocus(b.content.Table)
			} else {
				b.app.SetFocus(b.sideBar.TreeView)
			}
		case tcell.KeyCtrlS:
			// Toggle the side bar
			if _, ok := b.Flex.GetItem(0).(*SideBar); ok {
				b.Flex.RemoveItem(b.sideBar)
			} else {
				b.Flex.RemoveItem(b.content)
				b.Flex.AddItem(b.sideBar, 0, 1, false) // Add side bar back at position 0
				b.Flex.AddItem(b.content, 0, 4, false) // Add content back at position 1
			}
		case tcell.KeyCtrlD:
			// Toggle the content
			if b.Flex.GetItemCount() > 1 && b.Flex.GetItem(1) == b.content {
				b.Flex.RemoveItem(b.content)
			} else {
				b.Flex.AddItem(b.content, 1, 4, false) // Add content back at position 1
			}
		}

		return event
	})

	return b.Flex
}

func (b *Body) SetStyle() {
	b.Flex.SetBackgroundColor(tcell.ColorDefault)
}
