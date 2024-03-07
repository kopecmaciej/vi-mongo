package component

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/config"
	"github.com/kopecmaciej/mongui/manager"
	"github.com/rivo/tview"
)

const (
	HelpComponent manager.Component = "Help"
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
	currComponent := h.app.Manager.CurrentComponent()

	cKeys := h.app.Manager.KeyManager.GetKeysForComponent(currComponent)
	gKeys := h.app.Manager.KeyManager.GetKeysForComponent(manager.GlobalComponent)
	hKeys := h.app.Manager.KeyManager.GetKeysForComponent(HelpComponent)
	cKeys = append(cKeys, gKeys...)
	cKeys = append(cKeys, hKeys...)
	for i, key := range cKeys {
		h.Table.SetCell(i, 0, tview.NewTableCell(key.Name).SetTextColor(h.style.KeyColor.Color()))
		h.Table.SetCell(i, 1, tview.NewTableCell(" - ").SetTextColor(h.style.DescriptionColor.Color()))
		h.Table.SetCell(i, 2, tview.NewTableCell(key.Description).SetTextColor(h.style.DescriptionColor.Color()))
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
	manager := h.app.Manager.SetKeyHandlerForComponent(HelpComponent)
	manager(tcell.KeyEsc, 0, "Close Help", func() *tcell.EventKey {
		h.app.Root.RemovePage(HelpComponent)
		return nil
	})

	h.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return h.app.Manager.HandleKeyEvent(event)
	})
}
