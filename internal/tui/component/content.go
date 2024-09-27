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
	"github.com/kopecmaciej/vi-mongo/internal/manager"
	"github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/tui/modal"
	"github.com/kopecmaciej/vi-mongo/internal/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	ContentComponent  = "Content"
	JsonViewComponent = "JsonView"
	QueryBarComponent = "QueryBar"
	SortBarComponent  = "SortBar"
)

type ViewType int

const (
	TableView ViewType = iota
	JsonView
	SingleLineView
)

// Content is a view that displays documents in a table
type Content struct {
	*core.BaseElement
	*core.Flex

	tableFlex   *core.Flex
	tableHeader *core.TextView
	table       *core.Table
	view        *core.TextView
	style       *config.ContentStyle
	queryBar    *InputBar
	sortBar     *InputBar
	peeker      *Peeker
	deleteModal *modal.Delete
	docModifier *DocModifier
	state       *mongo.CollectionState
	stateMap    map[string]*mongo.CollectionState
	currentView ViewType
}

func NewContent() *Content {
	c := &Content{
		BaseElement: core.NewBaseElement(),
		Flex:        core.NewFlex(),

		tableFlex:   core.NewFlex(),
		tableHeader: core.NewTextView(),
		table:       core.NewTable(),
		view:        core.NewTextView(),
		queryBar:    NewInputBar(QueryBarComponent, "Query"),
		sortBar:     NewInputBar(SortBarComponent, "Sort"),
		peeker:      NewPeeker(),
		deleteModal: modal.NewDeleteModal(),
		docModifier: NewDocModifier(),
		state:       &mongo.CollectionState{},
		stateMap:    make(map[string]*mongo.CollectionState),
		currentView: TableView,
	}

	c.SetIdentifier(ContentComponent)
	// neccesarry if focus is get back to content component
	// it's related to how tview package works
	c.table.SetIdentifier(ContentComponent)
	c.SetAfterInitFunc(c.init)

	return c
}

func (c *Content) init() error {
	ctx := context.Background()

	c.setStaticLayout()
	c.setStyle()
	c.setKeybindings(ctx)

	if err := c.peeker.Init(c.App); err != nil {
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

	c.queryBarListener(ctx)
	c.sortBarListener(ctx)

	c.peeker.SetDoneFunc(func() {
		c.updateContent(ctx, true)
	})

	c.handleEvents()

	return nil
}

func (c *Content) handleEvents() {
	go c.HandleEvents(ContentComponent, func(event manager.EventMsg) {
		switch event.Message.Type {
		case manager.StyleChanged:
			c.setStyle()
			c.updateContent(context.Background(), true)
		}
	})
}

func (c *Content) UpdateDao(dao *mongo.Dao) {
	c.table.Clear()
	c.BaseElement.UpdateDao(dao)
	c.docModifier.UpdateDao(dao)
}

func (c *Content) setStyle() {
	c.style = &c.App.GetStyles().Content
	styles := c.App.GetStyles()

	c.tableFlex.SetStyle(styles)
	c.tableHeader.SetStyle(styles)
	c.view.SetStyle(styles)
	c.Flex.SetStyle(styles)
	c.table.SetStyle(styles)

	c.tableFlex.SetBorderColor(c.style.SeparatorColor.Color())
	c.tableHeader.SetTextColor(c.style.StatusTextColor.Color())

	c.table.SetBordersColor(c.style.SeparatorColor.Color())
	c.table.SetSeparator(c.style.SeparatorSymbol.Rune())
}

func (c *Content) setStaticLayout() {
	c.tableFlex.SetBorder(true)
	c.tableFlex.SetDirection(tview.FlexRow)
	c.tableFlex.SetTitle(" Content ")
	c.tableFlex.SetTitleAlign(tview.AlignCenter)
	c.tableFlex.SetBorderPadding(0, 0, 1, 1)

	c.tableHeader.SetText("Documents: 0, Page: 0, Limit: 0")

	c.view.SetBorder(true)
	c.view.SetTitle(" JSON View ")
	c.view.SetTitleAlign(tview.AlignCenter)
	c.view.SetBorderPadding(2, 0, 6, 0)

	c.Flex.SetDirection(tview.FlexRow)
}

func (c *Content) setKeybindings(ctx context.Context) {
	k := c.App.GetKeys()

	c.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, coll := c.table.GetSelection()
		c.handleScrolling(row)
		switch {
		case k.Contains(k.Content.ChangeView, event.Name()):
			return c.handleSwitchView(ctx)
		case k.Contains(k.Content.PeekDocument, event.Name()):
			return c.handlePeekDocument(ctx, row, coll)
		case k.Contains(k.Content.ViewDocument, event.Name()):
			return c.handleViewDocument(row, coll)
		case k.Contains(k.Content.AddDocument, event.Name()):
			return c.handleAddDocument(ctx)
		case k.Contains(k.Content.EditDocument, event.Name()):
			return c.handleEditDocument(ctx, row, coll)
		case k.Contains(k.Content.DuplicateDocument, event.Name()):
			return c.handleDuplicateDocument(ctx, row, coll)
		case k.Contains(k.Content.DeleteDocument, event.Name()):
			return c.handleDeleteDocument(ctx, row, coll)
		case k.Contains(k.Content.ToggleQuery, event.Name()):
			return c.handleToggleQuery()
		case k.Contains(k.Content.ToggleSort, event.Name()):
			return c.handleToggleSort()
		// TODO: Add automatic sort by given column
		case k.Contains(k.Content.Refresh, event.Name()):
			return c.handleRefresh(ctx)
		case k.Contains(k.Content.NextPage, event.Name()):
			return c.handleNextPage(ctx)
		case k.Contains(k.Content.NextDocument, event.Name()):
			return c.handleNextDocument(row, coll)
		case k.Contains(k.Content.PreviousDocument, event.Name()):
			return c.handlePreviousDocument(row, coll)
		case k.Contains(k.Content.PreviousPage, event.Name()):
			return c.handlePreviousPage(ctx)
		// TODO: use this in multiple delete, think of other usage
		// case k.Contains(k.Content.MultipleSelect, event.Name()):
		// 	return c.handleMultipleSelect(row)
		// case k.Contains(k.Content.ClearSelection, event.Name()):
		// 	return c.handleClearSelection()
		case k.Contains(k.Content.CopyLine, event.Name()):
			return c.handleCopyLine(row, coll)
		}

		return event
	})
}

// HandleDatabaseSelection is called when a database/collection is selected in the DatabaseTree
func (c *Content) HandleDatabaseSelection(ctx context.Context, db, coll string) error {
	c.queryBar.SetText("")
	c.sortBar.SetText("")

	state, ok := c.stateMap[db+"."+coll]
	if ok {
		c.state = state
	} else {
		state := mongo.CollectionState{
			Page: 0,
		}
		_, _, _, height := c.table.GetInnerRect()
		state.Limit = int64(height - 1)
		state.Db = db
		state.Coll = coll
		c.state = &state
	}

	err := c.updateContent(ctx, false)
	if err != nil {
		return err
	}
	c.App.SetFocus(c)
	return nil
}

// Rendering methods

func (c *Content) Render(setFocus bool) {
	c.Flex.Clear()
	c.tableFlex.Clear()

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

	c.tableFlex.AddItem(c.tableHeader, 2, 0, false)
	c.tableFlex.AddItem(c.table, 0, 1, true)

	c.Flex.AddItem(c.tableFlex, 0, 1, true)

	if setFocus {
		c.App.SetFocus(focusPrimitive)
	}
}

func (c *Content) renderTableView(startRow int, documents []primitive.M) {
	c.table.SetFixed(1, 0)
	sortedKeys := util.GetSortedKeysWithTypes(documents, c.style.ColumnTypeColor.Color().String())

	// Set the header row
	for col, key := range sortedKeys {
		c.table.SetCell(startRow, col, tview.NewTableCell(key).
			SetTextColor(c.style.ColumnKeyColor.Color()).
			SetSelectable(false).
			SetBackgroundColor(c.style.HeaderRowBackgroundColor.Color()).
			SetAlign(tview.AlignCenter))
	}
	startRow++

	// Populate the table with document values
	for row, doc := range documents {
		for col, key := range sortedKeys {
			var cellText string
			if val, ok := doc[strings.Split(key, " ")[0]]; ok {
				cellText = util.GetValueByType(val)
			} else {
				cellText = ""
			}
			if len(cellText) > 30 {
				cellText = cellText[0:30] + "..."
			}

			cell := tview.NewTableCell(cellText).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(30)

			// we'll set reference to _id for first column to not repeat the same _id in whole row
			if col == 0 {
				cell.SetReference(doc["_id"])
			}
			c.table.SetCell(startRow+row, col, cell)
		}
	}
}

func (c *Content) renderJsonView(startRow int, documents []primitive.M) {
	c.table.SetFixed(0, 0)
	row := startRow
	for _, doc := range documents {
		_id := doc["_id"]
		jsoned, err := mongo.ParseBsonDocument(doc)
		if err != nil {
			modal.ShowError(c.App.Pages, "Error stringifying document", err)
			return
		}
		c.jsonViewDocument(jsoned, &row, _id)
	}
	c.table.ScrollToBeginning()
}

func (c *Content) renderSingleRowView(startRow int, documents []primitive.M) {
	row := startRow
	for _, d := range documents {
		_id := d["_id"]
		jsoned, err := mongo.ParseBsonDocument(d)
		if err != nil {
			modal.ShowError(c.App.Pages, "Error stringifying document", err)
			return
		}
		dataCell := tview.NewTableCell(jsoned).
			SetAlign(tview.AlignLeft).
			SetReference(_id)

		c.table.SetCell(row, 0, dataCell)
		row++
	}
	c.table.ScrollToBeginning()
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

	documents, count, err := c.Dao.ListDocuments(ctx, c.state, filter, sort)
	if err != nil {
		return nil, 0, err
	}
	if len(documents) == 0 {
		return nil, 0, nil
	}

	c.state.Count = count
	c.state.PopulateDocs(documents)

	c.loadAutocompleteKeys(documents)

	return documents, count, nil
}

// loadAutocompleteKeys loads the autocomplete keys for the query and sort bars
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

func (c *Content) updateContent(ctx context.Context, useState bool) error {
	c.table.Clear()

	var documents []primitive.M
	var count int64

	if useState {
		documents = c.state.GetSortedDocs()
		count = c.state.Count
	} else {
		docs, c, err := c.listDocuments(ctx)
		if err != nil {
			return err
		}
		documents = docs
		count = c
	}

	headerInfo := fmt.Sprintf("Documents: %d, Page: %d, Limit: %d", count, c.state.Page, c.state.Limit)

	if c.state.Filter != "" {
		headerInfo += fmt.Sprintf(" | Filter: %s", c.state.Filter)
		c.queryBar.SetText(c.state.Filter)
	}
	if c.state.Sort != "" {
		headerInfo += fmt.Sprintf(" | Sort: %s", c.state.Sort)
		c.sortBar.SetText(c.state.Sort)
	}
	c.tableHeader.SetText(headerInfo)

	if count == 0 {
		// TODO: find why if selectable is set to false, program crashes
		c.table.SetCell(0, 0, tview.NewTableCell("No documents found"))
	}

	c.table.SetSelectable(true, c.currentView == TableView)
	startRow := 0
	switch c.currentView {
	case TableView:
		c.renderTableView(startRow, documents)
	case JsonView:
		c.renderJsonView(startRow, documents)
	case SingleLineView:
		c.renderSingleRowView(startRow, documents)
	}

	c.stateMap[c.state.Db+"."+c.state.Coll] = c.state

	return nil
}

func (c *Content) jsonViewDocument(doc string, row *int, _id interface{}) {
	indentedJson, err := mongo.IndentJson(doc)
	if err != nil {
		return
	}
	keyRegexWithIndent := regexp.MustCompile(`(?m)^\s{2}"([^"]+)":`)
	lines := strings.Split(indentedJson.String(), "\n")

	// we'll set reference of _id to first row of document, to not repeat the same _id in whole row
	c.table.SetCell(*row, 0, tview.
		NewTableCell(lines[0]).
		SetAlign(tview.AlignLeft).
		SetTextColor(tcell.ColorGreen).
		SetSelectable(false).
		SetReference(_id))

	*row++

	currLine := ""
	for i := 1; i < len(lines)-1; i++ {
		line := lines[i]
		if keyRegexWithIndent.MatchString(line) {
			if currLine != "" {
				c.table.SetCell(*row, 0, tview.NewTableCell(currLine).SetAlign(tview.AlignLeft))
				*row++
			}
			currLine = line
		} else {
			line = util.TrimMultipleSpaces(line)
			currLine += line
		}
	}

	if currLine != "" {
		c.table.SetCell(*row, 0, tview.NewTableCell(currLine).SetAlign(tview.AlignLeft))
		*row++
	}

	c.table.SetCell(*row, 0, tview.
		NewTableCell(lines[len(lines)-1]).
		SetAlign(tview.AlignLeft).
		SetTextColor(tcell.ColorGreen).
		SetSelectable(false).
		SetReference(_id))

	*row++
}

func (c *Content) queryBarListener(ctx context.Context) {
	acceptFunc := func(text string) {
		c.state.UpdateFilter(text)
		collectionKey := c.state.Db + "." + c.state.Coll
		c.stateMap[collectionKey] = c.state
		err := c.updateContent(ctx, false)
		if err != nil {
			modal.ShowError(c.App.Pages, "Error updating content", err)
			return
		}
		c.Flex.RemoveItem(c.queryBar)
		c.App.SetFocus(c.table)
	}
	rejectFunc := func() {
		c.Flex.RemoveItem(c.queryBar)
		c.App.SetFocus(c.table)
	}

	c.queryBar.DoneFuncHandler(acceptFunc, rejectFunc)
}

func (c *Content) sortBarListener(ctx context.Context) {
	acceptFunc := func(text string) {
		c.state.UpdateSort(text)
		collectionKey := c.state.Db + "." + c.state.Coll
		c.stateMap[collectionKey] = c.state
		c.updateContent(ctx, false)
		c.Flex.RemoveItem(c.sortBar)
		c.App.SetFocus(c.table)
	}
	rejectFunc := func() {
		c.Flex.RemoveItem(c.sortBar)
		c.App.SetFocus(c.table)
	}

	c.sortBar.DoneFuncHandler(acceptFunc, rejectFunc)
}

// refreshDocument refreshes the document in the table
func (c *Content) refreshDocument(ctx context.Context, doc string) {
	c.state.UpdateRawDoc(doc)
	c.updateContentBasedOnState(ctx)
}

func (c *Content) viewJson(jsonString string) error {
	c.view.Clear()

	c.App.Pages.AddPage(JsonViewComponent, c.view, true, true)

	indentedJson, err := mongo.IndentJson(jsonString)
	if err != nil {
		return err
	}

	c.view.SetText(indentedJson.String())
	c.view.ScrollToBeginning()

	c.view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			c.App.Pages.RemovePage(JsonViewComponent)
			c.App.SetFocus(c.table)
		}
		return event
	})

	return nil
}

func (c *Content) deleteDocument(ctx context.Context, jsonString string) error {
	objectId, err := mongo.GetIDFromJSON(jsonString)
	if err != nil {
		return err
	}

	stringifyId := mongo.StringifyId(objectId)

	c.deleteModal.SetText("Are you sure you want to delete document of id: [blue]" + stringifyId)
	c.deleteModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		defer c.App.Pages.RemovePage(c.deleteModal.GetIdentifier())
		if buttonLabel == "Cancel" {
			return
		}
		if buttonLabel == "Delete" {
			err = c.Dao.DeleteDocument(ctx, c.state.Db, c.state.Coll, objectId)
			if err != nil {
				modal.ShowError(c.App.Pages, "Error deleting document", err)
				return
			}
			c.state.DeleteDoc(objectId)
		}

		c.updateContentBasedOnState(ctx)

		row, col := c.table.GetSelection()
		if row == c.table.GetRowCount() {
			c.table.Select(row-1, col)
		} else {
			c.table.Select(row, col)
		}
	})

	c.App.Pages.AddPage(c.deleteModal.GetIdentifier(), c.deleteModal, true, true)

	return nil
}

func (c *Content) getDocumentBasedOnView(row, coll int) (string, error) {
	_id := c.getDocumentId(row, coll)
	return c.state.GetJsonDocById(_id)
}

// get document id based on view
func (c *Content) getDocumentId(row, coll int) interface{} {
	switch c.currentView {
	case JsonView:
		forWithReference := c.table.GetCellAboveThatMatch(row, coll, func(cell *tview.TableCell) bool {
			return strings.HasPrefix(cell.Text, `{`)
		})
		return forWithReference.GetReference()
	case TableView:
		return c.table.GetCell(row, 0).GetReference()
	case SingleLineView:
		return c.table.GetCell(row, 0).GetReference()
	default:
		return nil
	}
}

func (c *Content) handleScrolling(row int) {
	if row == 1 && c.currentView == JsonView {
		c.table.ScrollToBeginning()
	}
	if row == c.table.GetRowCount()-2 {
		c.table.ScrollToEnd()
	}
}

func (c *Content) handleSwitchView(ctx context.Context) *tcell.EventKey {
	switch c.currentView {
	case TableView:
		c.currentView = JsonView
	case JsonView:
		c.currentView = SingleLineView
	case SingleLineView:
		c.currentView = TableView
	}
	c.updateContent(ctx, true)
	return nil
}

func (c *Content) handlePeekDocument(ctx context.Context, row, coll int) *tcell.EventKey {
	_id := c.getDocumentId(row, coll)
	if _id == nil {
		return nil
	}
	c.peeker.Render(ctx, c.state, _id)
	return nil
}

func (c *Content) handleViewDocument(row, coll int) *tcell.EventKey {
	doc, err := c.getDocumentBasedOnView(row, coll)
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
	id, err := c.docModifier.Insert(ctx, c.state.Db, c.state.Coll)
	if err != nil {
		modal.ShowError(c.App.Pages, "Error adding document", err)
		return nil
	}
	insertedDoc, err := c.Dao.GetDocument(ctx, c.state.Db, c.state.Coll, id)
	if err != nil {
		modal.ShowError(c.App.Pages, "Error getting inserted document", err)
		return nil
	}
	c.state.AppendDoc(insertedDoc)
	c.updateContentBasedOnState(ctx)
	return nil
}

func (c *Content) handleEditDocument(ctx context.Context, row, coll int) *tcell.EventKey {
	_id := c.getDocumentId(row, coll)
	doc, err := c.state.GetJsonDocById(_id)
	if err != nil {
		modal.ShowError(c.App.Pages, "Error getting document", err)
		return nil
	}
	updated, err := c.docModifier.Edit(ctx, c.state.Db, c.state.Coll, _id, doc)
	if err != nil {
		modal.ShowError(c.App.Pages, "Error editing document", err)
		return nil
	}

	if updated != "" {
		c.refreshDocument(ctx, updated)
	}
	return nil
}

func (c *Content) handleDuplicateDocument(ctx context.Context, row, coll int) *tcell.EventKey {
	doc, err := c.getDocumentBasedOnView(row, coll)
	if err != nil {
		modal.ShowError(c.App.Pages, "Error duplicating document", err)
		return nil
	}
	id, err := c.docModifier.Duplicate(ctx, c.state.Db, c.state.Coll, doc)
	if err != nil {
		modal.ShowError(c.App.Pages, "Error duplicating document", err)
		return nil
	}
	duplicatedDoc, err := c.Dao.GetDocument(ctx, c.state.Db, c.state.Coll, id)
	if err != nil {
		modal.ShowError(c.App.Pages, "Error getting inserted document", err)
		return nil
	}
	c.state.AppendDoc(duplicatedDoc)
	c.updateContentBasedOnState(ctx)
	return nil
}

func (c *Content) handleToggleQuery() *tcell.EventKey {
	if c.state.Filter != "" {
		c.queryBar.Toggle(c.state.Filter)
	} else {
		c.queryBar.Toggle("")
	}
	c.Render(true)
	return nil
}

func (c *Content) handleToggleSort() *tcell.EventKey {
	if c.state.Sort != "" {
		c.sortBar.Toggle(c.state.Sort)
	} else {
		c.sortBar.Toggle("")
	}
	c.Render(true)
	return nil
}

func (c *Content) handleDeleteDocument(ctx context.Context, row, coll int) *tcell.EventKey {
	doc, err := c.getDocumentBasedOnView(row, coll)
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
	err := c.updateContent(ctx, false)
	if err != nil {
		defer modal.ShowError(c.App.Pages, "Error refreshing documents", err)
	}
	return nil
}

func (c *Content) handleNextDocument(row, col int) *tcell.EventKey {
	if c.currentView == JsonView {
		c.table.MoveDownUntil(row, col, func(cell *tview.TableCell) bool {
			return strings.HasPrefix(cell.Text, `{`)
		})
	} else {
		c.table.MoveDown()
	}
	return nil
}

func (c *Content) handlePreviousDocument(row, col int) *tcell.EventKey {
	if c.currentView == JsonView {
		c.table.MoveUpUntil(row, col, func(cell *tview.TableCell) bool {
			return strings.HasPrefix(cell.Text, `}`)
		})
	} else {
		c.table.MoveUp()
	}
	return nil
}

func (c *Content) handleNextPage(ctx context.Context) *tcell.EventKey {
	if c.state.Page+c.state.Limit >= c.state.Count {
		return nil
	}
	c.state.Page += c.state.Limit
	collectionKey := c.state.Db + "." + c.state.Coll
	c.stateMap[collectionKey] = c.state
	c.updateContent(ctx, false)
	return nil
}

func (c *Content) handlePreviousPage(ctx context.Context) *tcell.EventKey {
	if c.state.Page == 0 {
		return nil
	}
	c.state.Page -= c.state.Limit
	collectionKey := c.state.Db + "." + c.state.Coll
	c.stateMap[collectionKey] = c.state
	c.updateContent(ctx, false)
	return nil
}

// func (c *Content) handleMultipleSelect(row int) *tcell.EventKey {
// 	c.table.ToggleRowSelection(row)
// 	return nil
// }

func (c *Content) handleClearSelection() *tcell.EventKey {
	c.table.ClearSelection()
	return nil
}

func (c *Content) handleCopyLine(row, col int) *tcell.EventKey {
	selectedDoc := util.CleanJsonWhitespaces(c.table.GetCell(row, col).Text)
	err := clipboard.WriteAll(selectedDoc)
	if err != nil {
		modal.ShowError(c.App.Pages, "Error copying document", err)
	}
	return nil
}

func (c *Content) updateContentBasedOnState(ctx context.Context) error {
	if c.state.Filter != "" || c.state.Sort != "" {
		return c.updateContent(ctx, false)
	} else {
		return c.updateContent(ctx, true)
	}
}
