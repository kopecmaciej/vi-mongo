package component

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/config"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
)

const (
	HelpComponent tview.Identifier = "Help"
)

// Help is a component that provides a help screen for keybindings
type Help struct {
	*Component
	*tview.Table

	style *config.Help
}

// NewHelp creates a new Help component
func NewHelp() *Help {
	h := &Help{
		Component: NewComponent(HelpComponent),
		Table:     tview.NewTable(),
	}

	h.SetIdentifier(HelpComponent)

	h.SetAfterInitFunc(h.init)

	return h
}

func (h *Help) init() error {
	h.setStyle()
	h.setKeybindings()

	return nil
}

func (h *Help) Render() {
	h.Table.Clear()

	currectComponent := h.app.Manager.CurrentComponent()
	log.Info().Msgf("Current Component: %v", currectComponent)
	cKeys, _ := h.app.Keys.GetKeysForComponent(string(currectComponent))

	pos := 0
	h.addSectionHeader(string(currectComponent), pos)
	pos += 3
	for _, key := range cKeys {
		k := fmt.Sprintf("%v", key.Keys)
		h.Table.SetCell(pos, 0, tview.NewTableCell(k).SetTextColor(h.style.KeyColor.Color()))
		h.Table.SetCell(pos, 1, tview.NewTableCell(" - ").SetTextColor(h.style.DescriptionColor.Color()))
		h.Table.SetCell(pos, 2, tview.NewTableCell(key.Description).SetTextColor(h.style.DescriptionColor.Color()))
		pos += 1
	}

	// h.addSectionHeader("Global Keys", pos)
	// pos += 3
	// for _, key := range gKeys {
	// 	h.Table.SetCell(pos, 0, tview.NewTableCell(key.Name).SetTextColor(h.style.KeyColor.Color()))
	// 	h.Table.SetCell(pos, 1, tview.NewTableCell(" - ").SetTextColor(h.style.DescriptionColor.Color()))
	// 	h.Table.SetCell(pos, 2, tview.NewTableCell(key.Description).SetTextColor(h.style.DescriptionColor.Color()))
	// 	pos += 1
	// }
	// h.addSectionHeader("Help Keys", pos)
	// pos += 3
	// for _, key := range hKeys {
	// 	h.Table.SetCell(pos, 0, tview.NewTableCell(key.Name).SetTextColor(h.style.KeyColor.Color()))
	// 	h.Table.SetCell(pos, 1, tview.NewTableCell(" - ").SetTextColor(h.style.DescriptionColor.Color()))
	// 	h.Table.SetCell(pos, 2, tview.NewTableCell(key.Description).SetTextColor(h.style.DescriptionColor.Color()))
	// 	pos += 1
	// }
}

func (h *Help) addSectionHeader(name string, row int) {
	h.Table.SetCell(row, 0, tview.NewTableCell(" ").SetTextColor(h.style.DescriptionColor.Color()))
	h.Table.SetCell(row+1, 0, tview.NewTableCell(name).SetTextColor(h.style.TitleColor.Color()))
	h.Table.SetCell(row+2, 0, tview.NewTableCell("-----------").SetTextColor(h.style.DescriptionColor.Color()))
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
		case k.Contains(k.HelpKeys.Close, event.Name()):
			h.app.Root.RemovePage(HelpComponent)
			return nil
		}
		return event
	})
}
