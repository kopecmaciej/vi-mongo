package view

import (
	"slices"
	"strings"

	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TableColumns renders MongoDB documents as a spreadsheet-style table with a header row.
type TableColumns struct {
	Style      *config.ContentStyle
	HiddenCols []string
}

func NewTableColumns(style *config.ContentStyle) *TableColumns {
	return &TableColumns{Style: style}
}

func (t *TableColumns) Render(table *core.Table, startRow int, documents []primitive.M) {
	table.SetFixed(1, 0)
	allHeaderKeys := util.GetSortedKeysWithTypes(documents, t.Style.ColumnTypeColor.Color().String())

	var sortedHeaderKeys []string
	for _, key := range allHeaderKeys {
		columnName := strings.Split(key, " ")[0]
		if !slices.Contains(t.HiddenCols, columnName) {
			sortedHeaderKeys = append(sortedHeaderKeys, key)
		}
	}

	for col, key := range sortedHeaderKeys {
		table.SetCell(startRow, col, tview.NewTableCell(key).
			SetTextColor(t.Style.ColumnKeyColor.Color()).
			SetSelectable(false).
			SetBackgroundColor(t.Style.HeaderRowBackgroundColor.Color()).
			SetAlign(tview.AlignCenter))
	}
	startRow++

	for row, doc := range documents {
		for col, key := range sortedHeaderKeys {
			var cellText string
			if val, ok := doc[strings.Split(key, " ")[0]]; ok {
				cellText = util.StringifyMongoValueByType(val)
			}
			if len(cellText) > 30 {
				cellText = cellText[0:30] + "..."
			}

			cell := tview.NewTableCell(cellText).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(30)

			// store _id reference only on col 0 to avoid repetition across the row
			if col == 0 {
				cell.SetReference(doc["_id"])
			}
			table.SetCell(startRow+row, col, cell)
		}
	}
	table.Select(1, 0)
}
