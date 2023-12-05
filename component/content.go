package component

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/manager"
	"github.com/kopecmaciej/mongui/mongo"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
)

const (
	ContentComponent  manager.Component = "Content"
	JsonViewComponent manager.Component = "JsonView"
)

type Content struct {
	*tview.Flex

	Table       *tview.Table
	View        *tview.TextView
	app         *App
	dao         *mongo.Dao
	queryBar    *InputBar
	jsonPeeker  *JsonPeeker
	deleteModal *DeleteModal
	docModifier *DocModifier
	state       mongo.CollectionState
}

func NewContent(dao *mongo.Dao) *Content {
	state := mongo.CollectionState{
		Page:  0,
		Limit: 50,
	}

	flex := tview.NewFlex()
	return &Content{
		Table:       tview.NewTable(),
		Flex:        flex,
		View:        tview.NewTextView(),
		queryBar:    NewInputBar("Query"),
		jsonPeeker:  NewTextPeeker(dao),
		deleteModal: NewDeleteModal(),
		docModifier: NewDocModifier(dao),
		dao:         dao,
		state:       state,
	}
}

func (c *Content) Init(ctx context.Context) error {
	app, err := GetApp(ctx)
	if err != nil {
		return err
	}
	c.app = app

	c.setStyle()
	c.setShortcuts(ctx)

	if err := c.jsonPeeker.Init(ctx, c.Flex); err != nil {
		return err
	}
	if err := c.deleteModal.Init(ctx); err != nil {
		return err
	}
	c.queryBar.AutocompleteOn = true
	if err := c.queryBar.Init(ctx); err != nil {
		return err
	}
	if err := c.docModifier.Init(ctx); err != nil {
		return err
	}
	c.docModifier.Render = func() error {
    // TODO: change to return from editJson
		return c.refresh(ctx)
	}

	c.render(ctx, false)

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
			c.jsonPeeker.PeekJson(ctx, c.state.Db, c.state.Coll, c.Table.GetCell(c.Table.GetSelection()).Text)
		case 'e':
			c.docModifier.Edit(ctx, c.state.Db, c.state.Coll, c.Table.GetCell(c.Table.GetSelection()).Text)
		case 'v':
			c.viewJson(ctx, c.Table.GetCell(c.Table.GetSelection()).Text)
		case '/':
			c.toggleQueryBar(ctx)
			c.render(ctx, true)
		}
		switch event.Key() {
		case tcell.KeyCtrlD:
			c.deleteDocument(ctx, c.Table.GetCell(c.Table.GetSelection()).Text)
		case tcell.KeyCtrlA:
			// c.addDocument(ctx, c.state.db, c.state.coll, c.refresh)
		case tcell.KeyCtrlS:
			// c.duplicateDocument(ctx, c.state.db, c.state.coll, c.Table.GetCell(c.Table.GetSelection()).Text, c.refresh)
		case tcell.KeyCtrlN:
			c.goToNextMongoPage(ctx)
		case tcell.KeyCtrlP:
			c.goToPrevMongoPage(ctx)
		case tcell.KeyEnter:
			c.jsonPeeker.PeekJson(ctx, c.state.Db, c.state.Coll, c.Table.GetCell(c.Table.GetSelection()).Text)
		}

		return event
	})
}

func (c *Content) render(ctx context.Context, setFocus bool) {
	c.Flex.Clear()

	var focusPrimitive tview.Primitive
	focusPrimitive = c

	if c.queryBar.IsEnabled() {
		c.Flex.AddItem(c.queryBar, 3, 0, false)
		focusPrimitive = c.queryBar
	}
	if setFocus {
		defer c.app.SetFocus(focusPrimitive)
	}

	c.Flex.AddItem(c.Table, 0, 1, true)
}

func (c *Content) toggleQueryBar(ctx context.Context) {
	c.queryBar.Toggle()
	c.render(ctx, true)
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
					log.Error().Err(err).Msg("Error parsing query")
				}
				err = c.queryBar.SaveToHistory(text)
				if err != nil {
					log.Error().Err(err).Msg("Error saving query to history")
				}
				c.RenderContent(ctx, c.state.Db, c.state.Coll, filter)
				c.Table.ScrollToBeginning()
			})
		}
	}
}

func (c *Content) listDocuments(ctx context.Context, db, coll string, filters map[string]interface{}) ([]string, int64, error) {
	c.state.Db = db
	c.state.Coll = coll

	documents, count, err := c.dao.ListDocuments(ctx, db, coll, filters, c.state.Page, c.state.Limit)
	if err != nil {
		return nil, 0, err
	}
	if len(documents) == 0 {
		return nil, 0, nil
	}

	c.state.Count = count

	docsWithOid, err := mongo.ConvertIdsToOids(documents)
	if err != nil {
		return nil, 0, err
	}

	return docsWithOid, count, nil
}

func (c *Content) RenderContent(ctx context.Context, db, coll string, filter map[string]interface{}) error {
	c.Table.Clear()
	c.app.SetFocus(c.Table)

	documents, count, err := c.listDocuments(ctx, db, coll, filter)
	if err != nil {
		log.Error().Err(err).Msg("Error listing documents")
		return err
	}

	if count == 0 {
		noDocCell := tview.NewTableCell("No documents found").
			SetAlign(tview.AlignLeft).
			SetSelectable(false)

		c.Table.SetCell(1, 1, noDocCell)
		return nil
	}

	headerInfo := fmt.Sprintf("Documents: %d, Page: %d, Limit: %d", c.state.Count, c.state.Page, c.state.Limit)
	if filter != nil {
		prettyFilter, err := json.Marshal(filter)
		if err != nil {
			log.Error().Err(err).Msg("Error marshaling filter")
			return err
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

func (c *Content) refresh(ctx context.Context) error {
	return c.RenderContent(ctx, c.state.Db, c.state.Coll, nil)
}

func (c *Content) goToNextMongoPage(ctx context.Context) {
	if c.state.Page+c.state.Limit >= c.state.Count {
		return
	}
	c.state.Page += c.state.Limit
	c.RenderContent(ctx, c.state.Db, c.state.Coll, nil)
}

func (c *Content) goToPrevMongoPage(ctx context.Context) {
	if c.state.Page == 0 {
		return
	}
	c.state.Page -= c.state.Limit
	c.RenderContent(ctx, c.state.Db, c.state.Coll, nil)
}

func (c *Content) viewJson(ctx context.Context, jsonString string) error {
	c.View.Clear()

	c.app.Root.AddPage(JsonViewComponent, c.View, true, true)

	indentedJson, err := mongo.IndientJSON(jsonString)
	if err != nil {
		return err
	}

	c.View.SetText(indentedJson)
	c.View.ScrollToBeginning()

	c.View.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			c.app.Root.RemovePage(JsonViewComponent)
		}
		return event
	})

	return nil
}

func (c *Content) deleteDocument(ctx context.Context, jsonString string) error {
	objectID, err := mongo.GetIDFromJSON(jsonString)

	delMod := c.deleteModal
	delMod.SetText("Are you sure you want to delete this document?")
	delMod.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Delete" {
			err = c.dao.DeleteDocument(ctx, c.state.Db, c.state.Coll, objectID)
			if err != nil {
				log.Error().Err(err).Msg("Error deleting document")
			}
		}
		c.app.Root.RemovePage(DeleteModalComponent)
		c.RenderContent(ctx, c.state.Db, c.state.Coll, nil)
	})

	c.app.Root.AddPage(DeleteModalComponent, delMod, true, true)

	return nil
}

// EditJson opens the editor with the document and saves it if it was changed
func (c *Content) EditJson(ctx context.Context, db, coll string, rawDocument string, render func() error) error {
	var prettyJson bytes.Buffer
	err := json.Indent(&prettyJson, []byte(rawDocument), "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return err
	}

	tmpFile, err := os.CreateTemp("", "doc-*.json")
	if err != nil {
		return fmt.Errorf("Error creating temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write(prettyJson.Bytes())
	if err != nil {
		return fmt.Errorf("Error writing to temp file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("Error closing temp file: %v", err)
	}

	editor, err := exec.LookPath(os.Getenv("EDITOR"))
	if err != nil {
		return fmt.Errorf("Error finding editor: %v", err)
	}

	c.app.Suspend(func() {
		cmd := exec.Command(editor, tmpFile.Name())
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Printf("Error running editor: %v", err)
			return
		}

		editedBytes, err := os.ReadFile(tmpFile.Name())
		if err != nil {
			log.Printf("Error reading edited file: %v", err)
			return
		}
		if !json.Valid(editedBytes) {
			log.Printf("Edited JSON is not valid")
			return
		}
		if string(editedBytes) == string(prettyJson.Bytes()) {
			log.Debug().Msgf("Edited JSON is the same as original")
			return
		}

		err = c.saveDocument(ctx, db, coll, string(editedBytes))
		if err != nil {
			log.Printf("Error saving edited document: %v", err)
			return
		} else {
			log.Debug().Msg("Document saved")
			err := render()
			if err != nil {
				// TODO: show modal with error
				log.Printf("Error rendering: %v", err)
				return
			}
		}

	})

	return nil
}

func (c *Content) saveDocument(ctx context.Context, db, coll string, rawDocument string) error {
	if rawDocument == "" {
		return fmt.Errorf("Document cannot be empty")
	}
	var document map[string]interface{}
	err := json.Unmarshal([]byte(rawDocument), &document)
	if err != nil {
		log.Error().Msgf("Error unmarshaling JSON: %v", err)
		return nil
	}

	id, err := mongo.GetIDFromDocument(document)
	if err != nil {
		log.Error().Msgf("Error getting _id from document: %v", err)
		return nil
	}
	delete(document, "_id")

	err = c.dao.UpdateDocument(ctx, db, coll, id, document)
	if err != nil {
		log.Error().Msgf("Error updating document: %v", err)
		return nil
	}

	return nil
}
