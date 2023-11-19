package body

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mongo-ui/mongo"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Content struct {
	*tview.Table

	app     *tview.Application
	dao     *mongo.Dao
	docView *DocView
	label   string
	state   state
}

type state struct {
	page  int64
	limit int64
	db    string
	coll  string
}

func NewContent(dao *mongo.Dao) *Content {
	state := state{
		page:  0,
		limit: 50,
	}
	table := tview.NewTable()
	return &Content{
		Table: table,

		dao:     dao,
		docView: NewDocView(dao),
		label:   "content",
		state:   state,
	}
}

func (c *Content) Init(ctx context.Context) error {
	c.app = ctx.Value("app").(*tview.Application)
	c.setStyle()
	c.SetShortcuts(ctx)

	c.docView.Init(ctx, c)

	return nil
}

func (c *Content) RenderDocuments(db, coll string) error {
	ctx := context.Background()
	c.Clear()
	c.app.SetFocus(c)

	c.state.db = db
	c.state.coll = coll

	filters := map[string]interface{}{}

	documents, count, err := c.dao.ListDocuments(ctx, db, coll, filters, c.state.page, c.state.limit)
	if err != nil {
		return err
	}

	if len(documents) == 0 {
		c.SetCell(1, 1, tview.NewTableCell("No documents found").SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft))
	}

	headerInfo := fmt.Sprintf("Documents: %d, Page: %d, Limit: %d", count, c.state.page, c.state.limit)
	c.SetCell(0, 0, tview.NewTableCell(headerInfo).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft))

	for i, d := range documents {
		jsonBytes, err := json.Marshal(d)
		if err != nil {
			log.Printf("Error marshaling JSON: %v", err)
			continue
		}

		c.SetCell(i+1, 0, tview.NewTableCell(string(jsonBytes)).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft))
	}
	return nil
}

func (c *Content) refresh() {
	c.RenderDocuments(c.state.db, c.state.coll)
}

func (c *Content) setStyle() {
	c.SetBackgroundColor(tcell.ColorDefault)
	c.SetBorder(true)
	c.SetTitle("Content")
	c.SetTitleAlign(tview.AlignLeft)
	c.SetTitleColor(tcell.ColorSteelBlue)
	c.SetBorderColor(tcell.ColorSteelBlue)
	c.SetBorderPadding(0, 0, 3, 3)
	c.SetFixed(1, 1)
	c.SetSelectable(true, false)
}

func (c *Content) SetShortcuts(ctx context.Context) {
	c.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'v':
			c.docView.DocView(ctx, c.state.db, c.state.coll, c.GetCell(c.GetSelection()).Text)
		case 'e':
			c.docView.DocEdit(ctx, c.state.db, c.state.coll, c.GetCell(c.GetSelection()).Text, c.refresh)
		}
		switch event.Key() {
		case tcell.KeyCtrlN:
			c.goToNextMongoPage(ctx)
		case tcell.KeyCtrlP:
			c.goToPrevMongoPage(ctx)
		case tcell.KeyEnter:
			c.docView.DocView(ctx, c.state.db, c.state.coll, c.GetCell(c.GetSelection()).Text)
		}

		return event
	})
}

func (c *Content) goToNextMongoPage(ctx context.Context) {
	c.state.page += c.state.limit
	c.RenderDocuments(c.state.db, c.state.coll)
}

func (c *Content) goToPrevMongoPage(ctx context.Context) {
	c.state.page -= c.state.limit
	c.RenderDocuments(c.state.db, c.state.coll)
}
