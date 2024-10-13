package component

import (
	"context"
	"fmt"
	"strings"

	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
	"github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
)

const (
	IndexId = "Index"
)

type Index struct {
	*core.BaseElement
	*core.Table

	indexes []mongo.IndexInfo
}

func NewIndex() *Index {
	i := &Index{
		BaseElement: core.NewBaseElement(),
		Table:       core.NewTable(),
	}

	i.SetIdentifier(IndexId)
	i.SetAfterInitFunc(i.init)

	return i
}

func (i *Index) init() error {
	i.setStyle()
	i.setStaticLayour()

	i.handleEvents()
	return nil
}

func (i *Index) setStyle() {
	globalStyle := i.App.GetStyles()
	i.SetStyle(globalStyle)

	i.SetSeparator(globalStyle.Others.SeparatorSymbol.Rune())
	i.SetBordersColor(globalStyle.Others.SeparatorColor.Color())
}

func (i *Index) setStaticLayour() {
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
		}
	})
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
		i.SetCell(row+1, 3, tview.NewTableCell(fmt.Sprintf("%.1f KB", float64(index.Size)/1024)))
		i.SetCell(row+1, 4, tview.NewTableCell(index.Usage))
		i.SetCell(row+1, 5, tview.NewTableCell(strings.Join(index.Properties, ", ")))
	}

}

func (i *Index) HandleDatabaseSelection(ctx context.Context, db, coll string) error {
	indexes, err := i.Dao.GetIndexes(ctx, db, coll)
	if err != nil {
		return err
	}

	i.indexes = indexes
	i.Render()

	return nil
}
