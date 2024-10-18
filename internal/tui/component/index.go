package component

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
	int_mongo "github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/tui/modal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	IndexId            = "Index"
	IndexAddFormId     = "IndexAddForm"
	IndexDeleteModalId = "IndexDeleteModal"
)

type Index struct {
	*core.BaseElement
	*core.Flex

	table            *core.Table
	addForm          *core.Form
	indexes          []int_mongo.IndexInfo
	deleteModal      *modal.Delete
	currentDB        string
	currentColl      string
	docKeys          []string
	isAddFormVisible bool
}

func NewIndex() *Index {
	i := &Index{
		BaseElement:      core.NewBaseElement(),
		Flex:             core.NewFlex(),
		table:            core.NewTable(),
		addForm:          core.NewForm(),
		deleteModal:      modal.NewDeleteModal(IndexDeleteModalId),
		isAddFormVisible: false,
	}

	i.SetIdentifier(IndexId)
	i.SetAfterInitFunc(i.init)

	return i
}
func (i *Index) init() error {
	i.setLayout()
	i.setStyle()
	i.setKeybindings()

	if err := i.deleteModal.Init(i.App); err != nil {
		return err
	}

	i.handleEvents()

	return nil
}

func (i *Index) setStyle() {
	globalStyle := i.App.GetStyles()
	i.SetStyle(globalStyle)
	i.table.SetStyle(globalStyle)
	i.addForm.SetStyle(globalStyle)

	i.table.SetSeparator(globalStyle.Others.SeparatorSymbol.Rune())
	i.table.SetBordersColor(globalStyle.Others.SeparatorColor.Color())
}

func (i *Index) setLayout() {
	i.SetBorder(true)
	i.SetTitle(" Indexes ")
	i.SetTitleAlign(tview.AlignCenter)
	i.SetBorderPadding(0, 0, 1, 1)
	i.table.SetSelectable(true, true)
}

func (i *Index) handleEvents() {
	go i.HandleEvents(IndexId, func(event manager.EventMsg) {
		switch event.Message.Type {
		case manager.StyleChanged:
			i.setStyle()
			i.Render()
		case manager.UpdateAutocompleteKeys:
			i.docKeys = event.Message.Data.([]string)
		}
	})
}

func (i *Index) setKeybindings() {
	k := i.App.GetKeys()
	i.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case k.Contains(k.Index.ExitAddIndex, event.Name()):
			if i.isAddFormVisible {
				i.closeAddForm()
				return nil
			}
		case k.Contains(k.Index.AddIndex, event.Name()):
			if !i.isAddFormVisible {
				i.addIndexForm()
				return nil
			}
		case k.Contains(k.Index.DeleteIndex, event.Name()):
			i.showDeleteIndexModal()
			return nil
		}
		return event
	})
}

func (i *Index) HandleDatabaseSelection(ctx context.Context, db, coll string) error {
	i.currentDB = db
	i.currentColl = coll
	return i.refreshIndexes(ctx)
}

func (i *Index) refreshIndexes(ctx context.Context) error {
	indexes, err := i.Dao.GetIndexes(ctx, i.currentDB, i.currentColl)
	if err != nil {
		return err
	}
	i.indexes = indexes
	i.Render()
	return nil
}

func (i *Index) Render() {
	i.Flex.Clear()

	if i.isAddFormVisible {
		i.addForm.Clear(true)
		i.renderAddIndexForm()
		i.Flex.AddItem(i.addForm, 0, 1, true)
	} else {
		i.table.Clear()
		i.renderIndexTable()
		i.Flex.AddItem(i.table, 0, 1, true)
	}
}

func (i *Index) renderIndexTable() {
	headers := []string{"Name", "Definition", "Type", "Size", "Usage", "Properties"}
	for col, header := range headers {
		cell := tview.NewTableCell(" " + header + " ").SetSelectable(false).SetAlign(tview.AlignCenter)
		cell.SetTextColor(i.App.GetStyles().Content.ColumnKeyColor.Color())
		cell.SetBackgroundColor(i.App.GetStyles().Content.HeaderRowBackgroundColor.Color())
		i.table.SetCell(0, col, cell)
	}

	for row, index := range i.indexes {
		var definition string
		for key, value := range index.Definition {
			definition += fmt.Sprintf("%s: %v ", key, value)
		}
		i.table.SetCell(row+1, 0, tview.NewTableCell(index.Name))
		i.table.SetCell(row+1, 1, tview.NewTableCell(definition))
		i.table.SetCell(row+1, 2, tview.NewTableCell(index.Type))
		i.table.SetCell(row+1, 3, tview.NewTableCell(index.Size))
		i.table.SetCell(row+1, 4, tview.NewTableCell(index.Usage))
		i.table.SetCell(row+1, 5, tview.NewTableCell(strings.Join(index.Properties, ", ")))
	}
}

func (i *Index) renderAddIndexForm() {
	i.addForm.SetTitle("Add Index")
	i.addForm.AddInputFieldWithAutocomplete("Field to index", "", 30, i.setAutocompleteFunc, nil, nil)
	i.addForm.AddDropDown("Field Type", []string{"1 (Ascending)", "-1 (Descending)", "text", "2dsphere"}, 0, nil)
	i.addForm.AddTextView("Optionals", "----------------", 40, 1, false, false)
	i.addForm.AddInputField("Index Name", "", 30, nil, nil)
	i.addForm.AddCheckbox("Unique", false, nil)
	i.addForm.AddInputField("TTL (seconds)", "", 20, nil, nil)
	i.addForm.AddButton("Create", i.handleAddIndex)
	i.addForm.AddButton("Cancel", i.closeAddForm)
}

func (i *Index) setAutocompleteFunc(currentText string) (entries []tview.AutocompleteItem) {
	sortedEntries := make([]tview.AutocompleteItem, 0, len(i.docKeys))
	if i.docKeys != nil {
		for _, keyword := range i.docKeys {
			if matched, _ := regexp.MatchString("(?i)^"+currentText, keyword); matched {
				sortedEntries = append(sortedEntries, tview.AutocompleteItem{Main: keyword})
			}
		}
	}

	sort.SliceStable(sortedEntries, func(i, j int) bool {
		return strings.ToLower(sortedEntries[i].Main) < strings.ToLower(sortedEntries[j].Main)
	})

	return sortedEntries
}

func (i *Index) handleAddIndex() {
	fieldName := i.addForm.GetFormItem(0).(*tview.InputField).GetText()
	fieldType, _ := i.addForm.GetFormItem(1).(*tview.DropDown).GetCurrentOption()
	indexName := i.addForm.GetFormItem(3).(*tview.InputField).GetText()
	unique := i.addForm.GetFormItem(4).(*tview.Checkbox).IsChecked()
	ttlStr := i.addForm.GetFormItem(5).(*tview.InputField).GetText()

	var fieldValue interface{}
	switch fieldType {
	case 0:
		fieldValue = 1
	case 1:
		fieldValue = -1
	case 2:
		fieldValue = "text"
	case 3:
		fieldValue = "2dsphere"
	}

	options := options.Index()
	if unique {
		options.SetUnique(unique)
	}
	if indexName != "" {
		options.SetName(indexName)
	}
	if ttlStr != "" {
		ttl, err := strconv.Atoi(ttlStr)
		if err != nil {
			modal.ShowError(i.App.Pages, "Invalid TTL value", err)
			return
		}
		options.SetExpireAfterSeconds(int32(ttl))
	}

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: fieldName, Value: fieldValue}},
		Options: options,
	}

	ctx := context.Background()
	if err := i.Dao.CreateIndex(ctx, i.currentDB, i.currentColl, indexModel); err != nil {
		modal.ShowError(i.App.Pages, "Error creating index", err)
		return
	}

	i.closeAddForm()
	err := i.refreshIndexes(ctx)
	if err != nil {
		modal.ShowError(i.App.Pages, "Error refreshing indexes", err)
	}
}

func (i *Index) addIndexForm() {
	if i.isAddFormVisible {
		return
	}

	i.table.Clear()
	i.isAddFormVisible = true

	i.Render()
	i.App.SetFocus(i.addForm)
}

func (i *Index) closeAddForm() {
	i.addForm.Clear(true)
	i.isAddFormVisible = false

	i.Render()
	i.App.SetFocus(i)
}

func (i *Index) showDeleteIndexModal() {
	if i.table.GetCell(i.table.GetSelection()).Text == "" {
		return
	}
	row, _ := i.table.GetSelection()
	indexName := i.table.GetCell(row, 0).Text

	i.deleteModal.SetText(fmt.Sprintf("Are you sure you want to delete index [%s]%s[-:-:-]?", i.App.GetStyles().Content.ColumnKeyColor.Color(), indexName))
	i.deleteModal.SetDoneFunc(i.createDeleteIndexDoneFunc(indexName, row))
	i.App.Pages.AddPage(IndexDeleteModalId, i.deleteModal, true, true)
}

func (i *Index) createDeleteIndexDoneFunc(indexName string, row int) func(int, string) {
	return func(buttonIndex int, buttonLabel string) {
		defer i.App.Pages.RemovePage(IndexDeleteModalId)
		if buttonIndex == 0 {
			i.handleDeleteIndex(indexName)
		}
		i.table.RemoveRow(row)
		i.table.Select(row-1, 0)
	}
}

func (i *Index) handleDeleteIndex(indexName string) {
	ctx := context.Background()
	if err := i.Dao.DropIndex(ctx, i.currentDB, i.currentColl, indexName); err != nil {
		modal.ShowError(i.App.Pages, "Error deleting index", err)
		return
	}
}

func (i *Index) IsAddFormFocused() bool {
	return i.isAddFormVisible
}
