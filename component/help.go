package component

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/config"
	"github.com/kopecmaciej/mongui/manager"
	"github.com/kopecmaciej/tview"
)

const (
	HelpComponent manager.Component = "Help"
)

// Help is a component that provides a help screen for keybindings
type Help struct {
	*Component
	*tview.Table

	style *config.HelpStyle
}

// NewHelp creates a new Help component
func NewHelp() *Help {
	h := &Help{
		Component: NewComponent(HelpComponent),
		Table:     tview.NewTable(),
	}

	h.SetAfterInitFunc(h.init)

	return h
}

func (h *Help) init() error {
	h.setStyle()
	h.setKeybindings()

	return nil
}

func (h *Help) Render() error {
	h.Table.Clear()

	currectComponent := h.app.Manager.CurrentComponent()
	cKeys, err := h.app.Keys.GetKeysForComponent(string(currectComponent))
	if err != nil {
		ShowErrorModal(h.app.Root, "Error while getting keys for component", err)
		return err
	}

	pos := 0
	for _, keys := range cKeys {
		if len(keys.Keys) > 0 {
			h.addHeaderSection(keys.Component, pos)
			pos += 3
			h.AddKeySection(keys.Component, keys.Keys, &pos)
		}
	}

	gKeys, err := h.app.Keys.GetKeysForComponent("Global")
	for _, keys := range gKeys {
		h.addHeaderSection(keys.Component, pos)
		pos += 3
		h.AddKeySection(keys.Component, keys.Keys, &pos)
	}

	hKeys, err := h.app.Keys.GetKeysForComponent("Help")
	for _, keys := range hKeys {
		h.addHeaderSection(keys.Component, pos)
		pos += 3
		h.AddKeySection(keys.Component, keys.Keys, &pos)
	}

	return nil
}

func (h *Help) addHeaderSection(name string, row int) {
	h.Table.SetCell(row, 0, tview.NewTableCell(" ").SetTextColor(h.style.DescriptionColor.Color()))
	h.Table.SetCell(row+1, 0, tview.NewTableCell(name).SetTextColor(h.style.TitleColor.Color()))
	h.Table.SetCell(row+2, 0, tview.NewTableCell("-----------").SetTextColor(h.style.DescriptionColor.Color()))
}

func (h *Help) AddKeySection(name string, keys []config.Key, pos *int) {
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
				keyString = fmt.Sprintf("%s", k)
			} else {
				keyString = fmt.Sprintf("%s, %s", keyString, k)
			}
		}

		h.Table.SetCell(*pos, 0, tview.NewTableCell(keyString).SetTextColor(h.style.KeyColor.Color()))
		h.Table.SetCell(*pos, 1, tview.NewTableCell(" - ").SetTextColor(h.style.DescriptionColor.Color()))
		h.Table.SetCell(*pos, 2, tview.NewTableCell(key.Description).SetTextColor(h.style.DescriptionColor.Color()))
		*pos += 1
	}
}

func (h *Help) setStyle() {
	h.style = &h.app.Styles.Help
	h.SetBorder(true)
	h.SetTitle(" Help ")
	h.Table.SetTitleAlign(tview.AlignLeft)
	h.Table.SetBorderPadding(2, 2, 4, 4)
	h.Table.SetFixed(1, 1)
	h.Table.SetSelectable(false, false)
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
