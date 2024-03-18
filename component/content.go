package component

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/config"
	"github.com/kopecmaciej/mongui/mongo"
	"github.com/kopecmaciej/tview"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	ContentComponent  = "Content"
	JsonViewComponent = "JsonView"
)

// Content is a component that displays documents in a table
type Content struct {
	*Component
	*tview.Flex

	Table            *tview.Table
	View             *tview.TextView
	style            *config.ContentStyle
	queryBar         *InputBar
	jsonPeeker       *DocPeeker
	deleteModal      *DeleteModal
	docModifier      *DocModifier
	state            mongo.CollectionState
	autocompleteKeys []string
}

// NewContent creates a new Content component
// It also initializes all subcomponents
func NewContent() *Content {
	state := mongo.CollectionState{
		Page:  0,
		Limit: 50,
	}

	c := &Content{
		Component:   NewComponent("Content"),
		Table:       tview.NewTable(),
		Flex:        tview.NewFlex(),
		View:        tview.NewTextView(),
		queryBar:    NewInputBar("Query"),
		jsonPeeker:  NewDocPeeker(),
		deleteModal: NewDeleteModal(),
		docModifier: NewDocModifier(),
		state:       state,
	}

	c.SetAfterInitFunc(c.init)

	return c
}

func (c *Content) init() error {
	ctx := context.Background()

	c.setStyle()
	c.setKeybindings(ctx)

	if err := c.jsonPeeker.Init(c.app); err != nil {
		return err
	}
	if err := c.deleteModal.Init(c.app); err != nil {
		return err
	}
	if err := c.queryBar.Init(c.app); err != nil {
		return err
	}
	c.queryBar.EnableAutocomplete()
	c.queryBar.EnableHistory()
	c.queryBar.SetDefaultText("{ <$0> }")
	if err := c.docModifier.Init(c.app); err != nil {
		return err
	}

	c.render(false)

	c.queryBarListener(ctx)

	return nil
}

func (c *Content) setStyle() {
	c.style = &c.app.Styles.Content
	c.Table.SetBorder(true)
	c.Table.SetTitle(" Content ")
	c.Table.SetTitleAlign(tview.AlignLeft)
	c.Table.SetBorderPadding(0, 0, 1, 1)
	c.Table.SetFixed(1, 1)
	c.Table.SetSelectable(true, false)
	c.Table.SetBackgroundColor(c.style.BackgroundColor.Color())
	c.Table.SetBorderColor(c.style.BorderColor.Color())

	c.Flex.SetDirection(tview.FlexRow)
}

// SetKeybindings sets keybindings for the component
func (c *Content) setKeybindings(ctx context.Context) {
	k := c.app.Keys

	c.Table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case k.Contains(k.Root.Content.PeekDocument, event.Name()):
			err := c.jsonPeeker.Peek(ctx, c.state.Db, c.state.Coll, c.Table.GetCell(c.Table.GetSelection()).Text)
			if err != nil {
				defer ShowErrorModal(c.app.Root, "Error while peeking document", err)
			}
			return nil
		case k.Contains(k.Root.Content.ViewDocument, event.Name()):
			err := c.viewJson(c.Table.GetCell(c.Table.GetSelection()).Text)
			if err != nil {
				defer ShowErrorModal(c.app.Root, "Error while viewing document", err)
			}
			return nil
		case k.Contains(k.Root.Content.AddDocument, event.Name()):
			ID, err := c.docModifier.Insert(ctx, c.state.Db, c.state.Coll)
			if err != nil {
				defer ShowErrorModal(c.app.Root, "Error while adding document", err)
			}
			insertedDoc, err := c.dao.GetDocument(ctx, c.state.Db, c.state.Coll, ID)
			if err != nil {
				defer ShowErrorModal(c.app.Root, "Error while getting inserted document", err)
			}
			strDoc, err := mongo.StringifyDocument(insertedDoc)
			if err != nil {
				defer ShowErrorModal(c.app.Root, "Error while stringifying document", err)
			}
			c.addCell(strDoc)
			return nil
		case k.Contains(k.Root.Content.EditDocument, event.Name()):
			updated, err := c.docModifier.Edit(ctx, c.state.Db, c.state.Coll, c.Table.GetCell(c.Table.GetSelection()).Text)
			if err != nil {
				defer ShowErrorModal(c.app.Root, "Error while editing document", err)
			}
			c.refreshCell(updated)
			return nil
		case k.Contains(k.Root.Content.DuplicateDocument, event.Name()):
			ID, err := c.docModifier.Duplicate(ctx, c.state.Db, c.state.Coll, c.Table.GetCell(c.Table.GetSelection()).Text)
			if err != nil {
				defer ShowErrorModal(c.app.Root, "Error while duplicating document", err)
			}
			duplicatedDoc, err := c.dao.GetDocument(ctx, c.state.Db, c.state.Coll, ID)
			if err != nil {
				defer ShowErrorModal(c.app.Root, "Error while getting inserted document", err)
			}
			strDoc, err := mongo.StringifyDocument(duplicatedDoc)
			if err != nil {
				defer ShowErrorModal(c.app.Root, "Error while stringifying document", err)
			}
			c.addCell(strDoc)
			return nil
		case k.Contains(k.Root.Content.ToggleQuery, event.Name()):
			c.queryBar.Toggle()
			c.render(true)
			return nil
		case k.Contains(k.Root.Content.DeleteDocument, event.Name()):
			err := c.deleteDocument(ctx, c.Table.GetCell(c.Table.GetSelection()).Text)
			if err != nil {
				defer ShowErrorModal(c.app.Root, "Error while deleting document", err)
			}
			return nil
		case k.Contains(k.Root.Content.Refresh, event.Name()):
			err := c.refresh(ctx)
			if err != nil {
				defer ShowErrorModal(c.app.Root, "Error while refreshing documents", err)
			}
			return nil
		case k.Contains(k.Root.Content.NextPage, event.Name()):
			c.goToNextMongoPage(ctx)
			return nil
		case k.Contains(k.Root.Content.PreviousPage, event.Name()):
			c.goToPrevMongoPage(ctx)
			return nil
		}

		return event
	})
}

func (c *Content) render(setFocus bool) {
	c.Flex.Clear()

	var focusPrimitive tview.Primitive
	focusPrimitive = c

	if c.queryBar.IsEnabled() {
		c.Flex.AddItem(c.queryBar, 3, 0, false)
		focusPrimitive = c.queryBar
	}

	c.Flex.AddItem(c.Table, 0, 1, true)
	_, _, _, height := c.Flex.GetInnerRect()
	c.state.Limit = int64(height) - 4

	if setFocus {
		c.app.SetFocus(focusPrimitive)
	}
}

func (c *Content) queryBarListener(ctx context.Context) {
	accceptFunc := func(text string) {
		c.Flex.RemoveItem(c.queryBar)
		filter, err := mongo.ParseStringQuery(text)
		if err != nil {
			defer ShowErrorModal(c.app.Root, "Error parsing query", err)
		}
		c.RenderContent(ctx, c.state.Db, c.state.Coll, filter)
		c.Table.Select(2, 0)
	}
	rejectFunc := func() {
		c.render(true)
	}

	c.queryBar.DoneFuncHandler(accceptFunc, rejectFunc)
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

	c.loadAutocompleteKeys(documents)

	docsWithOid, err := mongo.ConvertIdsToOids(documents)
	if err != nil {
		return nil, 0, err
	}

	return docsWithOid, count, nil
}

func (c *Content) loadAutocompleteKeys(documents []primitive.M) {
	uniqueKeys := make(map[string]bool)

	var addKeys func(string, interface{})
	addKeys = func(prefix string, value interface{}) {
		switch v := value.(type) {
		case map[string]interface{}:
			for key, val := range v {
				fullKey := key
				if prefix != "" {
					fullKey = prefix + "." + key
				}
				addKeys(fullKey, val)
			}
		default:
			uniqueKeys[prefix] = true
		}
	}

	for _, doc := range documents {
		for key, value := range doc {
			if obj, ok := value.(primitive.M); ok {
				addKeys(key, obj)
				for k, v := range obj {
					fullKey := key + "." + k
					addKeys(fullKey, v)
				}
			} else {
				addKeys(key, value)
			}
		}
	}

	autocompleteKeys := make([]string, 0, len(uniqueKeys))
	for key := range uniqueKeys {
		autocompleteKeys = append(autocompleteKeys, key)
	}

	c.queryBar.LoadNewKeys(autocompleteKeys)
}

func (c *Content) RenderContent(ctx context.Context, db, coll string, filter map[string]interface{}) error {
	c.Table.Clear()
	c.app.SetFocus(c.Table)

	documents, count, err := c.listDocuments(ctx, db, coll, filter)
	if err != nil {
		log.Error().Err(err).Msg("Error listing documents")
		return err
	}

	headerInfo := fmt.Sprintf("Documents: %d, Page: %d, Limit: %d", count, c.state.Page, c.state.Limit)
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

	if count == 0 {
		// TODO: find why if selectable is set to false, program crashes
		c.Table.SetCell(2, 0, tview.NewTableCell("No documents found"))
	}

	for i, d := range documents {
		dataCell := tview.NewTableCell(d)
		dataCell.SetAlign(tview.AlignLeft)

		c.Table.SetCell(i+2, 0, dataCell)
		c.Table.Select(2, 0)
	}

	// c.Table.ScrollToBeginning()

	return nil
}

func (c *Content) refresh(ctx context.Context) error {
	return c.RenderContent(ctx, c.state.Db, c.state.Coll, nil)
}

// addCell adds a new cell to the table
func (c *Content) addCell(content string) {
	maxRow := c.Table.GetRowCount()
	c.Table.SetCell(maxRow, 0, tview.NewTableCell(content).SetAlign(tview.AlignLeft))
}

// refreshCell refreshes the cell with the new content
func (c *Content) refreshCell(content string) {
	// Trim the content, as in table we don't want to see new lines and spaces
	content = strings.ReplaceAll(content, "\n", "")
	content = strings.ReplaceAll(content, " ", "")
	row, col := c.Table.GetSelection()
	c.Table.SetCell(row, col, tview.NewTableCell(content).SetAlign(tview.AlignLeft))
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

func (c *Content) viewJson(jsonString string) error {
	c.View.Clear()

	c.app.Root.AddPage(JsonViewComponent, c.View, true, true)

	indentedJson, err := mongo.IndientJSON(jsonString)
	if err != nil {
		return err
	}

	c.View.SetText(string(indentedJson.Bytes()))
	c.View.ScrollToBeginning()

	c.View.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			c.app.Root.RemovePage(JsonViewComponent)
			c.app.SetFocus(c.Table)
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
				defer ShowErrorModal(c.app.Root, "Error deleting document", err)
			}
		}
		c.app.Root.RemovePage(c.deleteModal.GetIdentifier())
		c.RenderContent(ctx, c.state.Db, c.state.Coll, nil)
	})

	c.app.Root.AddPage(c.deleteModal.GetIdentifier(), c.deleteModal, true, true)

	return nil
}
