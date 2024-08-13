package component

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/config"
	"github.com/kopecmaciej/mongui/manager"
	"github.com/kopecmaciej/tview"
	"github.com/rs/zerolog/log"
)

const (
	HelpComponent manager.Component = "Help"
)

// Help is a component that provides a help screen for keybindings
type Help struct {
	*Component
	*tview.Flex

	Table *tview.Table
	style *config.HelpStyle
}

// NewHelp creates a new Help component
func NewHelp() *Help {
	h := &Help{
		Component: NewComponent(HelpComponent),
		Flex:      tview.NewFlex(),
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

func (h *Help) Render(fullScreen bool) error {
	h.Clear()
	h.Table.Clear()
	_, _, width, height := h.GetRect()
	log.Info().Int("width", width).Int("height", height).Msg("Dimensions")

	currectComponent := h.app.Manager.CurrentComponent()
	cKeys, err := h.app.Keys.GetKeysForComponent(string(currectComponent))
	if err != nil {
		ShowErrorModal(h.app.Root, "Error while getting keys for component", err)
		return err
	}

	h.fillWithEmptySpace(width, height)

	pos, col := 0, 0
	for _, keys := range cKeys {
		if len(keys.Keys) > 0 {
			if h.shouldMoveToNextColumn(height, pos+3+len(keys.Keys)) {
				pos = 0
				col += 3
			}
			h.addHeaderSection(keys.Component, pos, col)
			pos += 2
			h.AddKeySection(keys.Component, keys.Keys, &pos, col)
		}
	}

	gKeys, err := h.app.Keys.GetKeysForComponent("Global")
	if err != nil {
		ShowErrorModal(h.app.Root, "Error while getting keys for component", err)
		return err
	}
	for _, keys := range gKeys {
		if h.shouldMoveToNextColumn(height, pos+3+len(keys.Keys)) {
			pos = 0
			col += 3
		}
		h.addHeaderSection(keys.Component, pos, col)
		pos += 2
		h.AddKeySection(keys.Component, keys.Keys, &pos, col)
	}

	hKeys, err := h.app.Keys.GetKeysForComponent("Help")
	if err != nil {
		ShowErrorModal(h.app.Root, "Error while getting keys for component", err)
		return err
	}
	for _, keys := range hKeys {
		if h.shouldMoveToNextColumn(height, pos+3+len(keys.Keys)) {
			pos = 0
			col += 3
		}
		h.addHeaderSection(keys.Component, pos, col)
		pos += 2
		h.AddKeySection(keys.Component, keys.Keys, &pos, col)
	}

	h.Table.ScrollToBeginning()

	if fullScreen {
		h.Flex.AddItem(tview.NewBox(), 0, 1, false)
	}
	h.Flex.AddItem(h.Table, 0, 1, true)
	if fullScreen {
		h.Flex.AddItem(tview.NewBox(), 0, 1, false)
	}

	return nil
}

func (h *Help) shouldMoveToNextColumn(maxNumberOfRows int, currRows int) bool {
	return currRows >= maxNumberOfRows
}

func (h *Help) addHeaderSection(name string, row, col int) {
	h.Table.SetCell(row+0, col, tview.NewTableCell(name).SetTextColor(h.style.TitleColor.Color()))
	h.Table.SetCell(row+1, col, tview.NewTableCell("-------").SetTextColor(h.style.DescriptionColor.Color()))
}

func (h *Help) fillWithEmptySpace(width, height int) {
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			h.Table.SetCell(i, j, tview.NewTableCell(" ").SetTextColor(h.style.BackgroundColor.Color()))
		}
	}
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
