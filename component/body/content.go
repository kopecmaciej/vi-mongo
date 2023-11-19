package body

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mongo-ui/mongo"
	"mongo-ui/primitives"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Content struct {
	*tview.Table

	app   *tview.Application
	dao   *mongo.Dao
	label string
	state state
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
	return &Content{
		Table: tview.NewTable(),

		dao:   dao,
		label: "content",
		state: state,
	}
}

func (c *Content) Init(ctx context.Context) error {
	c.app = ctx.Value("app").(*tview.Application)
	c.setStyle()
	c.SetShortcuts(ctx)
	return nil
}

func (c *Content) RenderDocuments(ctx context.Context, db, coll string) error {
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

func (c *Content) showDocumentDetails(ctx context.Context, db, coll string, rawDocument string) error {
	var prettyJson bytes.Buffer
	err := json.Indent(&prettyJson, []byte(rawDocument), "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return nil
	}
	text := string(prettyJson.Bytes())

	modal := primitives.NewModalView()
	modal.SetBackgroundColor(tcell.ColorDefault)
	modal.SetBorder(true)
	modal.SetTitle("Document Details")
	modal.SetTitleAlign(tview.AlignLeft)
	modal.SetTitleColor(tcell.ColorSteelBlue)
	modal.SetBorderColor(tcell.ColorSteelBlue)

	modal.SetText(primitives.Text{
		Content: text,
		Color:   tcell.ColorWhite,
		Align:   tview.AlignLeft,
	})

	modal.AddButtons([]string{"Edit", "Close"})

	root := ctx.Value("root").(*tview.Pages)
	root.AddPage("details", modal, true, true)
	c.app.SetFocus(modal)
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Edit" {
			c.editDocument(ctx, db, coll, rawDocument)
		} else {
			root.RemovePage("details")
			c.app.SetFocus(c)
		}
	})
	return nil
}

func (c *Content) editDocument(ctx context.Context, db, coll string, rawDocument string) error {
	var prettyJson bytes.Buffer
	err := json.Indent(&prettyJson, []byte(rawDocument), "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return nil
	}
	text := string(prettyJson.Bytes())

	id := string(prettyJson.Bytes()[8:40])

	editor := tview.NewTextArea()
	editor.SetBackgroundColor(tcell.ColorDefault)
	editor.SetBorder(true)
	editor.SetTitle(id)
	editor.SetTitleAlign(tview.AlignLeft)
	editor.SetTitleColor(tcell.ColorSteelBlue)

	editor.SetText(text, false)

	root := ctx.Value("root").(*tview.Pages)
	root.AddPage("editor", editor, true, true)
	c.app.SetFocus(editor)
	editor.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlS:
			err := c.saveDocument(ctx, db, coll, editor.GetText())
			if err != nil {
				log.Printf("Error saving document: %v", err)
			}
			root.RemovePage("editor")
			c.app.SetFocus(c)
		case tcell.KeyCtrlC, tcell.KeyEsc:
			root.RemovePage("editor")
			c.app.SetFocus(c)
		}
		return event
	})

	return nil
}

func (c *Content) saveDocument(ctx context.Context, db, coll string, rawDocument string) error {
	var document map[string]interface{}
	err := json.Unmarshal([]byte(rawDocument), &document)
	if err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return nil
	}
	id := document["_id"].(string)

	if id == "" {
		return fmt.Errorf("Document must have an _id")
	}
	mongoId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("Invalid _id: %v", err)
	}

	err = c.dao.UpdateDocument(ctx, db, coll, mongoId, document)
	if err != nil {
		log.Printf("Error updating document: %v", err)
		return nil
	}

	return nil
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
		switch event.Key() {
		case tcell.KeyCtrlN:
			c.goToNextMongoPage(ctx)
		case tcell.KeyCtrlP:
			c.goToPrevMongoPage(ctx)
			// v or enter
		case tcell.KeyEnter:
			document := c.GetCell(c.GetSelection()).Text
			c.showDocumentDetails(ctx, c.state.db, c.state.coll, document)
		}

		return event
	})
}

func (c *Content) goToNextMongoPage(ctx context.Context) {
	c.state.page += c.state.limit
	c.RenderDocuments(ctx, c.state.db, c.state.coll)
}

func (c *Content) goToPrevMongoPage(ctx context.Context) {
	c.state.page -= c.state.limit
	c.RenderDocuments(ctx, c.state.db, c.state.coll)
}
