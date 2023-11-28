package component

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mongo-ui/mongo"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.mongodb.org/mongo-driver/bson"
)

type Content struct {
	*tview.Flex

	Table    *tview.Table
	app      *App
	dao      *mongo.Dao
	queryBar *InputBar
	docView  *DocViewer
	label    string
	mutex    sync.Mutex
	state    contentState
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

	flex := tview.NewFlex()
	return &Content{
		Table:    tview.NewTable(),
		Flex:     flex,
		queryBar: NewInputBar("Query"),
		dao:      dao,
		docView:  NewDocView(dao),
		mutex:    sync.Mutex{},
		label:    "content",
		state:    state,
	}
}

func (c *Content) Init(ctx context.Context) error {
	c.app = GetApp(ctx)
	c.setStyle()
	c.setShortcuts(ctx)

	if err := c.docView.Init(ctx, c.Flex); err != nil {
		return err
	}
	if err := c.queryBar.Init(ctx); err != nil {
		return err
	}
	c.queryBar.AutocompleteOn = true

	c.render(ctx)

	go c.queryBarListener(ctx)

	return nil
}

func (c *Content) setStyle() {
	c.Table.SetBackgroundColor(tcell.NewRGBColor(0, 10, 19))
	c.Table.SetBorder(true)
	c.Table.SetTitle(" Content ")
	c.Table.SetTitleAlign(tview.AlignLeft)
	c.Table.SetTitleColor(tcell.ColorSteelBlue)
	c.Table.SetBorderColor(tcell.ColorSteelBlue)
	c.Table.SetBorderPadding(0, 0, 1, 1)
	c.Table.SetFixed(1, 1)
	c.Table.SetSelectable(true, false)

	c.Flex.SetBackgroundColor(tcell.NewRGBColor(0, 10, 19))
	c.Flex.SetDirection(tview.FlexRow)
}

func (c *Content) setShortcuts(ctx context.Context) {
	c.Table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'v':
			c.docView.DocViewer(ctx, c.state.db, c.state.coll, c.Table.GetCell(c.Table.GetSelection()).Text)
		case 'e':
			c.docView.DocEdit(ctx, c.state.db, c.state.coll, c.Table.GetCell(c.Table.GetSelection()).Text, c.refresh)
		case '/':
			c.toggleQueryBar(ctx)
			c.render(ctx)
		}
		switch event.Key() {
		case tcell.KeyCtrlN:
			c.goToNextMongoPage(ctx)
		case tcell.KeyCtrlP:
			c.goToPrevMongoPage(ctx)
		case tcell.KeyEnter:
			c.docView.DocViewer(ctx, c.state.db, c.state.coll, c.Table.GetCell(c.Table.GetSelection()).Text)
		}

		return event
	})
}

func (c *Content) render(ctx context.Context) {
	c.Flex.Clear()

	if c.queryBar.IsEnabled() {
		c.Flex.AddItem(c.queryBar, 3, 0, false)
		defer c.app.SetFocus(c.queryBar)
	} else {
		defer c.app.SetFocus(c.Table)
	}

	c.Flex.AddItem(c.Table, 0, 1, true)
}

func (c *Content) toggleQueryBar(ctx context.Context) {
	c.queryBar.Toggle()
	c.render(ctx)
}

func (c *Content) queryBarListener(ctx context.Context) {
	eventChan := c.queryBar.EventChan

	for {
		key := <-eventChan
		if _, ok := key.(tcell.Key); !ok {
			continue
		}
		switch key {
		case tcell.KeyEsc:
			c.app.QueueUpdateDraw(func() {
				c.toggleQueryBar(ctx)

			})
		case tcell.KeyEnter:
			c.app.QueueUpdateDraw(func() {
				c.toggleQueryBar(ctx)
				filter := map[string]interface{}{}
				text := c.queryBar.GetText()
				if text != "" {
					err := bson.UnmarshalExtJSON([]byte(text), true, &filter)
					if err != nil {
						log.Printf("Error parsing query: %v", err)
					}
				}
				err := c.queryBar.SaveToHistory(text)
				if err != nil {
					log.Printf("Error saving to history: %v", err)
				}
				c.RenderContent(c.state.db, c.state.coll, filter)
				c.Table.ScrollToBeginning()
			})
		}
	}
}

func (c *Content) getPrimitiveByLabel(label string) tview.Primitive {
	switch label {
	case "query":
		return c.queryBar
	case "content":
		return c.Table
	default:
		return nil
	}
}

func (c *Content) listDocuments(db, coll string, filters map[string]interface{}) ([]string, int64, error) {
	ctx := context.Background()
	c.state.db = db
	c.state.coll = coll

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

func (c *Content) RenderContent(db, coll string, filter map[string]interface{}) error {
	c.Table.Clear()
	c.app.SetFocus(c.Table)

	c.state.db = db
	c.state.coll = coll

	documents, count, err := c.listDocuments(db, coll, filter)
	if err != nil {
		return err
	}

	if count == 0 {
		noDocCell := tview.NewTableCell("No documents found").
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft).
			SetSelectable(false)

		c.Table.SetCell(1, 1, noDocCell)
		return nil
	}

	headerInfo := fmt.Sprintf("Documents: %d, Page: %d, Limit: %d", count, c.state.page, c.state.limit)
	headerCell := tview.NewTableCell(headerInfo).
		SetTextColor(tcell.ColorWhite).
		SetAlign(tview.AlignLeft).
		SetSelectable(false)

	c.Table.SetCell(0, 0, headerCell)

	for i, d := range documents {
		dataCell := tview.NewTableCell(d).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft)

		c.Table.SetCell(i+1, 0, dataCell)
	}

	c.Table.ScrollToBeginning()

	return nil
}

func (c *Content) refresh() {
	c.RenderContent(c.state.db, c.state.coll, nil)
}

func (c *Content) goToNextMongoPage(ctx context.Context) {
	c.state.page += c.state.limit
	c.RenderContent(c.state.db, c.state.coll, nil)
}

func (c *Content) goToPrevMongoPage(ctx context.Context) {
	c.state.page -= c.state.limit
	c.RenderContent(c.state.db, c.state.coll, nil)
}
