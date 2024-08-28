package component

import (
	"regexp"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/tui/modal"
	"github.com/rs/zerolog/log"
)

type InputBar struct {
	*core.BaseElement
	*tview.InputField

	historyModal   *modal.History
	style          *config.InputBarStyle
	enabled        bool
	autocompleteOn bool
	docKeys        []string
	defaultText    string
}

func NewInputBar(viewName string, label string) *InputBar {
	i := &InputBar{
		BaseElement: core.NewBaseElement(viewName),
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

	cpFunc := func(text string) {
		err := clipboard.WriteAll(text)
		if err != nil {
			log.Error().Err(err).Msg("Error writing to clipboard")
		}
	}
	pasteFunc := func() string {
		text, err := clipboard.ReadAll()
		if err != nil {
			log.Error().Err(err).Msg("Error reading from clipboard")
			return ""
		}
		return strings.TrimSpace(text)
	}
	i.SetClipboard(cpFunc, pasteFunc)

	return nil
}

func (i *InputBar) setStyle() {
	i.style = &i.App.GetStyles().InputBar
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

		k := i.App.GetKeys()
		switch {
		case k.Contains(k.Root.Content.QueryBar.ShowHistory, event.Name()):
			if i.historyModal != nil {
				i.displayHistoryModal()
			}
		case k.Contains(k.Root.Content.QueryBar.ClearInput, event.Name()):
			i.SetText("")
			i.SetWordAtCursor(i.defaultText)
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
			i.Toggle("")
			reject()
		case tcell.KeyEnter:
			i.Toggle("")
			text := i.GetText()
			if i.historyModal != nil {
				err := i.historyModal.SaveToHistory(text)
				if err != nil {
					log.Error().Err(err).Msg("Error saving query to history")
				}
			}
			accept(text)
		}
	})
}

// EnableHistory enables history modal
func (i *InputBar) EnableHistory() {
	i.historyModal = modal.NewHistoryModal()

	if err := i.historyModal.Init(i.App); err != nil {
		log.Error().Err(err).Msg("Error initializing history modal")
	}
}

// EnableAutocomplete enables autocomplete
func (i *InputBar) EnableAutocomplete() {
	ma := mongo.NewMongoAutocomplete()
	mongoKeywords := ma.Operators

	i.SetAutocompleteFunc(func(currentText string) (entries []tview.AutocompleteItem) {
		currentText = strings.TrimPrefix(currentText, "\"")

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
		modal.ShowError(i.App.Pages, "Error rendering history modal", err)
	}
}

// Draws default text if input is empty
func (i *InputBar) Toggle(text string) {
	i.BaseElement.Toggle()
	if text == "" {
		text = i.GetText()
	}
	if text == "" {
		go i.App.QueueUpdateDraw(func() {
			i.SetWordAtCursor(i.defaultText)
		})
	}
}

func (i *InputBar) handleEvents() {
	for event := range i.Listener {
		sender, eventKey := event.Sender, event.EventKey
		switch sender {
		case i.historyModal.GetIdentifier():
			switch eventKey.Key() {
			case tcell.KeyEnter:
				i.App.QueueUpdateDraw(func() {
					i.SetText(i.historyModal.GetText())
					i.App.SetFocus(i)
				})
			case tcell.KeyEsc, tcell.KeyCtrlY:
				i.App.QueueUpdateDraw(func() {
					i.App.SetFocus(i)
				})
			}
		}
	}
}
