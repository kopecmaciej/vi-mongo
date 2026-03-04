package view

import (
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DocSeparator is the prefix used for document separator rows.
// Used externally to identify separator cells (e.g. for _id lookup).
const DocSeparator = "────────────────────────────────────────"

// TableJson renders MongoDB documents as indented, multi-line JSON rows in a table,
// separated by a horizontal line between documents.
type TableJson struct {
	SeparatorColor tcell.Color
}

func NewTableJson() *TableJson {
	return &TableJson{
		SeparatorColor: tcell.ColorGray,
	}
}

func (t *TableJson) Render(table *core.Table, startRow int, documents []primitive.M) error {
	table.SetFixed(0, 0)
	row := startRow
	for _, doc := range documents {
		_id := doc["_id"]
		jsoned, err := mongo.ParseBsonDocument(doc)
		if err != nil {
			return err
		}
		t.renderDocument(table, jsoned, &row, _id)
	}
	table.ScrollToBeginning()
	if table.GetRowCount() > 1 {
		table.Select(1, 0)
	}
	return nil
}

func (t *TableJson) renderDocument(table *core.Table, doc string, row *int, _id any) {
	// Separator row acts as the document boundary and stores the _id reference.
	table.SetCell(*row, 0, tview.NewTableCell(DocSeparator).
		SetAlign(tview.AlignLeft).
		SetTextColor(t.SeparatorColor).
		SetSelectable(false).
		SetReference(_id))
	*row++

	indentedJson, err := mongo.IndentJson(doc)
	if err != nil {
		return
	}
	keyRegexWithIndent := regexp.MustCompile(`(?m)^\s{2}"([^"]+)":`)
	// strip the outer { } lines — only render the fields
	lines := strings.Split(indentedJson.String(), "\n")
	if len(lines) < 2 {
		return
	}
	lines = lines[1 : len(lines)-1]

	currLine := ""
	for _, line := range lines {
		if keyRegexWithIndent.MatchString(line) {
			if currLine != "" {
				table.SetCell(*row, 0, tview.NewTableCell(currLine).SetAlign(tview.AlignLeft))
				*row++
			}
			currLine = line
		} else {
			currLine += util.TrimMultipleSpaces(line)
		}
	}
	if currLine != "" {
		table.SetCell(*row, 0, tview.NewTableCell(currLine).SetAlign(tview.AlignLeft))
		*row++
	}
}
