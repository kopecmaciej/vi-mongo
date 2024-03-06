package component

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/config"
	"github.com/kopecmaciej/mongui/manager"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
)

const (
	HelpComponent = "Help"
)

// Help is a component that provides a help screen for keybindings
type Help struct {
	*Component
	*tview.Table

	style      *config.Help
	keyManager *manager.KeyManager
}

// NewHelp creates a new Help component
func NewHelp() *Help {
	h := &Help{
		Component: NewComponent(HelpComponent),
		Table:     tview.NewTable(),

		keyManager: manager.NewKeyManager(),
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
	log.Debug().Msgf("Current component: %s", currComponent)

	keys := h.keyManager.GetKeysForComponent(currComponent)
	for i, key := range keys {
		h.Table.SetCell(i, 0, tview.NewTableCell(key.Name).SetTextColor(h.style.KeyColor.Color()))
		h.Table.SetCell(i, 1, tview.NewTableCell(key.Description).SetTextColor(h.style.DescriptionColor.Color()))
	}
}

func (h *Help) setStyle() {
	h.style = &h.app.Styles.Help
	h.SetBorder(true)
	h.SetTitle(" Help ")
	h.Table.SetTitleAlign(tview.AlignLeft)
	h.Table.SetBorderPadding(0, 0, 1, 1)
	h.Table.SetFixed(1, 1)
	h.Table.SetSelectable(false, false)
	h.Table.SetBackgroundColor(h.style.BackgroundColor.Color())
	h.Table.SetBorderColor(h.style.BorderColor.Color())
}

// setKeybindings sets a key binding for the help Component
func (h *Help) setKeybindings() {
	h.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			h.app.Root.Pages.RemovePage("help")
			return nil
		}

		return event
	})

}
