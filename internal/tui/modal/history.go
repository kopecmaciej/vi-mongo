package modal

import (
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/tui/primitives"
)

const (
	HistoryModal = "HistoryModal"
	QueryBar     = "QueryBar"

	maxHistory = 10
)

// History is a modal with history of queries
type History struct {
	*core.BaseElement
	*primitives.ListModal

	style *config.HistoryStyle
}

func NewHistoryModal() *History {
	h := &History{
		BaseElement: core.NewBaseElement(),
		ListModal:   primitives.NewListModal(),
	}

	h.SetIdentifier(HistoryModal)
	h.SetAfterInitFunc(h.init)

	return h
}

// Init initializes HistoryModal
func (h *History) init() error {
	h.setStyle()
	h.setKeybindings()

	return nil
}

func (h *History) setStyle() {
	h.style = &h.App.GetStyles().History

	h.SetTitle(" History ")
	h.SetBorder(true)
	h.ShowSecondaryText(false)
	h.SetBackgroundColor(h.style.BackgroundColor.Color())
	mainStyle := tcell.StyleDefault.
		Foreground(h.style.TextColor.Color()).
		Background(h.style.BackgroundColor.Color())
	h.SetMainTextStyle(mainStyle)

	selectedStyle := tcell.StyleDefault.
		Foreground(h.style.SelectedTextColor.Color()).
		Background(h.style.SelectedBackgroundColor.Color())
	h.SetSelectedStyle(selectedStyle)
}

func (h *History) setKeybindings() {
	keys := h.App.GetKeys()
	h.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case keys.Contains(keys.History.AcceptEntry, event.Name()):
			return h.sendEventAndClose(event)
		case keys.Contains(keys.History.CloseHistory, event.Name()):
			return h.sendEventAndClose(event)
		case keys.Contains(keys.History.ClearHistory, event.Name()):
			return h.clearHistory()
		}
		return event
	})
}

func (h *History) sendEventAndClose(event *tcell.EventKey) *tcell.EventKey {
	eventKey := manager.EventMsg{EventKey: event, Sender: h.GetIdentifier()}
	h.SendToElement(QueryBar, eventKey)
	h.App.Pages.RemovePage(h.GetIdentifier())

	return nil
}

func (h *History) clearHistory() *tcell.EventKey {
	err := os.WriteFile(getHisotryFilePath(), []byte{}, 0644)
	if err != nil {
		ShowError(h.App.Pages, "Failed to clear history", err)
	}
	h.App.Pages.RemovePage(h.GetIdentifier())
	ShowInfo(h.App.Pages, "History cleared")

	return nil
}

// Render loads history from file and renders it
func (h *History) Render() error {
	history, err := h.loadHistory()
	if err != nil {
		return err
	}

	h.Clear()

	// load in reverse order
	for i := len(history) - 1; i >= 0; i-- {
		rune := 57 - i
		entry := history[i]
		h.AddItem(entry, "", int32(rune), nil)
	}

	h.App.Pages.AddPage(h.GetIdentifier(), h, true, true)

	return nil
}

// SaveToHistory saves text to history file, if it's not already there.
// It will overwrite oldest entry if history is full.
func (h *History) SaveToHistory(text string) error {
	history, err := h.loadHistory()
	if err != nil {
		return err
	}

	var updatedHistory []string
	for _, line := range history {
		if line != text {
			updatedHistory = append(updatedHistory, line)
			if len(updatedHistory) >= maxHistory {
				updatedHistory = updatedHistory[1:]
			}
		}
	}
	updatedHistory = append(updatedHistory, text)

	historyFile, err := os.OpenFile(getHisotryFilePath(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer historyFile.Close()

	for _, entry := range updatedHistory {
		_, err = historyFile.WriteString(entry + "\n")
		if err != nil {
			return err
		}
	}

	return nil
}

// GetText returns text from selected item
func (h *History) GetText() string {
	text := h.ListModal.GetText()

	return strings.TrimSpace(text)
}

// loadHistory loads history from history file
func (h *History) loadHistory() ([]string, error) {
	bytes, err := os.ReadFile(getHisotryFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			err = os.WriteFile(getHisotryFilePath(), []byte{}, 0644)
		}
		return nil, err
	}

	history := []string{}
	lines := strings.Split(string(bytes), "\n")

	for _, line := range lines {
		if line != "" {
			history = append(history, line)
		}
	}

	return history, nil
}

func getHisotryFilePath() string {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return ""
	}

	return configDir + "/history.txt"
}
