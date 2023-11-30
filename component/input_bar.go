package component

import (
	"context"
	"os"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Event string

type InputBar struct {
	*tview.InputField

	EventChan      chan interface{}
	mutex          sync.Mutex
	label          string
	enabled        bool
	AutocompleteOn bool
}

func NewInputBar(label string) *InputBar {
	f := &InputBar{
		InputField:     tview.NewInputField(),
		mutex:          sync.Mutex{},
		label:          label,
		EventChan:      make(chan interface{}),
		enabled:        false,
		AutocompleteOn: false,
	}

	return f
}

func (i *InputBar) Init(ctx context.Context) error {
	i.setStyle()
	i.SetLabel(" " + i.label + ": ")

	i.SetEventFunc()

	if i.AutocompleteOn {
		i.Autocomplete()
	}

	return nil
}

func (i *InputBar) SetEventFunc() {
	i.SetDoneFunc(func(key tcell.Key) {
		i.EventChan <- key
	})
}

func (i *InputBar) setStyle() {
	i.SetBorder(true)
	i.SetFieldTextColor(tcell.ColorYellow)

	autocompleteBg := tcell.ColorGreen.TrueColor()
	autocompleteMainStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(autocompleteBg)
	autocompleteSecondaryStyle := tcell.StyleDefault.Foreground(tcell.ColorBlue).Background(autocompleteBg)
	i.SetAutocompleteStyles(autocompleteBg, autocompleteMainStyle, autocompleteSecondaryStyle)
}

func (i *InputBar) setShortcuts(ctx context.Context) {
	i.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlSpace:
			i.ToggleAutocomplete()
		}
		return event
	})
}

const (
	maxHistory = 20
)

func (i *InputBar) Autocomplete() {
	history, err := i.LoadHistory()
	if err != nil {
		return
	}

	i.SetAutocompleteFunc(func(currentText string) (entries []string) {
		for _, entry := range history {
			if strings.Contains(entry, currentText) {
				entries = append(entries, entry)
			}
		}
		return entries
	})
}

func (i *InputBar) SaveToHistory(text string) error {
	file, err := os.OpenFile("history.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	history, err := i.LoadHistory()
	if err != nil {
		return err
	}

	for _, entry := range history {
		if entry == text {
			return nil
		}
	}

	if _, err := file.WriteString(text + "\n"); err != nil {
		return err
	}

	return nil
}

func (i *InputBar) LoadHistory() ([]string, error) {
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

func (i *InputBar) IsEnabled() bool {
	return i.enabled
}

func (i *InputBar) Enable() {
	i.enabled = true
}

func (i *InputBar) Disable() {
	i.enabled = false
}

func (i *InputBar) Toggle() {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	if i.IsEnabled() {
		i.Disable()
	} else {
		i.Enable()
	}
}

func (i *InputBar) ToggleAutocomplete() {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	if i.AutocompleteOn {
		i.AutocompleteOn = false
	} else {
		i.AutocompleteOn = true
	}
}
