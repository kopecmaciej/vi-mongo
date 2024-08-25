package view

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/internal/config"
	"github.com/kopecmaciej/mongui/internal/mongo"
	"github.com/kopecmaciej/mongui/internal/utils"
	"github.com/kopecmaciej/tview"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	ContentView  = "Content"
	JsonViewView = "JsonView"
	QueryBarView = "QueryBar"
	SortBarView  = "SortBar"
)

// Content is a view that displays documents in a table
type Content struct {
	*BaseView
	*tview.Flex

	Table          *tview.Table
	View           *tview.TextView
	style          *config.ContentStyle
	queryBar       *InputBar
	sortBar        *InputBar
	jsonPeeker     *DocPeeker
	deleteModal    *DeleteModal
	docModifier    *DocModifier
	state          mongo.CollectionState
	stateMap       map[string]mongo.CollectionState
	isMultiRowView bool
}

func NewContent() *Content {
	c := &Content{
		BaseView:       NewBaseView("Content"),
		Table:          tview.NewTable(),
		Flex:           tview.NewFlex(),
		View:           tview.NewTextView(),
		queryBar:       NewInputBar(QueryBarView, "Query"),
		sortBar:        NewInputBar(SortBarView, "Sort"),
		jsonPeeker:     NewDocPeeker(),
		deleteModal:    NewDeleteModal(),
		docModifier:    NewDocModifier(),
		state:          mongo.CollectionState{},
		stateMap:       make(map[string]mongo.CollectionState),
		isMultiRowView: false,
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
	if err := c.docModifier.Init(c.app); err != nil {
		return err
	}
	if err := c.deleteModal.Init(c.app); err != nil {
		return err
	}
	if err := c.queryBar.Init(c.app); err != nil {
		return err
	}
	if err := c.sortBar.Init(c.app); err != nil {
		return err
	}

	c.queryBar.EnableAutocomplete()
	c.queryBar.EnableHistory()
	c.queryBar.SetDefaultText("{ <$0> }")

	c.sortBar.EnableAutocomplete()
	c.sortBar.SetDefaultText("{ <$0> }")

	c.render(false)

	c.queryBarListener(ctx)
	c.sortBarListener(ctx)

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

func (c *Content) setKeybindings(ctx context.Context) {
	k := c.app.Keys

	c.Table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := c.Table.GetSelection()
		if row == 3 {
			c.Table.ScrollToBeginning()
		}
		if row == c.Table.GetRowCount()-2 {
			c.Table.ScrollToEnd()
		}
		switch {
		case k.Contains(k.Root.Content.SwitchView, event.Name()):
			c.isMultiRowView = !c.isMultiRowView
			c.updateContent(ctx)
			return nil
		case k.Contains(k.Root.Content.PeekDocument, event.Name()):
			doc, err := c.getDocumentBasedOnView()
			if err != nil {
				ShowErrorModal(c.app.Pages, "Error peeking document", err)
				return nil
			}
			c.jsonPeeker.Peek(ctx, c.state.Db, c.state.Coll, doc)
			return nil
		case k.Contains(k.Root.Content.ViewDocument, event.Name()):
			doc, err := c.getDocumentBasedOnView()
			if err != nil {
				ShowErrorModal(c.app.Pages, "Error viewing document", err)
				return nil
			}
			err = c.viewJson(doc)
			if err != nil {
				ShowErrorModal(c.app.Pages, "Error viewing document", err)
				return nil
			}
			return nil
		case k.Contains(k.Root.Content.AddDocument, event.Name()):
			ID, err := c.docModifier.Insert(ctx, c.state.Db, c.state.Coll)
			if err != nil {
				ShowErrorModal(c.app.Pages, "Error adding document", err)
				return nil
			}
			insertedDoc, err := c.dao.GetDocument(ctx, c.state.Db, c.state.Coll, ID)
			if err != nil {
				ShowErrorModal(c.app.Pages, "Error getting inserted document", err)
				return nil
			}
			strDoc, err := mongo.ParseBsonDocument(insertedDoc)
			if err != nil {
				ShowErrorModal(c.app.Pages, "Error stringifying document", err)
				return nil
			}
			c.addCell(strDoc)
			return nil
		case k.Contains(k.Root.Content.EditDocument, event.Name()):
			doc, err := c.getDocumentBasedOnView()
			if err != nil {
				ShowErrorModal(c.app.Pages, "Error editing document", err)
				return nil
			}
			updated, err := c.docModifier.Edit(ctx, c.state.Db, c.state.Coll, doc)
			if err != nil {
				ShowErrorModal(c.app.Pages, "Error editing document", err)
				return nil
			}
			if updated != "" {
				c.refreshDocument(updated)
			}
			return nil
		case k.Contains(k.Root.Content.DuplicateDocument, event.Name()):
			doc, err := c.getDocumentBasedOnView()
			if err != nil {
				ShowErrorModal(c.app.Pages, "Error duplicating document", err)
				return nil
			}
			ID, err := c.docModifier.Duplicate(ctx, c.state.Db, c.state.Coll, doc)
			if err != nil {
				defer ShowErrorModal(c.app.Pages, "Error duplicating document", err)
			}
			duplicatedDoc, err := c.dao.GetDocument(ctx, c.state.Db, c.state.Coll, ID)
			if err != nil {
				defer ShowErrorModal(c.app.Pages, "Error getting inserted document", err)
			}
			strDoc, err := mongo.ParseBsonDocument(duplicatedDoc)
			if err != nil {
				defer ShowErrorModal(c.app.Pages, "Error stringifying document", err)
			}
			c.addCell(strDoc)
			return nil
		case k.Contains(k.Root.Content.ToggleQuery, event.Name()):
			if c.state.Filter != "" {
				c.queryBar.Toggle(c.state.Filter)
			} else {
				c.queryBar.Toggle("")
			}
			c.render(true)
			return nil
		case k.Contains(k.Root.Content.ToggleSort, event.Name()):
			if c.state.Sort != "" {
				c.sortBar.Toggle(c.state.Sort)
			} else {
				c.sortBar.Toggle("")
			}
			c.render(true)
			return nil
		case k.Contains(k.Root.Content.DeleteDocument, event.Name()):
			doc, err := c.getDocumentBasedOnView()
			if err != nil {
				ShowErrorModal(c.app.Pages, "Error deleting document", err)
				return nil
			}
			err = c.deleteDocument(ctx, doc)
			if err != nil {
				defer ShowErrorModal(c.app.Pages, "Error deleting document", err)
			}
			return nil
		case k.Contains(k.Root.Content.Refresh, event.Name()):
			err := c.updateContent(ctx)
			if err != nil {
				defer ShowErrorModal(c.app.Pages, "Error refreshing documents", err)
			}
			return nil
		case k.Contains(k.Root.Content.NextPage, event.Name()):
			c.goToNextMongoPage(ctx)
			return nil
		case k.Contains(k.Root.Content.PreviousPage, event.Name()):
			c.goToPrevMongoPage(ctx)
			return nil
		case k.Contains(k.Root.Content.CopyLine, event.Name()):
			selectedDoc := utils.TrimJson(c.Table.GetCell(c.Table.GetSelection()).Text)
			err := clipboard.WriteAll(selectedDoc)
			if err != nil {
				ShowErrorModal(c.app.Pages, "Error copying document", err)
			}
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

	if c.sortBar.IsEnabled() {
		c.Flex.AddItem(c.sortBar, 3, 0, false)
		focusPrimitive = c.sortBar
	}

	c.Flex.AddItem(c.Table, 0, 1, true)

	if setFocus {
		c.app.SetFocus(focusPrimitive)
	}
}

func (c *Content) listDocuments(ctx context.Context) ([]primitive.M, int64, error) {
	documents, count, err := c.dao.ListDocuments(ctx, &c.state)
	if err != nil {
		return nil, 0, err
	}
	if len(documents) == 0 {
		return nil, 0, nil
	}

	c.state.Count = count
	c.state.Docs = documents

	c.loadAutocompleteKeys(documents)

	return documents, count, nil
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
	c.sortBar.LoadNewKeys(autocompleteKeys)
}

// HandleDatabaseSelection is called when a database/collection is selected in the DatabaseTree
func (c *Content) HandleDatabaseSelection(ctx context.Context, db, coll string) error {
	c.queryBar.SetText("")
	c.sortBar.SetText("")

	state, ok := c.stateMap[db+"."+coll]
	if ok {
		c.state = state
	} else {
		state = mongo.CollectionState{
			Page: 0,
		}
		_, _, _, height := c.Table.GetInnerRect()
		state.Limit = int64(height) - 3
		state.Db = db
		state.Coll = coll
		c.state = state
	}
	return c.updateContent(ctx)
}

func (c *Content) updateContent(ctx context.Context) error {
	c.Table.Clear()
	c.app.SetFocus(c.Table)

	documents, count, err := c.listDocuments(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Error listing documents")
		return err
	}

	headerInfo := fmt.Sprintf("Documents: %d, Page: %d, Limit: %d", count, c.state.Page, c.state.Limit)
	if c.state.Filter != "" {
		headerInfo += fmt.Sprintf(", Filter: %v", c.state.Filter)
	}
	if c.state.Sort != "" {
		headerInfo += fmt.Sprintf(", Sort: %v", c.state.Sort)
	}
	headerCell := tview.NewTableCell(headerInfo).
		SetAlign(tview.AlignLeft).
		SetSelectable(false)

	c.Table.SetCell(0, 0, headerCell)

	if count == 0 {
		// TODO: find why if selectable is set to false, program crashes
		c.Table.SetCell(2, 0, tview.NewTableCell("No documents found"))
	}

	if c.isMultiRowView {
		c.renderMultiRowView(documents)
	} else {
		c.renderSingleRowView(documents)
	}

	c.stateMap[c.state.Db+"."+c.state.Coll] = c.state

	return nil
}

func (c *Content) renderSingleRowView(documents []primitive.M) {
	parsedDocs, _ := mongo.ParseBsonDocuments(documents)
	row := 2
	for _, d := range parsedDocs {
		dataCell := tview.NewTableCell(d)
		dataCell.SetAlign(tview.AlignLeft)
		c.Table.SetCell(row, 0, dataCell)
		row++
	}
	c.Table.ScrollToBeginning()
}

func (c *Content) renderMultiRowView(documents []primitive.M) {
	row := 2
	for _, doc := range documents {
		jsoned, err := mongo.ParseBsonDocument(doc)
		if err != nil {
			ShowErrorModal(c.app.Pages, "Error stringifying document", err)
			return
		}
		c.multiRowDocument(jsoned, &row)
	}
	c.Table.ScrollToBeginning()
}

func (c *Content) multiRowDocument(doc string, row *int) {
	indentedJson, err := mongo.IndentJson(doc)
	if err != nil {
		log.Error().Err(err).Msg("Error indenting JSON")
		return
	}
	keyRegexWithIndent := regexp.MustCompile(`(?m)^\s{2}"([^"]+)":`)
	lines := strings.Split(indentedJson.String(), "\n")

	c.Table.SetCell(*row, 0, tview.NewTableCell(lines[0]).SetAlign(tview.AlignLeft).SetTextColor(tcell.ColorGreen).SetSelectable(false))
	*row++

	currLine := ""
	for i := 1; i < len(lines)-1; i++ {
		line := lines[i]
		if keyRegexWithIndent.MatchString(line) {
			if currLine != "" {
				c.Table.SetCell(*row, 0, tview.NewTableCell(currLine).SetAlign(tview.AlignLeft))
				*row++
			}
			currLine = line
		} else {
			line = utils.TrimMultipleSpaces(line)
			currLine += line
		}
	}

	if currLine != "" {
		c.Table.SetCell(*row, 0, tview.NewTableCell(currLine).SetAlign(tview.AlignLeft))
		*row++
	}

	c.Table.SetCell(*row, 0, tview.NewTableCell(lines[len(lines)-1]).SetAlign(tview.AlignLeft).SetTextColor(tcell.ColorGreen).SetSelectable(false))
	*row++
}

func (c *Content) queryBarListener(ctx context.Context) {
	acceptFunc := func(text string) {
		c.state.Filter = utils.TrimJson(text)
		collectionKey := c.state.Db + "." + c.state.Coll
		c.stateMap[collectionKey] = c.state
		c.updateContent(ctx)
		if c.Flex.HasItem(c.queryBar) {
			c.Flex.RemoveItem(c.queryBar)
		}
	}
	rejectFunc := func() {
		c.render(true)
	}

	c.queryBar.DoneFuncHandler(acceptFunc, rejectFunc)
}

func (c *Content) sortBarListener(ctx context.Context) {
	acceptFunc := func(text string) {
		c.state.Sort = utils.TrimJson(text)
		collectionKey := c.state.Db + "." + c.state.Coll
		c.stateMap[collectionKey] = c.state
		c.updateContent(ctx)
		if c.Flex.HasItem(c.sortBar) {
			c.Flex.RemoveItem(c.sortBar)
		}
	}
	rejectFunc := func() {
		c.render(true)
	}

	c.sortBar.DoneFuncHandler(acceptFunc, rejectFunc)
}

// refreshDocument refreshes the document in the table
func (c *Content) refreshDocument(doc string) {
	if c.isMultiRowView {
		c.refreshMultiRowDocument(doc)
	} else {
		trimmed := regexp.MustCompile(`(?m)^\s+`).ReplaceAllString(doc, "")
		trimmed = regexp.MustCompile(`(?m):\s+`).ReplaceAllString(trimmed, ":")
		c.refreshCell(trimmed)
	}
}

func (c *Content) refreshMultiRowDocument(doc string) {
	row, _ := c.Table.GetSelection()

	for i := row; i >= 0; i-- {
		if strings.HasPrefix(c.Table.GetCell(i, 0).Text, "{") {
			row = i
			break
		}
	}

	doc = strings.TrimSpace(doc)

	c.multiRowDocument(doc, &row)
}

// addCell adds a new cell to the table
func (c *Content) addCell(content string) {
	maxRow := c.Table.GetRowCount()
	c.Table.SetCell(maxRow, 0, tview.NewTableCell(content).SetAlign(tview.AlignLeft))
}

// refreshCell refreshes the cell with the new content
func (c *Content) refreshCell(content string) {
	row, col := c.Table.GetSelection()
	c.Table.SetCell(row, col, tview.NewTableCell(content).SetAlign(tview.AlignLeft))
}

func (c *Content) goToNextMongoPage(ctx context.Context) {
	if c.state.Page+c.state.Limit >= c.state.Count {
		return
	}
	c.state.Page += c.state.Limit
	collectionKey := c.state.Db + "." + c.state.Coll
	c.stateMap[collectionKey] = c.state
	c.updateContent(ctx)
}

func (c *Content) goToPrevMongoPage(ctx context.Context) {
	if c.state.Page == 0 {
		return
	}
	c.state.Page -= c.state.Limit
	collectionKey := c.state.Db + "." + c.state.Coll
	c.stateMap[collectionKey] = c.state
	c.updateContent(ctx)
}

func (c *Content) viewJson(jsonString string) error {
	c.View.Clear()

	c.app.Pages.AddPage(JsonViewView, c.View, true, true)

	indentedJson, err := mongo.IndentJson(jsonString)
	if err != nil {
		return err
	}

	c.View.SetText(indentedJson.String())
	c.View.ScrollToBeginning()

	c.View.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			c.app.Pages.RemovePage(JsonViewView)
			c.app.SetFocus(c.Table)
		}
		return event
	})

	return nil
}

func (c *Content) deleteDocument(ctx context.Context, jsonString string) error {
	objectID, err := mongo.GetIDFromJSON(jsonString)
	if err != nil {
		return err
	}

	var stringifyId string
	if objectID, ok := objectID.(primitive.ObjectID); ok {
		stringifyId = objectID.Hex()
	}
	if strID, ok := objectID.(string); ok {
		stringifyId = strID
	}

	c.deleteModal.SetText("Are you sure you want to delete document of ID: [blue]" + stringifyId)
	c.deleteModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			err = c.dao.DeleteDocument(ctx, c.state.Db, c.state.Coll, objectID)
			if err != nil {
				defer ShowErrorModal(c.app.Pages, "Error deleting document", err)
			}
		}
		c.app.Pages.RemovePage(c.deleteModal.GetIdentifier())
		c.updateContent(ctx)
	})

	c.app.Pages.AddPage(c.deleteModal.GetIdentifier(), c.deleteModal, true, true)

	return nil
}

func (c *Content) getDocumentBasedOnView() (string, error) {
	if c.isMultiRowView {
		return c.getMultiRowDocument()
	}
	return c.Table.GetCell(c.Table.GetSelection()).Text, nil
}

func (c *Content) getMultiRowDocument() (string, error) {
	row, _ := c.Table.GetSelection()

	_id := ""

	for i := row; i >= 0; i-- {
		cell := c.Table.GetCell(i, 0)
		if strings.Contains(cell.Text, `"_id":`) {
			_id = cell.Text
			break
		}
	}

	var value string
	if strings.Contains(_id, `"$oid"`) {
		value = strings.Split(_id, `"`)[5]
	} else {
		value = strings.Split(_id, `"`)[3]
	}
	for _, doc := range c.state.Docs {
		var idValue string
		if doc_id, ok := doc["_id"].(primitive.M); ok {
			idValue = doc_id["$oid"].(string)
		} else if strID, ok := doc["_id"].(string); ok {
			idValue = strID
		}
		if idValue == value {
			jsoned, err := mongo.ParseBsonDocument(doc)
			if err != nil {
				return "", err
			}
			indentedJson, err := mongo.IndentJson(jsoned)
			if err != nil {
				return "", err
			}
			return indentedJson.String(), nil
		}
	}
	return "", fmt.Errorf("document not found")
}
