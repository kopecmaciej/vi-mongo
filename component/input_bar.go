package component

import (
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
	*Component
	*tview.InputField

	historyModal   *HistoryModal
	style          *config.InputBar
	mutex          sync.Mutex
	enabled        bool
	autocompleteOn bool
	docKeys        []string
	defaultText    string
}

func NewInputBar(label string) *InputBar {
	i := &InputBar{
		Component: NewComponent(InputBarComponent),
		InputField: tview.NewInputField().
			SetLabel(" " + label + ": "),
		enabled:        false,
		autocompleteOn: false,
	}

	i.SetAfterInitFunc(i.init)

	return i
}

func (i *InputBar) init() error {
	i.setStyle()
	i.setKeybindings()

	i.Subscribe()
	go i.handleEvents()

	return nil
}

func (i *InputBar) setStyle() {
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
	second := tcell.StyleDefault.
		Background(a.BackgroundColor.Color()).
		Foreground(a.SecondaryTextColor.Color()).
		Italic(true)

	i.SetAutocompleteStyles(background, main, selected, second, true)
}

func (i *InputBar) setKeybindings() {
	i.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case '{':
			if i.GetWordAtCursor() == "" {
				i.SetWordAtCursor("{ <$0> }")
				return nil
			}
		case '[':
			if i.GetWordAtCursor() == "" {
				i.SetWordAtCursor("[ <$0> ]")
				return nil
			}
		}
		switch event.Key() {
		case tcell.KeyCtrlH:
			if i.historyModal != nil {
				i.displayHistoryModal()
			}
		case tcell.KeyCtrlD:
			i.SetText("")
			i.SetWordAtCursor("{ <$0> }")
		}
		return event
	})
}

// SetDefaultText sets default text for the input bar
func (i *InputBar) SetDefaultText(text string) {
	i.defaultText = text
}

// DoneFuncHandler sets DoneFunc for the input bar
// It accepts two functions: accept and reject which are called
// when user accepts or rejects the input
func (i *InputBar) DoneFuncHandler(accept func(string), reject func()) {
	i.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEsc:
			i.Toggle()
			reject()
		case tcell.KeyEnter:
			i.Toggle()
			text := i.GetText()
			err := i.historyModal.SaveToHistory(text)
			if err != nil {
				log.Error().Err(err).Msg("Error saving query to history")
			}
			accept(text)
		}
	})
}

// EnableHistory enables history modal
func (i *InputBar) EnableHistory() {
	i.historyModal = NewHistoryModal()

	if err := i.historyModal.Init(i.app); err != nil {
		log.Error().Err(err).Msg("Error initializing history modal")
	}
}

// EnableAutocomplete enables autocomplete
func (i *InputBar) EnableAutocomplete() {
	ma := mongo.NewMongoAutocomplete()
	mongoKeywords := ma.Operators

	i.SetAutocompleteFunc(func(currentText string) (entries []tview.AutocompleteItem) {
		if strings.HasPrefix(currentText, "\"") {
			currentText = currentText[1:]
		}

		words := strings.Fields(currentText)
		if len(words) > 0 {
			currentWord := i.GetWordAtCursor()
			// if word starts with { or [ then we are inside object or array
			// and we should ommmit this character
			if strings.HasPrefix(currentWord, "{") || strings.HasPrefix(currentWord, "[") {
				currentWord = currentWord[1:]
			}
			if currentWord == "" {
				return nil
			}

			// support for mongo keywords
			for _, keyword := range mongoKeywords {
				escaped := regexp.QuoteMeta(currentWord)
				if matched, _ := regexp.MatchString("(?i)^"+escaped, keyword.Display); matched {
					entry := tview.AutocompleteItem{Main: keyword.Display, Secondary: keyword.Description}
					entries = append(entries, entry)
				}
			}

			// support for document keys
			if i.docKeys != nil {
				for _, keyword := range i.docKeys {
					if matched, _ := regexp.MatchString("(?i)^"+currentWord, keyword); matched {
						entries = append(entries, tview.AutocompleteItem{Main: keyword})
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

// Draws default text if input is empty
func (i *InputBar) DrawDefaultText() {
	if i.GetText() == "" {
		go i.app.QueueUpdateDraw(func() {
			i.SetWordAtCursor(i.defaultText)
		})
	}
}

func (i *InputBar) handleEvents() {
	for event := range i.listener {
		sender, eventKey := event.Sender, event.EventKey
		switch sender {
		case i.historyModal.GetIdentifier():
			switch eventKey.Key() {
			case tcell.KeyEnter:
				i.app.QueueUpdateDraw(func() {
					i.SetText(i.historyModal.GetText())
					i.app.SetFocus(i)
				})
			case tcell.KeyEsc:
				i.app.QueueUpdateDraw(func() {
					i.app.SetFocus(i)
				})
			}
		}
	}
}
