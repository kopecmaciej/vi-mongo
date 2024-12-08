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
	ContentId            = "Content"
	JsonViewId           = "JsonView"
	QueryBarId           = "QueryBar"
	SortBarId            = "SortBar"
	ContentDeleteModalId = "ContentDeleteModal"
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
	style       *config.ContentStyle
	queryBar    *InputBar
	sortBar     *InputBar
	peeker      *Peeker
	deleteModal *modal.Delete
	docModifier *DocModifier
	state       *mongo.CollectionState
	stateMap    *mongo.StateMap
	currentView ViewType
}

func NewContent() *Content {
	c := &Content{
		BaseElement: core.NewBaseElement(),
		Flex:        core.NewFlex(),

		tableFlex:   core.NewFlex(),
		tableHeader: core.NewTextView(),
		table:       core.NewTable(),
		queryBar:    NewInputBar(QueryBarId, "Query"),
		sortBar:     NewInputBar(SortBarId, "Sort"),
		peeker:      NewPeeker(),
		deleteModal: modal.NewDeleteModal(ContentDeleteModalId),
		docModifier: NewDocModifier(),
		state:       &mongo.CollectionState{},
		stateMap:    mongo.NewStateMap(),
		currentView: TableView,
	}

	c.SetIdentifier(ContentId)
	// neccesarry if focus is get back to content component
	// it's related to how tview package works
	c.table.SetIdentifier(ContentId)
	c.SetAfterInitFunc(c.init)

	return c
}

func (c *Content) init() error {
	ctx := context.Background()

	c.setLayout()
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

	c.queryBarHandler(ctx)
	c.sortBarHandler(ctx)

	c.peeker.SetDoneFunc(func() {
		c.updateContent(ctx, true)
	})

	c.handleEvents(ctx)

	return nil
}

func (c *Content) handleEvents(ctx context.Context) {
	go c.HandleEvents(ContentId, func(event manager.EventMsg) {
		switch event.Message.Type {
		case manager.StyleChanged:
			c.setStyle()
			c.updateContent(ctx, true)
		case manager.UpdateQueryBar:
			query, ok := event.Message.Data.(string)
			if !ok {
				modal.ShowError(c.App.Pages, "Invalid query", nil)
				return
			}
			go c.App.QueueUpdateDraw(func() {
				c.applyQuery(ctx, query)
				c.App.SetFocus(c)
			})
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
	c.Flex.SetStyle(styles)
	c.table.SetStyle(styles)

	c.tableFlex.SetBorderColor(styles.Others.SeparatorColor.Color())
	c.tableHeader.SetTextColor(c.style.StatusTextColor.Color())

	c.table.SetBordersColor(styles.Others.SeparatorColor.Color())
	c.table.SetSeparator(styles.Others.SeparatorSymbol.Rune())
}

func (c *Content) setLayout() {
	c.tableFlex.SetBorder(true)
	c.tableFlex.SetDirection(tview.FlexRow)
	c.tableFlex.SetTitle(" Content ")
	c.tableFlex.SetTitleAlign(tview.AlignCenter)
	c.tableFlex.SetBorderPadding(0, 0, 1, 1)

	c.tableHeader.SetText("Documents: 0, Page: 0, Limit: 0")

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
		case k.Contains(k.Content.FullPagePeek, event.Name()):
			return c.handleFullPagePeek(ctx, row, coll)
		case k.Contains(k.Content.AddDocument, event.Name()):
			return c.handleAddDocument(ctx)
		case k.Contains(k.Content.EditDocument, event.Name()):
			return c.handleEditDocument(ctx, row, coll)
		case k.Contains(k.Content.DuplicateDocument, event.Name()):
			return c.handleDuplicateDocument(ctx, row, coll)
		case k.Contains(k.Content.DeleteDocument, event.Name()):
			return c.handleDeleteDocument(ctx, row, coll)
		case k.Contains(k.Content.ToggleQueryBar, event.Name()):
			return c.handleToggleQuery()
		case k.Contains(k.Content.ToggleSortBar, event.Name()):
			return c.handleToggleSort()
		case k.Contains(k.Content.SortByColumn, event.Name()):
			return c.handleSortByColumn(ctx, coll)
		case k.Contains(k.Content.HideColumn, event.Name()):
			return c.handleHideColumn(ctx, coll)
		case k.Contains(k.Content.ResetHiddenColumns, event.Name()):
			return c.handleResetHiddenColumns(ctx)
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
		case k.Contains(k.Content.CopyHighlight, event.Name()):
			return c.handleCopyLine(row, coll)
		case k.Contains(k.Content.CopyDocument, event.Name()):
			return c.handleCopyDocument(row, coll)
		}

		return event
	})
}

// HandleDatabaseSelection is called when a database/collection is selected in the DatabaseTree
func (c *Content) HandleDatabaseSelection(ctx context.Context, db, coll string) error {
	c.queryBar.SetText("")
	c.sortBar.SetText("")

	state, ok := c.stateMap.Get(c.stateMap.Key(db, coll))
	if ok {
		c.state = state
	} else {
		c.state = mongo.NewCollectionState()
		c.state.Db = db
		c.state.Coll = coll
		_, _, _, height := c.table.GetInnerRect()
		c.state.Limit = int64(height - 1)
	}

	err := c.updateContent(ctx, false)
	if err != nil {
		return err
	}

	c.App.SetFocus(c)
	return nil
}

func (c *Content) Render() {
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

	c.App.SetFocus(focusPrimitive)
}

func (c *Content) renderTableView(startRow int, documents []primitive.M) {
	c.table.SetFixed(1, 0)
	allHeaderKeys := util.GetSortedKeysWithTypes(documents, c.style.ColumnTypeColor.Color().String())

	// Filter out hidden columns
	var sortedHeaderKeys []string
	hiddenCols := c.stateMap.GetHiddenColumns(c.state.Db, c.state.Coll)
	for _, key := range allHeaderKeys {
		columnName := strings.Split(key, " ")[0]
		if !contains(hiddenCols, columnName) {
			sortedHeaderKeys = append(sortedHeaderKeys, key)
		}
	}

	// Set the header row
	for col, key := range sortedHeaderKeys {
		c.table.SetCell(startRow, col, tview.NewTableCell(key).
			SetTextColor(c.style.ColumnKeyColor.Color()).
			SetSelectable(false).
			SetBackgroundColor(c.style.HeaderRowBackgroundColor.Color()).
			SetAlign(tview.AlignCenter))
	}
	startRow++

	// Populate the table with document values
	for row, doc := range documents {
		for col, key := range sortedHeaderKeys {
			var cellText string
			if val, ok := doc[strings.Split(key, " ")[0]]; ok {
				cellText = util.StringifyMongoValueByType(val)
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
	c.table.Select(1, 0)
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
	c.table.Select(1, 0)
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
	c.table.Select(0, 0)
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
	c.App.GetManager().Broadcast(manager.EventMsg{
		Sender:  c.GetIdentifier(),
		Message: manager.Message{Type: manager.UpdateAutocompleteKeys, Data: autocompleteKeys},
	})
}

func (c *Content) updateContent(ctx context.Context, useState bool) error {
	var documents []primitive.M
	var count int64

	if useState {
		documents = c.state.GetAllDocs()
		count = c.state.Count
	} else {
		docs, c, err := c.listDocuments(ctx)
		if err != nil {
			return err
		}
		documents = docs
		count = c
	}

	c.table.Clear()

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

	c.stateMap.Set(c.stateMap.Key(c.state.Db, c.state.Coll), c.state)

	if count == 0 {
		c.table.SetCell(0, 0, tview.NewTableCell("No documents found"))
		return nil
	}

	c.renderView(documents)
	return nil
}

func (c *Content) renderView(documents []primitive.M) {
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

func (c *Content) applyQuery(ctx context.Context, query string) error {
	c.state.SetFilter(query)
	err := c.updateContent(ctx, false)
	if err != nil {
		c.state.SetFilter("")
		return err
	}
	if query != "" && strings.ReplaceAll(query, " ", "") != "{}" {
		err = c.queryBar.historyModal.SaveToHistory(query)
		if err != nil {
			return err
		}
	}
	c.stateMap.Set(c.stateMap.Key(c.state.Db, c.state.Coll), c.state)
	return nil
}

func (c *Content) applySort(ctx context.Context, sort string) error {
	c.state.SetSort(sort)
	err := c.updateContent(ctx, false)
	if err != nil {
		c.state.SetSort("")
		return err
	}

	c.stateMap.Set(c.stateMap.Key(c.state.Db, c.state.Coll), c.state)
	return nil
}

func (c *Content) queryBarHandler(ctx context.Context) {
	acceptFunc := func(text string) {
		err := c.applyQuery(ctx, text)
		if err != nil {
			modal.ShowError(c.App.Pages, "Error applying query", err)
		} else {
			c.Flex.RemoveItem(c.queryBar)
			c.App.SetFocus(c.table)
		}
	}
	rejectFunc := func() {
		c.Flex.RemoveItem(c.queryBar)
		c.App.SetFocus(c.table)
	}

	c.queryBar.DoneFuncHandler(acceptFunc, rejectFunc)
}

func (c *Content) sortBarHandler(ctx context.Context) {
	acceptFunc := func(text string) {
		err := c.applySort(ctx, text)
		if err != nil {
			modal.ShowError(c.App.Pages, "Error applying sort", err)
		} else {
			c.Flex.RemoveItem(c.sortBar)
			c.App.SetFocus(c.table)
		}
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
	c.peeker.SetFullScreen(false)
	c.peeker.Render(ctx, c.state, _id)
	return nil
}

func (c *Content) handleFullPagePeek(ctx context.Context, row, coll int) *tcell.EventKey {
	_id := c.getDocumentId(row, coll)
	if _id == nil {
		return nil
	}
	c.peeker.SetFullScreen(true)
	c.peeker.Render(ctx, c.state, _id)
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
	c.Render()
	return nil
}

func (c *Content) handleToggleSort() *tcell.EventKey {
	if c.state.Sort != "" {
		c.sortBar.Toggle(c.state.Sort)
	} else {
		c.sortBar.Toggle("")
	}
	c.Render()
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
			return strings.HasPrefix(strings.TrimSpace(cell.Text), `"_id"`)
		})
	} else {
		c.table.MoveDown()
	}
	return nil
}

func (c *Content) handlePreviousDocument(row, col int) *tcell.EventKey {
	if c.currentView == JsonView {
		c.table.MoveUpUntil(row, col, func(cell *tview.TableCell) bool {
			return strings.HasPrefix(strings.TrimSpace(cell.Text), `"_id"`)
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
	c.stateMap.Set(c.stateMap.Key(c.state.Db, c.state.Coll), c.state)
	c.updateContent(ctx, false)
	return nil
}

func (c *Content) handlePreviousPage(ctx context.Context) *tcell.EventKey {
	if c.state.Page == 0 {
		return nil
	}
	c.state.Page -= c.state.Limit
	c.stateMap.Set(c.stateMap.Key(c.state.Db, c.state.Coll), c.state)
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

func (c *Content) handleCopyDocument(row, col int) *tcell.EventKey {
	docId := c.getDocumentId(row, col)
	doc, err := c.state.GetJsonDocById(docId)
	if err != nil {
		modal.ShowError(c.App.Pages, "Error copying document", err)
		return nil
	}
	err = clipboard.WriteAll(doc)
	if err != nil {
		modal.ShowError(c.App.Pages, "Error copying document", err)
	}
	return nil
}

// Automatic sort (1 or -1) for given column, only in TableView
func (c *Content) handleSortByColumn(ctx context.Context, col int) *tcell.EventKey {
	if c.currentView != TableView {
		return nil
	}

	headerCell := c.table.GetCell(0, col)
	if headerCell == nil {
		return nil
	}

	columnName := strings.Split(headerCell.Text, " ")[0]
	currentSort := c.state.Sort

	var newSort string
	if currentSort == fmt.Sprintf(`{ "%s": 1 }`, columnName) {
		newSort = fmt.Sprintf(`{ "%s": -1 }`, columnName)
	} else {
		newSort = fmt.Sprintf(`{ "%s": 1 }`, columnName)
	}

	err := c.applySort(ctx, newSort)
	if err != nil {
		modal.ShowError(c.App.Pages, "Error applying sort", err)
	}

	c.table.Select(1, col)

	return nil
}

func (c *Content) handleHideColumn(ctx context.Context, col int) *tcell.EventKey {
	if c.currentView != TableView {
		return nil
	}

	headerCell := c.table.GetCell(0, col)
	if headerCell == nil {
		return nil
	}

	columnName := strings.Split(headerCell.Text, " ")[0]
	c.stateMap.AddHiddenColumn(c.state.Db, c.state.Coll, columnName)

	c.updateContent(ctx, true)
	return nil
}

func (c *Content) handleResetHiddenColumns(ctx context.Context) *tcell.EventKey {
	if c.currentView != TableView {
		return nil
	}

	c.stateMap.ResetHiddenColumns(c.state.Db, c.state.Coll)
	c.updateContent(ctx, true)
	return nil
}

func (c *Content) updateContentBasedOnState(ctx context.Context) error {
	useState := c.state.Filter == "" && c.state.Sort == ""
	return c.updateContent(ctx, useState)
}
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
