package component

import (
	"context"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/config"
	"github.com/kopecmaciej/mongui/primitives"
)

const (
	HistoryModalComponent = "HistoryModal"
	maxHistory            = 10
)

// HistoryModal is a modal with history of queries
type HistoryModal struct {
	*primitives.ListModal

	app   *App
	style *config.Others
}

func NewHistoryModal() *HistoryModal {
	return &HistoryModal{
		ListModal: primitives.NewListModal(),
	}
}

// Init initializes HistoryModal
func (h *HistoryModal) Init(ctx context.Context) error {
	app, err := GetApp(ctx)
	if err != nil {
		return err
	}
	h.app = app

	h.setStyle()

	return nil
}

func (h *HistoryModal) setStyle() {
	h.style = &h.app.Styles.Others

	h.SetBorder(true)
	h.SetTitle(" History ")
	h.SetBorderPadding(0, 0, 1, 1)
}

func (h *HistoryModal) setShortcuts() {
	h.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'h':
			return tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone)
		case 'l':
			return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
		}
		switch event.Key() {
		case tcell.KeyEnter:
      // TODO: handle enter
			h.app.Root.RemovePage(HistoryModalComponent)
			return nil
		}
		return event
	})
}

// Render loads history from file and renders it
func (h *HistoryModal) Render() error {
	history, err := h.loadHistory()
	if err != nil {
		return err
	}

	for i, entry := range history {
		i += 48
		h.AddItem(entry, "", int32(i), nil)
	}

	h.app.Root.AddPage(HistoryModalComponent, h, true, true)

	return nil
}

// SaveToHistory saves text to history file, if it's not already there.
// It will overwrite oldest entry if history is full.
func (h *HistoryModal) SaveToHistory(text string) error {
	file, err := os.OpenFile("history.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	history, err := h.loadHistory()
	if err != nil {
		return err
	}

	for _, entry := range history {
		if entry == text {
			return nil
		}
	}

	if len(history) >= maxHistory {
		history = history[1:]
	}

	history = append(history, text)

	for _, entry := range history {
		_, err = file.WriteString(entry + "\n")
		if err != nil {
			return err
		}
	}

	return nil
}

// loadHistory loads history from history file
func (h *HistoryModal) loadHistory() ([]string, error) {
	file, err := os.ReadFile("history.txt")
	if err != nil {
		return nil, err
	}

	history := []string{}
	lines := strings.Split(string(file), "\n")

	for _, line := range lines {
		if line != "" {
			history = append(history, line)
		}
	}

	return history, nil
}
