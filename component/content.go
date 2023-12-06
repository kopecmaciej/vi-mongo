package component

import (
	"context"
	"encoding/json"
	"fmt"

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
	jsonPeeker  *DocPeeker
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
		jsonPeeker:  NewDocPeeker(dao),
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

	if err := c.jsonPeeker.Init(ctx); err != nil {
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

	c.queryBarListener(ctx)

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
			c.jsonPeeker.Peek(ctx, c.state.Db, c.state.Coll, c.Table.GetCell(c.Table.GetSelection()).Text)
		case 'e':
			c.docModifier.Edit(ctx, c.state.Db, c.state.Coll, c.Table.GetCell(c.Table.GetSelection()).Text)
		case 'v':
			c.viewJson(ctx, c.Table.GetCell(c.Table.GetSelection()).Text)
		case '/':
			c.queryBar.Toggle()
			c.render(ctx, true)
		}
		switch event.Key() {
		case tcell.KeyCtrlD:
			c.deleteDocument(ctx, c.Table.GetCell(c.Table.GetSelection()).Text)
		case tcell.KeyCtrlA:
			// c.addDocument(ctx, c.state.db, c.state.coll, c.refresh)
		case tcell.KeyCtrlS:
			// c.duplicateDocument(ctx, c.state.db, c.state.coll, c.Table.GetCell(c.Table.GetSelection()).Text, c.refresh)
		case tcell.KeyCtrlR:
			c.refresh(ctx)
		case tcell.KeyCtrlN:
			c.goToNextMongoPage(ctx)
		case tcell.KeyCtrlP:
			c.goToPrevMongoPage(ctx)
		case tcell.KeyEnter:
			c.jsonPeeker.Peek(ctx, c.state.Db, c.state.Coll, c.Table.GetCell(c.Table.GetSelection()).Text)
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

func (c *Content) queryBarListener(ctx context.Context) {
	accceptFunc := func() {
		text := c.queryBar.GetText()
		filter, err := mongo.ParseStringQuery(text)
		if err != nil {
			log.Error().Err(err).Msg("Error parsing query")
		}
		c.RenderContent(ctx, c.state.Db, c.state.Coll, filter)
		c.Table.Select(2, 0)
	}
	rejectFunc := func() {
		c.render(ctx, true)
	}

	go c.queryBar.EventListener(accceptFunc, rejectFunc)
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

	c.deleteModal.SetText("Are you sure you want to delete document of ID: [blue]" + objectID.Hex())
	c.deleteModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			err = c.dao.DeleteDocument(ctx, c.state.Db, c.state.Coll, objectID)
			if err != nil {
				errMsg := fmt.Sprintf("Error deleting document: %v", err)
				log.Error().Err(err).Msg(errMsg)
				defer ShowErrorModal(c.app.Root, errMsg)
			}
		}
		c.app.Root.RemovePage(DeleteModalComponent)
		c.RenderContent(ctx, c.state.Db, c.state.Coll, nil)
	})

	c.app.Root.AddPage(DeleteModalComponent, c.deleteModal, true, true)

	return nil
}
