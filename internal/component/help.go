package component

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/internal/config"
	"github.com/kopecmaciej/tview"
)

const (
	HelpComponent = "Help"
)

// Help is a component that provides a help screen for keybindings
type Help struct {
	*BaseComponent
	*tview.Flex

	Table *tview.Table
	style *config.HelpStyle
}

// NewHelp creates a new Help component
func NewHelp() *Help {
	h := &Help{
		BaseComponent: NewBaseComponent(HelpComponent),
		Flex:          tview.NewFlex(),
		Table:         tview.NewTable(),
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

	currectComponent := h.app.Manager.CurrentComponent()
	cKeys, err := h.app.Keys.GetKeysForComponent(string(currectComponent))
	if err != nil {
		ShowErrorModal(h.app.Root, "No keys found for current component", err)
		return err
	}

	row := 0
	h.renderKeySection(cKeys, &row)

	gKeys, err := h.app.Keys.GetKeysForComponent("Global")
	if err != nil {
		ShowErrorModal(h.app.Root, "Error while getting keys for component", err)
		return err
	}
	h.renderKeySection(gKeys, &row)

	hKeys, err := h.app.Keys.GetKeysForComponent("Help")
	if err != nil {
		ShowErrorModal(h.app.Root, "Error while getting keys for component", err)
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
	for _, componentKeys := range keys {
		if componentKeys.Component == "Root" {
			componentKeys.Component = "Main Layout"
		}
		h.addHeaderSection(componentKeys.Component, *row, 0)
		*row += 2
		h.AddKeySection(componentKeys.Component, componentKeys.Keys, row, 0)
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
	h.style = &h.app.Styles.Help
	h.Table.SetBorder(true)
	h.Table.SetTitle(" Help ")
	h.Table.SetTitleAlign(tview.AlignLeft)
	h.Table.SetBorderPadding(0, 0, 1, 1)
	h.Table.SetSelectable(true, false) // Allow row selection for scrolling
	h.Table.SetBackgroundColor(h.style.BackgroundColor.Color())
	h.Table.SetBorderColor(h.style.BorderColor.Color())
}

// setKeybindings sets a key binding for the help Component
func (h *Help) setKeybindings() {
	k := h.app.Keys

	h.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case k.Contains(k.Help.Close, event.Name()):
			h.app.Root.RemovePage(HelpComponent)
			return nil
		}
		return event
	})
}
