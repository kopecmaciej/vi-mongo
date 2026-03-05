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

// TableJson renders MongoDB documents as indented, multi-line JSON rows in a table.
// Documents are wrapped with { } braces; between documents } { is used as a separator.
// Brace rows are non-selectable and store the document's _id as a cell reference,
// which is the mechanism used to identify document boundaries during lookup.
type TableJson struct {
	BraceColor tcell.Color
}

func NewTableJson() *TableJson {
	return &TableJson{
		BraceColor: tcell.ColorGreen,
	}
}

func (t *TableJson) Render(table *core.Table, startRow int, documents []primitive.M) error {
	table.SetFixed(0, 0)
	row := startRow
	for i, doc := range documents {
		_id := doc["_id"]
		jsoned, err := mongo.ParseBsonDocument(doc)
		if err != nil {
			return err
		}
		t.renderDocument(table, jsoned, &row, _id, i == 0)
	}
	if len(documents) > 0 {
		table.SetCell(row, 0, tview.NewTableCell("}").
			SetAlign(tview.AlignLeft).
			SetTextColor(t.BraceColor).
			SetSelectable(false))
	}
	table.ScrollToBeginning()
	if table.GetRowCount() > 1 {
		table.Select(1, 0)
	}
	return nil
}

func (t *TableJson) renderDocument(table *core.Table, doc string, row *int, _id any, isFirst bool) {
	// Opening brace row stores the _id reference — used externally to identify document boundaries.
	openBrace := "{"
	if !isFirst {
		openBrace = "} {"
	}
	table.SetCell(*row, 0, tview.NewTableCell(openBrace).
		SetAlign(tview.AlignLeft).
		SetTextColor(t.BraceColor).
		SetSelectable(false).
		SetReference(_id))
	*row++

	indentedJson, err := mongo.IndentJson(doc)
	if err != nil {
		return
	}
	keyRegexWithIndent := regexp.MustCompile(`(?m)^\s{2}"([^"]+)":`)
	lines := strings.Split(indentedJson.String(), "\n")
	if len(lines) < 2 {
		return
	}
	// strip outer { } — only render the fields
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
