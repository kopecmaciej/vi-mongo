package component

import (
	"context"
	"os"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
)

const (
	InputBarComponent = "InputBar"
)

type InputBar struct {
	*tview.TextArea

	app            *App
	eventChan      chan interface{}
	mutex          sync.Mutex
	label          string
	enabled        bool
	autocompleteOn bool
	docKeys        []string
}

func NewInputBar(label string) *InputBar {
	f := &InputBar{
		TextArea:       tview.NewTextArea(),
		mutex:          sync.Mutex{},
		label:          label,
		eventChan:      make(chan interface{}),
		enabled:        false,
		autocompleteOn: false,
	}

	return f
}

func (i *InputBar) Init(ctx context.Context) error {
	app, err := GetApp(ctx)
	if err != nil {
		return err
	}
	i.app = app
	i.setStyle()
	i.setShortcuts()
	i.SetLabel(" " + i.label + ": ")

	return nil
}

func (i *InputBar) setStyle() {
	i.SetBorder(true)

}

func (i *InputBar) setShortcuts() {
	i.TextArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter, tcell.KeyEsc:
			i.eventChan <- event.Key()
			return nil
		}
		return event
	})
}

const (
	maxHistory = 20
)

// EnableAutocomplete enables autocomplete

func (i *InputBar) LoadNewKeys(keys []string) {
	i.docKeys = keys
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
	i.mutex.Lock()
	defer i.mutex.Unlock()

	i.enabled = true
	i.app.ComponentManager.PushComponent(InputBarComponent)
}

func (i *InputBar) Disable() {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	i.enabled = false
	i.app.ComponentManager.PopComponent()
}

// Toggle enables/disables the input bar but does not force any redraws
func (i *InputBar) Toggle() {
	if i.IsEnabled() {
		i.Disable()
	} else {
		i.Enable()
	}
}

// EventListener listens for events on the input bar
func (i *InputBar) EventListener(accept func(string), reject func()) {
	for {
		key := <-i.eventChan
		if _, ok := key.(tcell.Key); !ok {
			continue
		}
		switch key {
		case tcell.KeyEsc:
			i.app.QueueUpdateDraw(func() {
				i.Disable()
				reject()
			})
		case tcell.KeyEnter:
			i.app.QueueUpdateDraw(func() {
        i.Disable()
				text := i.GetText()
				err := i.SaveToHistory(text)
				if err != nil {
					log.Error().Err(err).Msg("Error saving query to history")
				}
				accept(text)
			})
		}
	}
}

// ToggleAutocomplete toggles autocomplete on and off
func (i *InputBar) ToggleAutocomplete() {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	if i.autocompleteOn {
		i.autocompleteOn = false
	} else {
		i.autocompleteOn = true
	}
}
