package component

import (
	"context"
	"regexp"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/config"
	"github.com/kopecmaciej/mongui/mongo"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
)

const (
	InputBarComponent = "InputBar"
)

type InputBar struct {
	*tview.InputField

	historyModal   *HistoryModal
	app            *App
	style          *config.InputBar
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
		historyModal:   NewHistoryModal(),
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
	i.setShortcuts(ctx)

	i.SetEventFunc()

	if err := i.historyModal.Init(ctx); err != nil {
		log.Error().Err(err).Msg("Error initializing history modal")
	}

	return nil
}

func (i *InputBar) SetEventFunc() {
	i.SetDoneFunc(func(key tcell.Key) {
		i.eventChan <- key
	})
}

func (i *InputBar) setStyle() {
	i.SetLabel(" " + i.label + ": ")

	i.style = &i.app.Styles.InputBar
	i.SetBorder(true)
	i.SetFieldTextColor(i.style.InputColor.Color())

	// Autocomplete styles
	a := i.style.Autocomplete
	background := a.BackgroundColor.Color()
	main := tcell.StyleDefault.
		Background(a.BackgroundColor.Color()).
		Foreground(a.TextColor.Color())
	selected := tcell.StyleDefault.
		Background(a.ActiveBackgroundColor.Color()).
		Foreground(a.ActiveTextColor.Color())
	i.SetAutocompleteStyles(background, main, selected)
}

func (i *InputBar) setShortcuts(ctx context.Context) {
	i.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlH {
			i.displayHistoryModal()
		}
		return event
	})
}

// EnableAutocomplete enables autocomplete
func (i *InputBar) EnableAutocomplete() {
	ma := mongo.NewMongoAutocomplete()
	mongoKeywords := ma.Operators

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

			// support for mongo keywords
			for _, keyword := range mongoKeywords {
				escaped := regexp.QuoteMeta(currentWord)
				if matched, _ := regexp.MatchString("(?i)^"+escaped, keyword.Display); matched {
					entries = append(entries, keyword.Display)
				}
			}

			// support for document keys
			if i.docKeys != nil {
				for _, keyword := range i.docKeys {
					if matched, _ := regexp.MatchString("(?i)^"+currentWord, keyword); matched {
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

		key := ma.GetOperatorByDisplay(text)
		if key != nil {
			text = key.InsertText
		}

		i.SetWordAtCursor(text)

		return true
	})
}

// LoadNewKeys loads new keys for autocomplete
// It is used when switching databases or collections
func (i *InputBar) LoadNewKeys(keys []string) {
	i.docKeys = keys
}

// Display HistoryModal on the screen
func (i *InputBar) displayHistoryModal() {
	err := i.historyModal.Render()
	if err != nil {
		log.Error().Err(err).Msg("Error rendering history modal")
		ShowErrorModal(i.app.Root, err.Error())
	}
}

// IsEnabled returns true if the input bar is enabled
func (i *InputBar) IsEnabled() bool {
	return i.enabled
}

// Enable enables the input bar, adds component to the stack and forces a redraw
func (i *InputBar) Enable() {
	i.enabled = true
	i.app.ComponentManager.PushComponent(InputBarComponent)
}

// Disable disables the input bar and removes it from the stack
func (i *InputBar) Disable() {
	i.enabled = false
	i.app.ComponentManager.PopComponent()
}

// Toggle toggles the input bar
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
				err := i.historyModal.SaveToHistory(text)
				if err != nil {
					log.Error().Err(err).Msg("Error saving query to history")
				}
				accept(text)
				i.SetText("")
			})
		}
	}
}
