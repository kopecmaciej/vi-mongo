package component

import (
	"context"
	"os"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/mongo"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
)

const (
	InputBarComponent = "InputBar"
)

type InputBar struct {
	*tview.InputField

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
		InputField:     tview.NewInputField(),
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
	i.SetLabel(" " + i.label + ": ")

	i.SetEventFunc()

	return nil
}

func (i *InputBar) SetEventFunc() {
	i.SetDoneFunc(func(key tcell.Key) {
		i.eventChan <- key
	})
}

func (i *InputBar) setStyle() {
	i.SetBorder(true)
	i.SetFieldTextColor(tcell.ColorYellow)
	i.SetTextSurroudings("{", "}", 2)

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

func (i *InputBar) AutocompleteHistory() {
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

// EnableAutocomplete enables autocomplete
func (i *InputBar) EnableAutocomplete() {
	mongoAutocomplete := mongo.NewMongoAutocomplete()
	mongoKeywords := mongoAutocomplete.Operators

	i.SetAutocompleteFunc(func(currentText string) (entries []string) {
		if strings.HasPrefix(currentText, "\"") {
			currentText = currentText[1:]
		}

		words := strings.Fields(currentText)
		if len(words) > 0 {
			currentWord := i.GetWordUnderCursor()
			if currentWord == "" {
				return nil
			}
			// if word starts with { or [ then we are inside object or array
			// and we should ommmit this character
			if strings.HasPrefix(currentWord, "{") || strings.HasPrefix(currentWord, "[") {
				currentWord = currentWord[1:]
			}
			if strings.HasPrefix(currentWord, "$") {
				for _, keyword := range mongoKeywords {
					if strings.HasPrefix(keyword, currentWord) {
						entries = append(entries, keyword)
					}
				}
			}
			// support for objectID
			if strings.HasPrefix(currentWord, "O") {
				objectId := mongoAutocomplete.ObjectID.InsertText
				if strings.HasPrefix(objectId, currentWord) {
					entries = append(entries, mongoAutocomplete.ObjectID.Display)
				}
			}

			if i.docKeys != nil {
				for _, keyword := range i.docKeys {
					if strings.HasPrefix(keyword, currentWord) {
						entries = append(entries, keyword)
					}
				}
			}
		}

		return entries
	})

	i.SetAutocompletedFunc(func(text string, index, source int) bool {
		if source == 0 {
			return false
		}

		i.SetWordAtCursor(text)

		return true
	})
}

func (i *InputBar) LoadNewKeys(keys []string) {
	i.docKeys = keys
}

// DisableAutocomplete disables autocomplete
func (i *InputBar) DisableAutocomplete() {
	i.SetAutocompleteFunc(nil)
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
	i.app.ComponentManager.PushComponent(InputBarComponent)
}

func (i *InputBar) Disable() {
	i.enabled = false
	i.app.ComponentManager.PopComponent()
}

// Toggle enables/disables the input bar but does not force any redraws
func (i *InputBar) Toggle() {
	i.mutex.Lock()
	defer i.mutex.Unlock()

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
				i.Toggle()
				reject()
			})
		case tcell.KeyEnter:
			i.app.QueueUpdateDraw(func() {
				i.Toggle()
				text := i.GetText()
				err := i.SaveToHistory(text)
				if err != nil {
					log.Error().Err(err).Msg("Error saving query to history")
				}
				accept(text)
				i.SetText("")
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
