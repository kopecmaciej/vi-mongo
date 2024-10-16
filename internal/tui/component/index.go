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
	IndexAddModalId    = "IndexAddModal"
	IndexDeleteModalId = "IndexDeleteModal"
)

type IndexField struct {
	Name string
	Type string
}

type Index struct {
	*core.BaseElement
	*core.Table

	indexes     []int_mongo.IndexInfo
	addModal    *core.Form
	deleteModal *modal.Delete
	currentDB   string
	currentColl string
	docKeys     []string
	indexFields []IndexField
}

func NewIndex() *Index {
	i := &Index{
		BaseElement: core.NewBaseElement(),
		Table:       core.NewTable(),
		addModal:    core.NewForm(),
		deleteModal: modal.NewDeleteModal(IndexDeleteModalId),
		indexFields: []IndexField{},
	}

	i.SetIdentifier(IndexId)
	i.SetAfterInitFunc(i.init)

	return i
}
func (i *Index) init() error {
	i.setStyle()
	i.setStaticLayout()

	i.handleEvents()
	i.setKeybindings()

	if err := i.deleteModal.Init(i.App); err != nil {
		return err
	}

	return nil
}

func (i *Index) setStyle() {
	globalStyle := i.App.GetStyles()
	i.SetStyle(globalStyle)

	i.SetSeparator(globalStyle.Others.SeparatorSymbol.Rune())
	i.SetBordersColor(globalStyle.Others.SeparatorColor.Color())
}

func (i *Index) setStaticLayout() {
	i.SetBorder(true)
	i.SetTitle(" Indexes ")
	i.SetTitleAlign(tview.AlignCenter)
	i.SetBorderPadding(0, 0, 1, 1)
	i.SetSelectable(true, true)
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
		case k.Contains(k.Index.ExitModal, event.Name()):
			i.closeAddModal()
			return nil
		case k.Contains(k.Index.AddIndex, event.Name()):
			i.showAddIndexModal()
			return nil
		case k.Contains(k.Index.DeleteIndex, event.Name()):
			i.showDeleteIndexModal()
			return nil
		}
		return event
	})
}

func (i *Index) showAddIndexModal() {
	i.setupAddIndexForm()
	i.App.Pages.AddPage(IndexAddModalId, i.addModal, true, true)
}

func (i *Index) setupAddIndexForm() {
	i.addModal.SetBorder(true)
	i.addModal.SetTitle("Add Index")
	i.addModal.AddInputFieldWithAutocomplete("Field to index", "", 30, i.setAutocompleteFunc, nil, nil)
	i.addModal.AddDropDown("Field Type", []string{"1 (Ascending)", "-1 (Descending)", "text", "2dsphere"}, 0, nil)
	i.addModal.AddTextView("Optionals", "----------------", 40, 1, false, false)
	i.addModal.AddInputField("Index Name", "", 30, nil, nil)
	i.addModal.AddCheckbox("Unique", false, nil)
	i.addModal.AddInputField("TTL (seconds)", "", 20, nil, nil)
	i.addModal.AddButton("Create", i.handleAddIndex)
	i.addModal.AddButton("Cancel", func() {
		i.closeAddModal()
	})
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
	fieldName := i.addModal.GetFormItem(0).(*tview.InputField).GetText()
	fieldType, _ := i.addModal.GetFormItem(1).(*tview.DropDown).GetCurrentOption()
	indexName := i.addModal.GetFormItem(3).(*tview.InputField).GetText()
	unique := i.addModal.GetFormItem(4).(*tview.Checkbox).IsChecked()
	ttlStr := i.addModal.GetFormItem(5).(*tview.InputField).GetText()

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
	} else if indexName != "" {
		options.SetName(indexName)
	} else if ttlStr != "" {
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

	i.closeAddModal()
	err := i.refreshIndexes(ctx)
	if err != nil {
		modal.ShowError(i.App.Pages, "Error refreshing indexes", err)
	}
}

func (i *Index) closeAddModal() {
	i.addModal.Clear(true)
	i.App.Pages.RemovePage(IndexAddModalId)
}

func (i *Index) showDeleteIndexModal() {
	if i.GetCell(i.GetSelection()).Text == "" {
		return
	}
	row, _ := i.GetSelection()
	indexName := i.GetCell(row, 0).Text

	i.deleteModal.SetText(fmt.Sprintf("Are you sure you want to delete index [%s]%s[-:-:-]?", i.App.GetStyles().Content.ColumnKeyColor.Color(), indexName))
	i.deleteModal.SetDoneFunc(i.createDeleteIndexDoneFunc(indexName, row))
	i.App.Pages.AddPage(IndexDeleteModalId, i.deleteModal, true, true)
}

func (i *Index) createDeleteIndexDoneFunc(indexName string, row int) func(int, string) {
	return func(buttonIndex int, buttonLabel string) {
		defer i.App.Pages.RemovePage(IndexDeleteModalId)
		if buttonIndex == 0 {
			i.handleDeleteIndex(indexName, row)
		}
		i.Table.RemoveRow(row)
		i.Table.Select(row-1, 0)
	}
}

func (i *Index) handleDeleteIndex(indexName string, row int) {
	ctx := context.Background()
	if err := i.Dao.DropIndex(ctx, i.currentDB, i.currentColl, indexName); err != nil {
		modal.ShowError(i.App.Pages, "Error deleting index", err)
		return
	}
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

func (i *Index) HandleDatabaseSelection(ctx context.Context, db, coll string) error {
	i.currentDB = db
	i.currentColl = coll
	return i.refreshIndexes(ctx)
}

func (i *Index) Render() {
	i.Clear()

	headers := []string{"Name", "Definition", "Type", "Size", "Usage", "Properties"}
	for col, header := range headers {
		cell := tview.NewTableCell(" " + header + " ").SetSelectable(false).SetAlign(tview.AlignCenter)
		cell.SetTextColor(i.App.GetStyles().Content.ColumnKeyColor.Color())
		cell.SetBackgroundColor(i.App.GetStyles().Content.HeaderRowBackgroundColor.Color())
		i.SetCell(0, col, cell)
	}

	for row, index := range i.indexes {
		var definition string
		for key, value := range index.Definition {
			definition += fmt.Sprintf("%s: %v ", key, value)
		}
		i.SetCell(row+1, 0, tview.NewTableCell(index.Name))
		i.SetCell(row+1, 1, tview.NewTableCell(definition))
		i.SetCell(row+1, 2, tview.NewTableCell(index.Type))
		i.SetCell(row+1, 3, tview.NewTableCell(index.Size))
		i.SetCell(row+1, 4, tview.NewTableCell(index.Usage))
		i.SetCell(row+1, 5, tview.NewTableCell(strings.Join(index.Properties, ", ")))
	}
}
