package core

import (
	"github.com/kopecmaciej/tview"
)

type Table struct {
	*tview.Table
}

func NewTable() *Table {
	return &Table{
		Table: tview.NewTable(),
	}
}

// MoveUpUntil moves the selection up until a condition is met
func (t *Table) MoveUpUntil(row, col int, condition func(cell *tview.TableCell) bool) {
	for row > 0 {
		row--
		cell := t.GetCell(row, col)
		if condition(cell) {
			t.Select(row, col)
			return
		}
	}
}

// MoveDownUntil moves the selection down until a condition is met
func (t *Table) MoveDownUntil(row, col int, condition func(cell *tview.TableCell) bool) {
	for row < t.GetRowCount()-1 {
		row++
		cell := t.GetCell(row, col)
		if condition(cell) {
			t.Select(row, col)
			return
		}
	}
}

// GetCellAboveThatMatch returns the cell above the current selection that matches the given condition
func (t *Table) GetCellAboveThatMatch(row, col int, condition func(cell *tview.TableCell) bool) *tview.TableCell {
	for row > 0 {
		row--
		cell := t.GetCell(row, col)
		if condition(cell) {
			return cell
		}
	}
	return nil
}

// GetCellBelowThatMatch returns the cell below the current selection that matches the given condition
func (t *Table) GetCellBelowThatMatch(row, col int, condition func(cell *tview.TableCell) bool) *tview.TableCell {
	for row < t.GetRowCount()-1 {
		row++
		cell := t.GetCell(row, col)
		if condition(cell) {
			return cell
		}
	}
	return nil
}

// GetContentFromRows returns the content of the table from the selected rows
func (t *Table) GetContentFromRows(rows []int) []string {
	content := []string{}
	for _, row := range rows {
		content = append(content, t.GetCell(row, 0).GetReference().(string))
	}
	return content
}

// ImproveScrolling allows to scroll to the beginning and end of the table
// while moving up and down when reaching the end of selectable area
// but there are rows non selectable that we like to see
// TODO: Implement
func (t *Table) ImproveScrolling() {
}
