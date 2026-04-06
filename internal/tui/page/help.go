package page

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
)

const (
	HelpPageId = "Help"
)

// sectionOrder defines the preferred display order for key sections.
// Sections absent from this list are appended at the end.
var sectionOrder = []string{
	"Navigation", "Global", "Help", "Connection",
	"Main", "Databases", "FilterBar", "Content",
	"Peeker", "QueryBar", "SortBar", "Index", "AIQuery", "History", "Aggregation",
}

// Help is a view that provides a searchable, two-panel help screen for keybindings.
type Help struct {
	*core.BaseElement
	*core.Flex

	style     *config.HelpStyle
	leftFlex  *core.Flex
	rightFlex *core.Flex

	sectionList    *core.List
	keysTable      *core.Table
	hintView       *core.TextView
	searchInput    *core.InputField
	capturePanel   *tview.Flex
	captureDisplay *tview.TextView
	capturedKey    config.Key

	allSections      []config.OrderedKeys
	filteredSections []config.OrderedKeys
	searchMode       bool
	editMode         bool
	editSectionIdx   int
	editKeyIdx       int
}

func NewHelp() *Help {
	h := &Help{
		BaseElement:    core.NewBaseElement(),
		Flex:           core.NewFlex(),
		leftFlex:       core.NewFlex(),
		rightFlex:      core.NewFlex(),
		sectionList:    core.NewList(),
		keysTable:      core.NewTable(),
		hintView:       core.NewTextView(),
		searchInput:    core.NewInputField(),
		capturePanel:   tview.NewFlex(),
		captureDisplay: tview.NewTextView(),
	}

	h.SetIdentifier(HelpPageId)
	h.SetAfterInitFunc(h.init)

	return h
}

func (h *Help) init() error {
	h.setLayout()
	h.setStyle()
	h.setKeybindings()
	h.handleEvents()
	return nil
}

func (h *Help) handleEvents() {
	go h.HandleEvents(HelpPageId, func(event manager.EventMsg) {
		switch event.Message.Type {
		case manager.StyleChanged:
			h.setStyle()
			go h.App.QueueUpdateDraw(func() {
				h.Render()
			})
		}
	})
}

func (h *Help) setLayout() {
	h.Flex.SetBorder(true)
	h.Flex.SetTitle(" Help ")
	h.Flex.SetTitleAlign(tview.AlignLeft)
	h.Flex.SetDirection(tview.FlexRow)

	h.leftFlex.SetDirection(tview.FlexRow)
	h.rightFlex.SetDirection(tview.FlexRow)

	h.sectionList.SetTitle(" Sections ")
	h.sectionList.SetBorder(true)
	h.sectionList.ShowSecondaryText(false)
	h.sectionList.SetBorderPadding(0, 0, 1, 0)

	h.keysTable.SetTitle(" Keys ")
	h.keysTable.SetBorder(true)
	h.keysTable.SetBorderPadding(0, 0, 1, 1)
	h.keysTable.SetSelectable(true, false)
	h.keysTable.SetScrollBarEnabled(true)
	h.keysTable.SetEvaluateAllRows(true)

	h.searchInput.SetLabel(" / ")
	h.searchInput.SetBorder(true)

	captureHint := tview.NewTextView()
	captureHint.SetDynamicColors(true)
	captureHint.SetText(" [::d]any key=add  Enter=save  Esc=cancel  Backspace=clear[-:-:-]")

	h.captureDisplay.SetDynamicColors(true)
	h.captureDisplay.SetText(" [::d]Press a key combination to bind...[-:-:-]")

	h.capturePanel.SetBorder(true)
	h.capturePanel.SetDirection(tview.FlexRow)
	h.capturePanel.AddItem(h.captureDisplay, 1, 0, true)
	h.capturePanel.AddItem(captureHint, 1, 0, false)

	h.hintView.SetTextAlign(tview.AlignCenter)
	h.hintView.SetDynamicColors(true)

	contentFlex := tview.NewFlex()
	contentFlex.AddItem(h.leftFlex, 28, 0, true)
	contentFlex.AddItem(h.rightFlex, 0, 1, false)

	h.leftFlex.AddItem(h.sectionList, 0, 1, true)
	h.rightFlex.AddItem(h.keysTable, 0, 1, false)

	h.Flex.AddItem(contentFlex, 0, 1, true)
	h.Flex.AddItem(h.hintView, 1, 0, false)
}

func (h *Help) setStyle() {
	h.style = &h.App.GetStyles().Help
	h.SetStyle(h.App.GetStyles())
	h.leftFlex.SetStyle(h.App.GetStyles())
	h.rightFlex.SetStyle(h.App.GetStyles())
	h.sectionList.SetStyle(h.App.GetStyles())
	h.keysTable.SetStyle(h.App.GetStyles())
	h.hintView.SetStyle(h.App.GetStyles())
	h.searchInput.SetStyle(h.App.GetStyles())

	s := h.App.GetStyles()
	h.capturePanel.SetBackgroundColor(s.Global.BackgroundColor.Color())
	h.capturePanel.SetBorderColor(s.Global.BorderColor.Color())
	h.capturePanel.SetTitleColor(s.Global.TitleColor.Color())
	h.captureDisplay.SetBackgroundColor(s.Global.BackgroundColor.Color())
	h.captureDisplay.SetTextColor(s.Global.TextColor.Color())

	textColor := h.App.GetStyles().Global.TextColor.Color()
	globalBg := h.App.GetStyles().Global.BackgroundColor.Color()
	selectedFg := h.style.SelectedTextColor.Color()
	selectedBg := h.style.SelectedBackgroundColor.Color()
	h.sectionList.SetMainTextStyle(tcell.StyleDefault.
		Foreground(textColor).
		Background(globalBg))
	h.sectionList.SetSelectedStyle(tcell.StyleDefault.
		Foreground(selectedFg).
		Background(selectedBg))

	h.keysTable.SetSelectedStyle(tcell.StyleDefault.
		Foreground(selectedFg).
		Background(selectedBg))

	h.keysTable.SetScrollBarStyle(
		tcell.StyleDefault.Foreground(h.style.ScrollBarThumbColor.Color()),
		tcell.StyleDefault.Foreground(h.style.ScrollBarTrackColor.Color()),
	)
}

func (h *Help) setKeybindings() {
	k := h.App.GetKeys()

	h.sectionList.SetChangedFunc(func(index int, _, _ string, _ rune) {
		h.renderKeysForSection(index)
	})

	h.sectionList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case k.Contains(k.Help.Close, event.Name()):
			h.App.Pages.RemovePage(HelpPageId)
			return nil
		case k.Contains(k.Main.FocusNext, event.Name()):
			h.App.SetFocusInternal(h.keysTable)
			return nil
		case k.Contains(k.Help.Search, event.Name()):
			h.enterSearchMode()
			return nil
		case k.Contains(k.Navigation.MoveDown, event.Name()):
			curr := h.sectionList.GetCurrentItem()
			h.sectionList.SetCurrentItem(curr + 1)
			return nil
		case k.Contains(k.Navigation.MoveUp, event.Name()):
			if curr := h.sectionList.GetCurrentItem(); curr > 0 {
				h.sectionList.SetCurrentItem(curr - 1)
			}
			return nil
		}
		return event
	})

	h.keysTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case k.Contains(k.Help.Close, event.Name()):
			h.App.Pages.RemovePage(HelpPageId)
			return nil
		case k.Contains(k.Main.FocusPrevious, event.Name()):
			h.App.SetFocusInternal(h.sectionList)
			return nil
		case k.Contains(k.Help.EditKey, event.Name()):
			row, _ := h.keysTable.GetSelection()
			h.enterEditMode(row)
			return nil
		case k.Contains(k.Navigation.MoveDown, event.Name()):
			row, _ := h.keysTable.GetSelection()
			if row < h.keysTable.GetRowCount()-1 {
				h.keysTable.Select(row+1, 0)
			}
			return nil
		case k.Contains(k.Navigation.MoveUp, event.Name()):
			if row, _ := h.keysTable.GetSelection(); row > 0 {
				h.keysTable.Select(row-1, 0)
			}
			return nil
		}
		return event
	})

	h.searchInput.SetDoneFunc(func(key tcell.Key) {
		h.exitSearchMode(key == tcell.KeyEsc)
	})

	h.searchInput.SetChangedFunc(func(text string) {
		h.filterSections(text)
	})

	h.captureDisplay.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			h.exitEditMode()
		case tcell.KeyEnter:
			h.saveEdit()
		case tcell.KeyBackspace, tcell.KeyBackspace2, tcell.KeyDelete:
			h.capturedKey = config.Key{}
			h.updateCaptureDisplay()
		default:
			captured := eventKeyToConfigKey(event)
			if captured.Keys != nil || captured.Runes != nil {
				h.capturedKey.Keys = append(h.capturedKey.Keys, captured.Keys...)
				h.capturedKey.Runes = append(h.capturedKey.Runes, captured.Runes...)
				h.updateCaptureDisplay()
			}
		}
		return nil
	})
}

func (h *Help) enterSearchMode() {
	h.searchMode = true
	h.leftFlex.AddItem(h.searchInput, 3, 0, true)
	h.App.SetFocus(h.searchInput)
}

func (h *Help) exitSearchMode(reset bool) {
	h.searchMode = false
	h.leftFlex.RemoveItem(h.searchInput)
	if reset {
		h.searchInput.SetText("")
		h.filteredSections = h.allSections
		h.renderSectionList(0)
	}
	h.App.SetFocusInternal(h.sectionList)
}

func (h *Help) enterEditMode(row int) {
	sectionIdx := h.sectionList.GetCurrentItem()
	if sectionIdx >= len(h.filteredSections) {
		return
	}
	section := h.filteredSections[sectionIdx]
	if row < 0 || row >= len(section.Keys) {
		return
	}

	h.editMode = true
	h.editSectionIdx = sectionIdx
	h.editKeyIdx = row
	h.capturedKey = config.Key{}

	desc := section.Keys[row].Description
	h.capturePanel.SetTitle(fmt.Sprintf(" Editing: %s ", desc))
	h.updateCaptureDisplay()

	h.rightFlex.RemoveItem(h.keysTable)
	h.rightFlex.AddItem(h.capturePanel, 4, 0, false)
	h.rightFlex.AddItem(h.keysTable, 0, 1, false)
	h.App.SetFocusInternal(h.captureDisplay)
}

func (h *Help) exitEditMode() {
	h.editMode = false
	h.rightFlex.RemoveItem(h.capturePanel)
	h.App.SetFocusInternal(h.keysTable)
}

func (h *Help) saveEdit() {
	if h.capturedKey.Keys == nil && h.capturedKey.Runes == nil {
		h.exitEditMode()
		return
	}
	if h.editSectionIdx >= len(h.filteredSections) {
		h.exitEditMode()
		return
	}
	section := h.filteredSections[h.editSectionIdx]
	if h.editKeyIdx >= len(section.Keys) {
		h.exitEditMode()
		return
	}

	newKey := config.Key{
		Keys:        h.capturedKey.Keys,
		Runes:       h.capturedKey.Runes,
		Description: section.Keys[h.editKeyIdx].Description,
	}

	kb := h.App.GetKeys()
	if err := kb.SetKeyAt(section.Element, h.editKeyIdx, newKey); err != nil {
		h.exitEditMode()
		return
	}
	if err := kb.SaveKeybindings(); err != nil {
		h.exitEditMode()
		return
	}

	row := h.editKeyIdx
	h.exitEditMode()

	h.keysTable.SetCell(row, 0,
		tview.NewTableCell(formatHelpKeyString(newKey)).SetTextColor(h.style.KeyColor.Color()))
	h.keysTable.Select(row, 0)
}

func (h *Help) updateCaptureDisplay() {
	var parts []string
	parts = append(parts, h.capturedKey.Keys...)
	parts = append(parts, h.capturedKey.Runes...)
	if len(parts) == 0 {
		h.captureDisplay.SetText(" [::d]Press a key combination to bind...[-:-:-]")
	} else {
		h.captureDisplay.SetText(fmt.Sprintf(" [::b]► %s[-:-:-]", strings.Join(parts, ", ")))
	}
}

// eventKeyToConfigKey converts a tcell key event to a config.Key entry.
func eventKeyToConfigKey(event *tcell.EventKey) config.Key {
	var key config.Key

	if event.Key() == tcell.KeyRune {
		if event.Modifiers()&tcell.ModAlt != 0 {
			key.Keys = []string{"Alt+" + string(event.Rune())}
		} else {
			key.Runes = []string{string(event.Rune())}
		}
		return key
	}

	name, ok := tcell.KeyNames[event.Key()]
	if !ok || name == "" {
		return key
	}
	// tcell uses "Ctrl-L" (uppercase); config convention is "Ctrl+l" (lowercase)
	if strings.HasPrefix(name, "Ctrl-") && len(name) == 6 {
		name = "Ctrl+" + strings.ToLower(string(name[5]))
	}
	key.Keys = []string{name}
	return key
}

func (h *Help) filterSections(query string) {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		h.filteredSections = h.allSections
	} else {
		h.filteredSections = nil
		for _, s := range h.allSections {
			if strings.Contains(strings.ToLower(s.Element), query) {
				h.filteredSections = append(h.filteredSections, s)
				continue
			}
			for _, key := range s.Keys {
				if strings.Contains(strings.ToLower(key.Description), query) {
					h.filteredSections = append(h.filteredSections, s)
					break
				}
			}
		}
	}
	h.renderSectionList(0)
	if len(h.filteredSections) > 0 {
		h.renderKeysForSection(0)
	} else {
		h.keysTable.Clear()
	}
}

func (h *Help) Render() {
	allKeys := h.App.GetKeys().GetAvaliableKeys()
	h.allSections = h.sortAndFilter(allKeys)
	h.filteredSections = h.allSections

	h.renderSectionList(0)
	if len(h.filteredSections) > 0 {
		h.renderKeysForSection(0)
	}
	h.renderHints()
}

func (h *Help) renderHints() {
	k := h.App.GetKeys()
	dim := h.App.GetStyles().Global.TextColor.Color()
	accent := h.App.GetStyles().Global.FocusColor.Color()

	dimHex := fmt.Sprintf("#%06x", dim.Hex())
	accentHex := fmt.Sprintf("#%06x", accent.Hex())

	hint := func(key, desc string) string {
		return fmt.Sprintf("[%s]%s[-] [%s]%s[-]", accentHex, key, dimHex, desc)
	}

	h.hintView.SetText(strings.Join([]string{
		hint(k.Navigation.MoveUp.String(), "up"),
		hint(k.Navigation.MoveDown.String(), "down"),
		hint(k.Main.FocusNext.String(), "→ panel"),
		hint(k.Main.FocusPrevious.String(), "← panel"),
		hint(k.Help.Search.String(), "search"),
		hint(k.Help.EditKey.String(), "edit key"),
		hint(k.Help.Close.String(), "close"),
	}, "  "))
}

func (h *Help) sortAndFilter(keys []config.OrderedKeys) []config.OrderedKeys {
	orderIndex := make(map[string]int, len(sectionOrder))
	for i, name := range sectionOrder {
		orderIndex[name] = i
	}

	known := make([]config.OrderedKeys, len(sectionOrder))
	var unknown []config.OrderedKeys
	for _, ok := range keys {
		if len(ok.Keys) == 0 {
			continue
		}
		if idx, exists := orderIndex[ok.Element]; exists {
			known[idx] = ok
		} else {
			unknown = append(unknown, ok)
		}
	}

	var result []config.OrderedKeys
	for _, ok := range known {
		if ok.Element != "" {
			result = append(result, ok)
		}
	}
	return append(result, unknown...)
}

func (h *Help) renderSectionList(selectIdx int) {
	h.sectionList.Clear()
	for _, s := range h.filteredSections {
		h.sectionList.AddItem(s.Element, "", 0, nil)
	}
	if len(h.filteredSections) > 0 {
		if selectIdx >= len(h.filteredSections) {
			selectIdx = 0
		}
		h.sectionList.SetCurrentItem(selectIdx)
	}
}

func (h *Help) renderKeysForSection(idx int) {
	h.keysTable.Clear()
	if idx >= len(h.filteredSections) {
		return
	}
	section := h.filteredSections[idx]
	for row, key := range section.Keys {
		keyString := formatHelpKeyString(key)
		h.keysTable.SetCell(row, 0,
			tview.NewTableCell(keyString).SetTextColor(h.style.KeyColor.Color()))
		h.keysTable.SetCell(row, 1,
			tview.NewTableCell(key.Description).SetTextColor(h.style.DescriptionColor.Color()))
	}
	h.keysTable.ScrollToBeginning()
	if h.keysTable.GetRowCount() > 0 {
		h.keysTable.Select(0, 0)
	}
}

func formatHelpKeyString(key config.Key) string {
	var parts []string
	if len(key.Keys) > 0 {
		parts = append(parts, strings.Join(key.Keys, ", "))
	}
	if len(key.Runes) > 0 {
		parts = append(parts, strings.Join(key.Runes, ", "))
	}
	return strings.Join(parts, ", ")
}
