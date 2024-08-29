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
func (t *Table) MoveUpUntil(condition func(row int, cell *tview.TableCell) bool) {
	row, col := t.GetSelection()
	for row > 0 {
		row--
		cell := t.GetCell(row, col)
		if condition(row, cell) {
			t.Select(row, col)
			return
		}
	}
}

// MoveDownUntil moves the selection down until a condition is met
func (t *Table) MoveDownUntil(condition func(row int, cell *tview.TableCell) bool) {
	row, col := t.GetSelection()
	for row < t.GetRowCount()-1 {
		row++
		cell := t.GetCell(row, col)
		if condition(row, cell) {
			t.Select(row, col)
			return
		}
	}
}

// ImproveScrolling allows to scroll to the beginning and end of the table
// while moving up and down when reaching the end of selectable area
// but there are rows non selectable that we like to see
// TODO: Implement
func (t *Table) ImproveScrolling() {
}
