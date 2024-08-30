package page

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
)

const (
	HelpPage = "Help"
)

// Help is a view that provides a help screen for keybindings
type Help struct {
	*core.BaseElement
	*tview.Table

	style *config.HelpStyle

	keyWidth, descWidth int
}

// NewHelp creates a new Help view
func NewHelp() *Help {
	h := &Help{
		BaseElement: core.NewBaseElement(),
		Table:       tview.NewTable(),
	}

	h.SetIdentifier(HelpPage)
	h.SetAfterInitFunc(h.init)

	return h
}

func (h *Help) init() error {
	h.setStyle()
	h.setKeybindings()

	return nil
}

func (h *Help) Render() error {
	h.Clear()
	h.Table.Clear()

	allKeys := h.App.GetKeys().GetAvaliableKeys()
	if h.keyWidth == 0 || h.descWidth == 0 {
		h.keyWidth, h.descWidth = h.calculateMaxWidth(allKeys)
	}

	secondRowElements := []config.OrderedKeys{}
	thirdRowElements := []config.OrderedKeys{}
	row := 0
	col := 0
	for _, viewKeys := range allKeys {
		if viewKeys.Element == "Global" || viewKeys.Element == "Help" {
			thirdRowElements = append(thirdRowElements, viewKeys)
		} else if viewKeys.Element == "Welcome" || viewKeys.Element == "Connector" {
			secondRowElements = append(secondRowElements, viewKeys)
		} else {
			h.renderKeySection([]config.OrderedKeys{viewKeys}, &row, col)
		}
	}

	row = 0
	col = 2
	for _, viewKeys := range secondRowElements {
		h.renderKeySection([]config.OrderedKeys{viewKeys}, &row, col)
	}

	row = 0
	col = 4
	for _, viewKeys := range thirdRowElements {
		h.renderKeySection([]config.OrderedKeys{viewKeys}, &row, col)
	}

	h.Table.ScrollToBeginning()

	return nil
}

// calculateMaxWidth calculates the maximum width of the row
func (h *Help) calculateMaxWidth(keys []config.OrderedKeys) (int, int) {
	keyWidth, descWidth := 0, 0
	for _, viewKeys := range keys {
		for _, key := range viewKeys.Keys {
			if len(key.Keys) > 0 {
				keyWidth = len(key.Keys)
			} else {
				keyWidth = len(key.Runes)
			}

			if len(key.Description) > descWidth {
				descWidth = len(key.Description)
			}
		}
	}
	return keyWidth, descWidth
}

// Add this new method to render key sections
func (h *Help) renderKeySection(keys []config.OrderedKeys, row *int, col int) {
	for _, viewKeys := range keys {
		viewName := viewKeys.Element
		if viewName == "Root" {
			viewName = "Main Layout"
		}
		h.addHeaderSection(viewName, *row, col)
		*row += 2
		h.AddKeySection(viewName, viewKeys.Keys, row, col)
		*row++
	}
}

func (h *Help) addHeaderSection(name string, row, col int) {
	h.Table.SetCell(row+0, col, tview.NewTableCell(name).SetTextColor(h.style.TitleColor.Color()))
	h.Table.SetCell(row+1, col, tview.NewTableCell("-------").SetTextColor(h.style.DescriptionColor.Color()))
	// let's fill blank cells with empty strings
	h.Table.SetCell(row+0, col+1, tview.NewTableCell("").SetTextColor(h.style.TitleColor.Color()))
	h.Table.SetCell(row+1, col+1, tview.NewTableCell("").SetTextColor(h.style.DescriptionColor.Color()))
}

func (h *Help) AddKeySection(name string, keys []config.Key, pos *int, col int) {
	for _, key := range keys {
		var keyString string
		var iter []string
		if len(key.Keys) > 0 {
			iter = key.Keys
		} else {
			iter = key.Runes
		}
		for i, k := range iter {
			if i == 0 {
				keyString = k
			} else {
				keyString = fmt.Sprintf("%s, %s", keyString, k)
			}
		}

		h.Table.SetCell(*pos, col, tview.NewTableCell(keyString).SetTextColor(h.style.KeyColor.Color()))
		h.Table.SetCell(*pos, col+1, tview.NewTableCell(key.Description).SetTextColor(h.style.DescriptionColor.Color()))
		*pos += 1
	}
}

func (h *Help) setStyle() {
	h.style = &h.App.GetStyles().Help
	h.Table.SetBorder(true)
	h.Table.SetTitle(" Help ")
	h.Table.SetBorderPadding(1, 1, 3, 3)
	h.Table.SetSelectable(false, false)
	h.Table.SetBackgroundColor(h.style.BackgroundColor.Color())
	h.Table.SetBorderColor(h.style.BorderColor.Color())
	h.Table.SetTitleColor(h.style.TitleColor.Color())
	h.Table.SetTitleAlign(tview.AlignLeft)
	h.Table.SetEvaluateAllRows(true)
}

func (h *Help) setKeybindings() {
	k := h.App.GetKeys()

	h.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case k.Contains(k.Help.Close, event.Name()):
			h.App.Pages.RemovePage(HelpPage)
			return nil
		}
		return event
	})
}
