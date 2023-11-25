package component

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

	app     *App
	dao     *mongo.Dao
	docView *DocViewer
	label   string
	state   contentState
}

type contentState struct {
	page  int64
	limit int64
	db    string
	coll  string
}

func NewContent(dao *mongo.Dao) *Content {
	state := contentState{
		page:  0,
		limit: 50,
	}
	table := tview.NewTable()

	return &Content{
		Table:   table,
		dao:     dao,
		docView: NewDocView(dao),
		label:   "content",
		state:   state,
	}
}

func (c *Content) Init(ctx context.Context) error {
	c.app = GetApp(ctx)
	c.setStyle()
	c.SetShortcuts(ctx)

	c.docView.Init(ctx, c)

	return nil
}

func (c *Content) setStyle() {
	c.SetBackgroundColor(tcell.NewRGBColor(0, 10, 19))
	c.SetBorder(true)
	c.SetTitle(" Content ")
	c.SetTitleAlign(tview.AlignLeft)
	c.SetTitleColor(tcell.ColorSteelBlue)
	c.SetBorderColor(tcell.ColorSteelBlue)
	c.SetFixed(1, 1)
	c.SetSelectable(true, false)
}

func (c *Content) listDocuments(db, coll string) ([]string, int64, error) {
	ctx := context.Background()
	c.state.db = db
	c.state.coll = coll

	filters := map[string]interface{}{}
	documents, count, err := c.dao.ListDocuments(ctx, db, coll, filters, c.state.page, c.state.limit)
	if err != nil {
		return nil, 0, err
	}

	if len(documents) == 0 {
		return nil, 0, nil
	}

	var docs []string
	for _, d := range documents {
		jsonBytes, err := json.Marshal(d)
		if err != nil {
			log.Printf("Error marshaling JSON: %v", err)
			continue
		}
		docs = append(docs, string(jsonBytes))
	}

	return docs, count, nil
}

func (c *Content) RenderContent(db, coll string) error {
	c.Clear()
	c.app.SetFocus(c)

	c.state.db = db
	c.state.coll = coll

	documents, count, err := c.listDocuments(db, coll)
	if err != nil {
		return err
	}

	if count == 0 {
		noDocCell := tview.NewTableCell("No documents found").
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft).
			SetSelectable(false)

		c.SetCell(1, 1, noDocCell)
		return nil
	}

	headerInfo := fmt.Sprintf("Documents: %d, Page: %d, Limit: %d", count, c.state.page, c.state.limit)
	headerCell := tview.NewTableCell(headerInfo).
		SetTextColor(tcell.ColorWhite).
		SetAlign(tview.AlignLeft).
		SetSelectable(false)

	c.SetCell(0, 0, headerCell)

	for i, d := range documents {
		dataCell := tview.NewTableCell(d).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft)

		c.SetCell(i+1, 0, dataCell)
	}
	return nil
}

func (c *Content) refresh() {
	c.RenderContent(c.state.db, c.state.coll)
}

func (c *Content) Filter(ctx context.Context, db, coll, filter string, refresh func()) {
	c.state.page = 0
	c.state.limit = 50
	c.state.db = db
	c.state.coll = coll

	filters := map[string]interface{}{}

	documents, count, err := c.dao.ListDocuments(ctx, db, coll, filters, c.state.page, c.state.limit)
	if err != nil {
		return
	}

	if len(documents) == 0 {
		noDocCell := tview.NewTableCell("No documents found").
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft).
			SetSelectable(false)

		c.SetCell(1, 1, noDocCell)
		return
	}

	headerInfo := fmt.Sprintf("Documents: %d, Page: %d, Limit: %d", count, c.state.page, c.state.limit)
	headerCell := tview.NewTableCell(headerInfo).
		SetTextColor(tcell.ColorWhite).
		SetAlign(tview.AlignLeft).
		SetSelectable(false)

	c.SetCell(0, 0, headerCell)
}

func (c *Content) SetShortcuts(ctx context.Context) {
	c.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'v':
			c.docView.DocViewer(ctx, c.state.db, c.state.coll, c.GetCell(c.GetSelection()).Text)
		case 'e':
			c.docView.DocEdit(ctx, c.state.db, c.state.coll, c.GetCell(c.GetSelection()).Text, c.refresh)
		case '/':
			c.Filter(ctx, c.state.db, c.state.coll, c.GetCell(c.GetSelection()).Text, c.refresh)
		}
		switch event.Key() {
		case tcell.KeyCtrlN:
			c.goToNextMongoPage(ctx)
		case tcell.KeyCtrlP:
			c.goToPrevMongoPage(ctx)
		case tcell.KeyEnter:
			c.docView.DocViewer(ctx, c.state.db, c.state.coll, c.GetCell(c.GetSelection()).Text)
		}

		return event
	})
}

func (c *Content) goToNextMongoPage(ctx context.Context) {
	c.state.page += c.state.limit
	c.RenderContent(c.state.db, c.state.coll)
}

func (c *Content) goToPrevMongoPage(ctx context.Context) {
	c.state.page -= c.state.limit
	c.RenderContent(c.state.db, c.state.coll)
}
