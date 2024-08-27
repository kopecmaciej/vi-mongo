package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/tui/modal"
)

const (
	HelpView = "Help"
)

// Help is a view that provides a help screen for keybindings
type Help struct {
	*core.BaseView
	*tview.Flex

	Table *tview.Table
	style *config.HelpStyle
}

// NewHelp creates a new Help view
func NewHelp() *Help {
	h := &Help{
		BaseView: core.NewBaseView(HelpView),
		Flex:     tview.NewFlex(),
		Table:    tview.NewTable(),
	}

	h.SetAfterInitFunc(h.init)

	return h
}

func (h *Help) init() error {
	h.setStyle()
	h.setKeybindings()

	return nil
}

func (h *Help) Render(fullScreen bool) error {
	h.Clear()
	h.Table.Clear()

	currectView := h.App.Manager.CurrentView()
	cKeys, err := h.App.GetKeys().GetKeysForView(string(currectView))
	if err != nil {
		modal.ShowError(h.App.Pages, "No keys found for current view", err)
		return err
	}

	row := 0
	h.renderKeySection(cKeys, &row)

	gKeys, err := h.App.GetKeys().GetKeysForView("Global")
	if err != nil {
		modal.ShowError(h.App.Pages, "Error while getting keys for view", err)
		return err
	}
	h.renderKeySection(gKeys, &row)

	hKeys, err := h.App.GetKeys().GetKeysForView("Help")
	if err != nil {
		modal.ShowError(h.App.Pages, "Error while getting keys for view", err)
		return err
	}
	h.renderKeySection(hKeys, &row)

	h.Table.ScrollToBeginning()

	if fullScreen {
		h.Flex.AddItem(tview.NewBox(), 0, 1, false)
		h.Flex.AddItem(h.Table, 0, 3, true)
		h.Flex.AddItem(tview.NewBox(), 0, 1, false)
	} else {
		h.Flex.AddItem(h.Table, 0, 1, true)
	}

	return nil
}

// Add this new method to render key sections
func (h *Help) renderKeySection(keys []config.OrderedKeys, row *int) {
	for _, viewKeys := range keys {
		if viewKeys.View == "Root" {
			viewKeys.View = "Main Layout"
		}
		h.addHeaderSection(viewKeys.View, *row, 0)
		*row += 2
		h.AddKeySection(viewKeys.View, viewKeys.Keys, row, 0)
		*row++
	}
}

func (h *Help) addHeaderSection(name string, row, col int) {
	h.Table.SetCell(row+0, col, tview.NewTableCell(name).SetTextColor(h.style.TitleColor.Color()))
	h.Table.SetCell(row+1, col, tview.NewTableCell("-------").SetTextColor(h.style.DescriptionColor.Color()))
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
		h.Table.SetCell(*pos, col+1, tview.NewTableCell(" - ").SetTextColor(h.style.DescriptionColor.Color()))
		h.Table.SetCell(*pos, col+2, tview.NewTableCell(key.Description).SetTextColor(h.style.DescriptionColor.Color()))
		*pos += 1
	}
}

func (h *Help) setStyle() {
	h.style = &h.App.GetStyles().Help
	h.Table.SetBorder(true)
	h.Table.SetTitle(" Help ")
	h.Table.SetTitleAlign(tview.AlignLeft)
	h.Table.SetBorderPadding(0, 0, 1, 1)
	h.Table.SetSelectable(true, false) // Allow row selection for scrolling
	h.Table.SetBackgroundColor(h.style.BackgroundColor.Color())
	h.Table.SetBorderColor(h.style.BorderColor.Color())
}

func (h *Help) setKeybindings() {
	k := h.App.GetKeys()

	h.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case k.Contains(k.Help.Close, event.Name()):
			h.App.Pages.RemovePage(HelpView)
			return nil
		}
		return event
	})
}
