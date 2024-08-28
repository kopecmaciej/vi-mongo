package component

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/tui/modal"
	"github.com/kopecmaciej/vi-mongo/internal/util"
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
	*core.BaseElement
	*tview.Flex

	Table          *tview.Table
	View           *tview.TextView
	style          *config.ContentStyle
	queryBar       *InputBar
	sortBar        *InputBar
	jsonPeeker     *DocPeeker
	deleteModal    *modal.DeleteModal
	docModifier    *DocModifier
	state          mongo.CollectionState
	stateMap       map[string]mongo.CollectionState
	isMultiRowView bool
}

func NewContent() *Content {
	c := &Content{
		BaseElement:    core.NewBaseElement("Content"),
		Table:          tview.NewTable(),
		Flex:           tview.NewFlex(),
		View:           tview.NewTextView(),
		queryBar:       NewInputBar(QueryBarView, "Query"),
		sortBar:        NewInputBar(SortBarView, "Sort"),
		jsonPeeker:     NewDocPeeker(),
		deleteModal:    modal.NewDeleteModal(),
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

	if err := c.jsonPeeker.Init(c.App); err != nil {
		return err
	}
	if err := c.docModifier.Init(c.App); err != nil {
		return err
	}
	if err := c.deleteModal.Init(c.App); err != nil {
		return err
	}
	if err := c.queryBar.Init(c.App); err != nil {
		return err
	}
	if err := c.sortBar.Init(c.App); err != nil {
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
	c.style = &c.App.GetStyles().Content
	c.Table.SetBorder(true)
	c.Table.SetTitle(" Content ")
	c.Table.SetTitleAlign(tview.AlignLeft)
	c.Table.SetBorderPadding(0, 0, 1, 1)
	c.Table.SetFixed(1, 1)
	c.Table.SetSelectable(true, false)
	c.Table.SetBackgroundColor(c.style.BackgroundColor.Color())
	c.Table.SetBorderColor(c.style.BorderColor.Color())

	c.View.SetBorder(true)
	c.View.SetTitle(" JSON View ")
	c.View.SetTitleAlign(tview.AlignCenter)
	c.View.SetBorderPadding(2, 0, 6, 0)
	c.View.SetBorderColor(c.style.BorderColor.Color())

	c.Flex.SetDirection(tview.FlexRow)
}

func (c *Content) setKeybindings(ctx context.Context) {
	k := c.App.GetKeys()

	c.Table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := c.Table.GetSelection()
		c.handleScrolling(row)
		switch {
		case k.Contains(k.Root.Content.SwitchView, event.Name()):
			return c.handleSwitchView(ctx)
		case k.Contains(k.Root.Content.PeekDocument, event.Name()):
			return c.handlePeekDocument(ctx, row)
		case k.Contains(k.Root.Content.ViewDocument, event.Name()):
			return c.handleViewDocument(row)
		case k.Contains(k.Root.Content.AddDocument, event.Name()):
			return c.handleAddDocument(ctx)
		case k.Contains(k.Root.Content.EditDocument, event.Name()):
			return c.handleEditDocument(ctx, row)
		case k.Contains(k.Root.Content.DuplicateDocument, event.Name()):
			return c.handleDuplicateDocument(ctx, row)
		case k.Contains(k.Root.Content.DeleteDocument, event.Name()):
			return c.handleDeleteDocument(ctx, row)
		case k.Contains(k.Root.Content.ToggleQuery, event.Name()):
			return c.handleToggleQuery()
		case k.Contains(k.Root.Content.ToggleSort, event.Name()):
			return c.handleToggleSort()
		case k.Contains(k.Root.Content.Refresh, event.Name()):
			return c.handleRefresh(ctx)
		case k.Contains(k.Root.Content.NextPage, event.Name()):
			return c.handleNextPage(ctx)
		case k.Contains(k.Root.Content.NextDocument, event.Name()):
			return c.handleNextDocument(row)
		case k.Contains(k.Root.Content.PreviousDocument, event.Name()):
			return c.handlePreviousDocument(row)
		case k.Contains(k.Root.Content.PreviousPage, event.Name()):
			return c.handlePreviousPage(ctx)
		// TODO: use this in multiple delete, think of other usage
		case k.Contains(k.Root.Content.MultipleSelect, event.Name()):
			return c.handleMultipleSelect(row)
		case k.Contains(k.Root.Content.ClearSelection, event.Name()):
			return c.handleClearSelection()
		case k.Contains(k.Root.Content.CopyLine, event.Name()):
			return c.handleCopyLine(row)
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
		c.App.SetFocus(focusPrimitive)
	}
}

func (c *Content) listDocuments(ctx context.Context) ([]primitive.M, int64, error) {
	filter, err := mongo.ParseStringQuery(c.state.Filter)
	if err != nil {
		return nil, 0, err
	}
	sort, err := mongo.ParseStringQuery(c.state.Sort)
	if err != nil {
		return nil, 0, err
	}

	documents, count, err := c.Dao.ListDocuments(ctx, &c.state, filter, sort)
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
	c.App.SetFocus(c.Table)

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
			modal.ShowError(c.App.Pages, "Error stringifying document", err)
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
			line = util.TrimMultipleSpaces(line)
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
		c.state.Filter = util.TrimJson(text)
		collectionKey := c.state.Db + "." + c.state.Coll
		c.stateMap[collectionKey] = c.state
		err := c.updateContent(ctx)
		if err != nil {
			modal.ShowError(c.App.Pages, "Error updating content", err)
		}
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
		c.state.Sort = util.TrimJson(text)
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

	c.App.Pages.AddPage(JsonViewView, c.View, true, true)

	indentedJson, err := mongo.IndentJson(jsonString)
	if err != nil {
		return err
	}

	c.View.SetText(indentedJson.String())
	c.View.ScrollToBeginning()

	c.View.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			c.App.Pages.RemovePage(JsonViewView)
			c.App.SetFocus(c.Table)
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
			err = c.Dao.DeleteDocument(ctx, c.state.Db, c.state.Coll, objectID)
			if err != nil {
				defer modal.ShowError(c.App.Pages, "Error deleting document", err)
			}
		}
		c.App.Pages.RemovePage(c.deleteModal.GetIdentifier())
		c.updateContent(ctx)
	})

	c.App.Pages.AddPage(c.deleteModal.GetIdentifier(), c.deleteModal, true, true)

	return nil
}

func (c *Content) getDocumentBasedOnView(row int) (string, error) {
	if c.isMultiRowView {
		return c.getMultiRowDocument(row)
	}
	return c.Table.GetCell(row, 0).Text, nil
}

func (c *Content) getMultiRowDocument(row int) (string, error) {
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

// Helper functions (implement these separately)

func (c *Content) handleScrolling(row int) {
	if row == 3 {
		c.Table.ScrollToBeginning()
	}
	if row == c.Table.GetRowCount()-2 {
		c.Table.ScrollToEnd()
	}
}

func (c *Content) handleSwitchView(ctx context.Context) *tcell.EventKey {
	c.isMultiRowView = !c.isMultiRowView
	c.updateContent(ctx)
	return nil
}

func (c *Content) handlePeekDocument(ctx context.Context, row int) *tcell.EventKey {
	doc, err := c.getDocumentBasedOnView(row)
	if err != nil {
		modal.ShowError(c.App.Pages, "Error peeking document", err)
		return nil
	}
	c.jsonPeeker.Peek(ctx, c.state.Db, c.state.Coll, doc)
	return nil
}

func (c *Content) handleViewDocument(row int) *tcell.EventKey {
	doc, err := c.getDocumentBasedOnView(row)
	if err != nil {
		modal.ShowError(c.App.Pages, "Error viewing document", err)
		return nil
	}
	err = c.viewJson(doc)
	if err != nil {
		modal.ShowError(c.App.Pages, "Error viewing document", err)
		return nil
	}
	return nil
}

func (c *Content) handleAddDocument(ctx context.Context) *tcell.EventKey {
	ID, err := c.docModifier.Insert(ctx, c.state.Db, c.state.Coll)
	if err != nil {
		modal.ShowError(c.App.Pages, "Error adding document", err)
		return nil
	}
	insertedDoc, err := c.Dao.GetDocument(ctx, c.state.Db, c.state.Coll, ID)
	if err != nil {
		modal.ShowError(c.App.Pages, "Error getting inserted document", err)
		return nil
	}
	strDoc, err := mongo.ParseBsonDocument(insertedDoc)
	if err != nil {
		modal.ShowError(c.App.Pages, "Error stringifying document", err)
		return nil
	}
	c.addCell(strDoc)
	return nil
}

func (c *Content) handleEditDocument(ctx context.Context, row int) *tcell.EventKey {
	doc, err := c.getDocumentBasedOnView(row)
	if err != nil {
		modal.ShowError(c.App.Pages, "Error editing document", err)
		return nil
	}
	updated, err := c.docModifier.Edit(ctx, c.state.Db, c.state.Coll, doc)
	if err != nil {
		modal.ShowError(c.App.Pages, "Error editing document", err)
		return nil
	}
	if updated != "" {
		c.refreshDocument(updated)
	}
	return nil
}

func (c *Content) handleDuplicateDocument(ctx context.Context, row int) *tcell.EventKey {
	doc, err := c.getDocumentBasedOnView(row)
	if err != nil {
		modal.ShowError(c.App.Pages, "Error duplicating document", err)
		return nil
	}
	ID, err := c.docModifier.Duplicate(ctx, c.state.Db, c.state.Coll, doc)
	if err != nil {
		defer modal.ShowError(c.App.Pages, "Error duplicating document", err)
	}
	duplicatedDoc, err := c.Dao.GetDocument(ctx, c.state.Db, c.state.Coll, ID)
	if err != nil {
		defer modal.ShowError(c.App.Pages, "Error getting inserted document", err)
	}
	strDoc, err := mongo.ParseBsonDocument(duplicatedDoc)
	if err != nil {
		defer modal.ShowError(c.App.Pages, "Error stringifying document", err)
	}
	c.addCell(strDoc)
	return nil
}

func (c *Content) handleToggleQuery() *tcell.EventKey {
	if c.state.Filter != "" {
		c.queryBar.Toggle(c.state.Filter)
	} else {
		c.queryBar.Toggle("")
	}
	c.render(true)
	return nil
}

func (c *Content) handleToggleSort() *tcell.EventKey {
	if c.state.Sort != "" {
		c.sortBar.Toggle(c.state.Sort)
	} else {
		c.sortBar.Toggle("")
	}
	c.render(true)
	return nil
}

func (c *Content) handleDeleteDocument(ctx context.Context, row int) *tcell.EventKey {
	doc, err := c.getDocumentBasedOnView(row)
	if err != nil {
		modal.ShowError(c.App.Pages, "Error deleting document", err)
		return nil
	}
	err = c.deleteDocument(ctx, doc)
	if err != nil {
		modal.ShowError(c.App.Pages, "Error deleting document", err)
		return nil
	}
	return nil
}

func (c *Content) handleRefresh(ctx context.Context) *tcell.EventKey {
	err := c.updateContent(ctx)
	if err != nil {
		defer modal.ShowError(c.App.Pages, "Error refreshing documents", err)
	}
	return nil
}

func (c *Content) handleNextDocument(row int) *tcell.EventKey {
	if c.isMultiRowView {
		for i := row; i < c.Table.GetRowCount(); i++ {
			if strings.HasPrefix(c.Table.GetCell(i, 0).Text, `{`) {
				c.Table.Select(i, 0)
				return nil
			}
		}
	} else {
		c.Table.MoveDown()
	}
	return nil
}

func (c *Content) handlePreviousDocument(row int) *tcell.EventKey {
	if c.isMultiRowView {
		for i := row; i >= 0; i-- {
			if strings.HasPrefix(c.Table.GetCell(i, 0).Text, `}`) {
				c.Table.Select(i-1, 0)
				return nil
			}
		}
	} else {
		c.Table.MoveUp()
	}
	return nil
}

func (c *Content) handleNextPage(ctx context.Context) *tcell.EventKey {
	c.goToNextMongoPage(ctx)
	return nil
}

func (c *Content) handlePreviousPage(ctx context.Context) *tcell.EventKey {
	c.goToPrevMongoPage(ctx)
	return nil
}

func (c *Content) handleMultipleSelect(row int) *tcell.EventKey {
	c.Table.ToggleRowSelection(row)
	return nil
}

func (c *Content) handleClearSelection() *tcell.EventKey {
	c.Table.ClearSelection()
	return nil
}

func (c *Content) handleCopyLine(row int) *tcell.EventKey {
	selectedDoc := util.TrimJson(c.Table.GetCell(row, 0).Text)
	err := clipboard.WriteAll(selectedDoc)
	if err != nil {
		modal.ShowError(c.App.Pages, "Error copying document", err)
	}
	return nil
}
