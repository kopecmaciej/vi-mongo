package component

import (
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
	"github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/tui/modal"
	"github.com/kopecmaciej/vi-mongo/internal/util"
	"github.com/rs/zerolog/log"
)

type InputBar struct {
	*core.BaseElement
	*core.InputField

	historyModal   *modal.History
	style          *config.InputBarStyle
	enabled        bool
	autocompleteOn bool
	docKeys        []string
	defaultText    string
	pasteFunc      func() string
}

func NewInputBar(barId tview.Identifier, label string) *InputBar {
	i := &InputBar{
		BaseElement:    core.NewBaseElement(),
		InputField:     core.NewInputField(),
		enabled:        false,
		autocompleteOn: false,
	}

	i.InputField.SetLabel(" " + label + ": ")

	i.SetIdentifier(barId)
	i.SetAfterInitFunc(i.init)

	return i
}

func (i *InputBar) init() error {
	i.setStyle()
	i.setKeybindings()
	i.setLayout()

	cpFunc, pasteFunc := util.GetClipboard()
	i.pasteFunc = pasteFunc
	i.SetClipboard(cpFunc, pasteFunc)

	i.handleEvents()

	return nil
}

func (i *InputBar) setLayout() {
	i.SetBorder(true)
	i.SetAutocompleteMaxHeight(12)
}

func (i *InputBar) setStyle() {
	i.SetStyle(i.App.GetStyles())
	i.style = &i.App.GetStyles().InputBar
	i.SetLabelColor(i.style.LabelColor.Color())
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

func (i *InputBar) isInitialState() bool {
	trimmed := strings.TrimSpace(i.GetText())
	if trimmed == "" {
		return true
	}
	noSpaces := strings.ReplaceAll(trimmed, " ", "")
	return noSpaces == "{}" || noSpaces == "[]"
}

func (i *InputBar) setKeybindings() {
	i.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		k := i.App.GetKeys()

		if event.Key() == tcell.KeyCtrlV && i.pasteFunc != nil && i.isInitialState() {
			clipText := strings.TrimSpace(i.pasteFunc())
			if clipText != "" && strings.HasPrefix(clipText, "{") && strings.HasSuffix(clipText, "}") {
				i.SetText(clipText)
				return nil
			}
		}

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

		switch {
		case k.Contains(k.QueryBar.ShowHistory, event.Name()):
			if i.historyModal != nil {
				i.historyModal.Render()
			}
		case k.Contains(k.QueryBar.ClearInput, event.Name()):
			i.SetText("")
			go i.SetWordAtCursor(i.defaultText)
		}

		return event
	})
}

func (i *InputBar) handleEvents() {
	go i.HandleEvents(i.GetIdentifier(), func(event manager.EventMsg) {
		switch event.Message.Type {
		case manager.StyleChanged:
			i.setStyle()
		}

		switch sender := event.Sender; {
		case i.historyModal != nil && sender == i.historyModal.GetIdentifier():
			i.handleHistoryModalEvent(event.EventKey)
		}
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
			i.Close()
			reject()
		case tcell.KeyEnter:
			i.Close()
			text := i.GetText()
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

// braceDepth counts the number of unclosed '{' braces in s.
func braceDepth(s string) int {
	depth := 0
	for _, c := range s {
		if c == '{' {
			depth++
		} else if c == '}' {
			depth--
		}
	}
	return depth
}

// EnableAggregationAutocomplete enables context-aware autocomplete for aggregation stages.
// At the outer level (depth 1) it suggests pipeline stage operators ($match, $group, …).
// Inside the stage value (depth > 1) it suggests general MongoDB operators.
func (i *InputBar) EnableAggregationAutocomplete() {
	stageOperators := mongo.GetAggregationPipelineOperators()
	ma := mongo.NewMongoAutocomplete()

	i.SetAutocompleteFunc(func(currentText string) (entries []tview.AutocompleteItem) {
		currentText = strings.TrimPrefix(currentText, "\"")
		if len(strings.Fields(currentText)) == 0 {
			return nil
		}

		currentWord := i.GetWordAtCursor()
		if strings.HasPrefix(currentWord, "{") || strings.HasPrefix(currentWord, "[") {
			currentWord = currentWord[1:]
		}
		if currentWord == "" {
			return nil
		}

		textBefore := i.GetTextBeforeCursor()
		depth := braceDepth(textBefore)

		escaped := regexp.QuoteMeta(currentWord)

		if depth <= 1 {
			// Outer level: suggest stage operators only
			for _, keyword := range stageOperators {
				if matched, _ := regexp.MatchString("(?i)^"+escaped, keyword.Display); matched {
					entries = append(entries, tview.AutocompleteItem{Main: keyword.Display, Secondary: keyword.Description})
				}
			}
		} else {
			// Inside value: suggest general mongo operators
			for _, keyword := range ma.Operators {
				if matched, _ := regexp.MatchString("(?i)^"+escaped, keyword.Display); matched {
					entries = append(entries, tview.AutocompleteItem{Main: keyword.Display, Secondary: keyword.Description})
				}
			}
			// Also include doc keys if available
			for _, keyword := range i.docKeys {
				if matched, _ := regexp.MatchString("(?i)^"+currentWord, keyword); matched {
					entries = append(entries, tview.AutocompleteItem{Main: keyword})
				}
			}
		}

		return entries
	})

	i.SetAutocompletedFunc(func(text string, index, source int) bool {
		if source == 0 {
			return false
		}
		// Look up in stage operators first, then general operators
		var key *mongo.MongoKeyword
		for idx := range stageOperators {
			if stageOperators[idx].Display == text {
				key = &stageOperators[idx]
				break
			}
		}
		if key == nil {
			key = ma.GetOperatorByDisplay(text)
		}
		if key != nil {
			text = key.InsertText
		}
		i.SetWordAtCursor(text)
		return true
	})
}

// LoadAutocomleteKeys loads new keys for autocomplete
// It is used when switching databases or collections
func (i *InputBar) LoadAutocomleteKeys(keys []string) {
	i.docKeys = keys
}

// SetTextPreserveCursor sets the text of the bar while keeping the cursor
// at its current position. If the cursor is beyond the new text length,
// it ends up at the end naturally via moveCursor clamping.
func (i *InputBar) SetTextPreserveCursor(text string) {
	col := i.GetCursorPosition()
	i.SetText(text)
	i.SetCursorPosition(col)
}

// Open enables the bar and populates it with text.
// If text is empty, it shows the default placeholder instead.
func (i *InputBar) Open(text string) {
	i.Enable()
	if text != "" {
		i.SetTextPreserveCursor(text)
	} else if i.GetText() == "" {
		go i.App.QueueUpdateDraw(func() {
			i.SetWordAtCursor(i.defaultText)
		})
	}
}

// Close disables the bar. Safe to call when already closed.
func (i *InputBar) Close() {
	i.Disable()
}

// Toggle opens the bar with text if closed, or closes it if open.
func (i *InputBar) Toggle(text string) {
	if i.IsEnabled() {
		i.Close()
	} else {
		i.Open(text)
	}
}

func (i *InputBar) handleHistoryModalEvent(eventKey *tcell.EventKey) {
	switch {
	case i.App.GetKeys().Contains(i.App.GetKeys().History.AcceptEntry, eventKey.Name()):
		go i.App.QueueUpdateDraw(func() {
			i.SetText(i.historyModal.GetText())
			i.App.SetFocus(i)
		})
	case i.App.GetKeys().Contains(i.App.GetKeys().History.CloseHistory, eventKey.Name()):
		go i.App.QueueUpdateDraw(func() {
			i.App.SetFocus(i)
		})
	default:
		return
	}
}
