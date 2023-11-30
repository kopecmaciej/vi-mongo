package component

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mongo-ui/mongo"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Content struct {
	*tview.Flex

	Table      *tview.Table
	View       *tview.TextView
	app        *App
	dao        *mongo.Dao
	queryBar   *InputBar
	textPeeker *TextPeeker
	label      string
	mutex      sync.Mutex
	state      contentState
}

type contentState struct {
	page  int64
	limit int64
	db    string
	coll  string
	count int64
}

func NewContent(dao *mongo.Dao) *Content {
	state := contentState{
		page:  0,
		limit: 50,
	}

	flex := tview.NewFlex()
	return &Content{
		Table:      tview.NewTable(),
		Flex:       flex,
		View:       tview.NewTextView(),
		queryBar:   NewInputBar("Query"),
		dao:        dao,
		textPeeker: NewTextPeeker(dao),
		mutex:      sync.Mutex{},
		label:      "content",
		state:      state,
	}
}

func (c *Content) Init(ctx context.Context) error {
	c.app = GetApp(ctx)
	c.setStyle()
	c.setShortcuts(ctx)

	if err := c.textPeeker.Init(ctx, c.Flex); err != nil {
		return err
	}
	c.queryBar.AutocompleteOn = true
	if err := c.queryBar.Init(ctx); err != nil {
		return err
	}

	c.render(ctx)

	go c.queryBarListener(ctx)

	return nil
}

func (c *Content) setStyle() {
	c.Table.SetBorder(true)
	c.Table.SetTitle(" Content ")
	c.Table.SetTitleAlign(tview.AlignLeft)
	c.Table.SetBorderPadding(0, 0, 1, 1)
	c.Table.SetFixed(1, 1)
	c.Table.SetSelectable(true, false)

	c.Flex.SetDirection(tview.FlexRow)
}

func (c *Content) setShortcuts(ctx context.Context) {
	c.Table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'p':
			c.textPeeker.PeekJson(ctx, c.state.db, c.state.coll, c.Table.GetCell(c.Table.GetSelection()).Text)
		case 'e':
			c.textPeeker.EditJson(ctx, c.state.db, c.state.coll, c.Table.GetCell(c.Table.GetSelection()).Text, c.refresh)
		case 'v':
			c.ViewJson(ctx, c.Table.GetCell(c.Table.GetSelection()).Text)
		case 'D':
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
			c.textPeeker.PeekJson(ctx, c.state.db, c.state.coll, c.Table.GetCell(c.Table.GetSelection()).Text)
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
				text := c.queryBar.GetText()
				filter, err := mongo.ParseStringQuery(text)
				if err != nil {
					log.Printf("Error parsing query: %v", err)
				}
				err = c.queryBar.SaveToHistory(text)
				if err != nil {
					log.Printf("Error saving to history: %v", err)
				}
				log.Printf("filter: %v", filter)
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	c.state.db = db
	c.state.coll = coll

	documents, count, err := c.dao.ListDocuments(ctx, db, coll, filters, c.state.page, c.state.limit)
	if err != nil {
		return nil, 0, err
	}

	c.state.count = count

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

	documents, count, err := c.listDocuments(db, coll, filter)
	if err != nil {
		log.Printf("Error listing documents: %v", err)
		return err
	}

	if count == 0 {
		noDocCell := tview.NewTableCell("No documents found").
			SetAlign(tview.AlignLeft).
			SetSelectable(false)

		c.Table.SetCell(1, 1, noDocCell)
		return nil
	}

	headerInfo := fmt.Sprintf("Documents: %d, Page: %d, Limit: %d", count, c.state.page, c.state.limit)
	if filter != nil {
		prettyFilter, err := json.Marshal(filter)
		if err != nil {
			log.Printf("Error marshaling JSON: %v", err)
		}
		headerInfo += fmt.Sprintf(", Filter: %v", string(prettyFilter))
	}
	headerCell := tview.NewTableCell(headerInfo).
		SetAlign(tview.AlignLeft).
		SetSelectable(false)

	c.Table.SetCell(0, 0, headerCell)

	for i, d := range documents {
		dataCell := tview.NewTableCell(d).
			SetAlign(tview.AlignLeft)

		c.Table.SetCell(i+2, 0, dataCell)
	}

	c.Table.ScrollToBeginning()

	return nil
}

func (c *Content) refresh() {
	c.RenderContent(c.state.db, c.state.coll, nil)
}

func (c *Content) goToNextMongoPage(ctx context.Context) {
	if c.state.page+c.state.limit >= c.state.count {
		return
	}
	c.state.page += c.state.limit
	c.RenderContent(c.state.db, c.state.coll, nil)
}

func (c *Content) goToPrevMongoPage(ctx context.Context) {
	if c.state.page == 0 {
		return
	}
	c.state.page -= c.state.limit
	c.RenderContent(c.state.db, c.state.coll, nil)
}

func (c *Content) ViewJson(ctx context.Context, jsonString string) error {
	c.View.Clear()

	c.app.Root.AddPage("json", c.View, true, true)

	var prettyJson bytes.Buffer
	err := json.Indent(&prettyJson, []byte(jsonString), "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return nil
	}
	text := string(prettyJson.Bytes())
	c.View.SetText(text)
	c.View.ScrollToBeginning()

	c.app.SetFocus(c.View)

	c.View.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			c.app.Root.RemovePage("json")
			c.app.SetFocus(c.Table)
		}
		return event
	})

	return nil
}

func (c *Content) deleteDocument(ctx context.Context, jsonString string) error {
	var doc map[string]interface{}
	err := json.Unmarshal([]byte(jsonString), &doc)
	if err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return nil
	}

	objectID, err := primitive.ObjectIDFromHex(doc["_id"].(string))
	if err != nil {
		log.Printf("Error converting _id to ObjectID: %v", err)
		return nil
	}

	// show modal
	text := "Are you sure you want to delete this document?"
	modal := tview.NewModal().
		SetText(text).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				c.deleteDocument(ctx, c.Table.GetCell(c.Table.GetSelection()).Text)
			}
			c.app.Root.RemovePage("modal")
			c.app.SetFocus(c.Table)
		})

	c.app.Root.AddPage("modal", modal, true, true)
	c.app.SetFocus(modal)

	err = c.dao.DeleteDocument(ctx, c.state.db, c.state.coll, objectID)
	if err != nil {
		log.Printf("Error deleting document: %v", err)
		return nil
	}

	c.app.Root.RemovePage("modal")
	c.app.SetFocus(c.Table)

	c.RenderContent(c.state.db, c.state.coll, nil)

	return nil
}
