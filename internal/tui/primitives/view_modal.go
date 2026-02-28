package primitives

// TODO!!!: Rethink the way of handling those margins and padddings
// as they are becoming really hard to manage

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/util"
)

// Text is the text to be displayed in the modal.
type Text struct {
	Content string
	Color   tcell.Color
	Align   int
}

// ViewModal is a centered message window used to inform the user or prompt them
// for an immediate decision. It needs to have at least one button (added via
// [ViewModal.AddButtons]) or it will never disappear.
type ViewModal struct {
	*tview.Box
	// The frame embedded in the modal.
	frame *tview.Frame

	// The form embedded in the modal's frame.
	form *tview.Form

	// The text to be displayed in the modal.
	text Text

	// Whether or not the modal is scrollable
	scrollable bool

	// The position of the scroll
	scrollPosition int

	// The end position of the scroll
	endPosition int

	// The optional callback for when the user clicked one of the buttons. It
	// receives the index of the clicked button and the button's label.
	done func(buttonIndex int, buttonLabel string)

	// The selected line index
	selectedLine int

	// The color of the highlighted text.
	highlightColor tcell.Color

	// The colors for document elements.
	keyColor, valueColor, bracketColor, arrayColor tcell.Color

	// The margin of the modal (only top and bottom)
	marginTop, marginBottom int

	// Whether the view is in full-screen mode
	isFullScreen bool

	// Optional key handler injected by the parent component to handle
	// navigation keys via the keybindings config system
	keyHandler func(event *tcell.EventKey) *tcell.EventKey
}

// NewViewModal returns a new modal message window.
func NewViewModal() *ViewModal {
	m := &ViewModal{
		Box: tview.NewBox(),
		text: Text{
			Color: tview.Styles.PrimaryTextColor,
			Align: tview.AlignLeft,
		},
		scrollable:     true,
		scrollPosition: 0,
		marginTop:      6,
		marginBottom:   6,
		isFullScreen:   false,
	}
	m.form = tview.NewForm().
		SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		SetButtonTextColor(tview.Styles.PrimaryTextColor)
	m.form.SetBackgroundColor(tview.Styles.ContrastBackgroundColor).SetBorderPadding(0, 0, 0, 0)
	m.form.SetCancelFunc(func() {
		if m.done != nil {
			m.done(-1, "")
		}
	})
	m.frame = tview.NewFrame(m.form).SetBorders(0, 0, 1, 0, 0, 0)
	m.frame.SetBorder(true).
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor).
		SetBorderPadding(1, 1, 1, 1)

	return m
}

// SetBackgroundColor sets the color of the modal frame background.
func (m *ViewModal) SetBackgroundColor(color tcell.Color) *tview.Box {
	m.Box.SetBackgroundColor(color)
	m.form.SetBackgroundColor(color)
	m.frame.SetBackgroundColor(color)
	return m.Box
}

// SetTextColor sets the color of the message text.
func (m *ViewModal) SetTextColor(color tcell.Color) *tview.Box {
	m.text.Color = color
	return m.Box
}

// SetBorderColor sets the color of the modal frame border.
func (m *ViewModal) SetBorderColor(color tcell.Color) *tview.Box {
	m.Box.SetBorderColor(color)
	m.frame.SetBorderColor(color)
	return m.Box
}

// SetTitleColor sets the color of the modal frame title.
func (m *ViewModal) SetTitleColor(color tcell.Color) *tview.Box {
	m.Box.SetTitleColor(color)
	m.frame.SetTitleColor(color)
	return m.Box
}

// SetFocusStyle sets the style of the modal when it is focused.
func (m *ViewModal) SetFocusStyle(style tcell.Style) *tview.Box {
	m.Box.SetFocusStyle(style)
	m.frame.SetFocusStyle(style)
	return m.Box
}

// SetButtonBackgroundColor sets the background color of the buttons.
func (m *ViewModal) SetButtonBackgroundColor(color tcell.Color) *ViewModal {
	m.form.SetButtonBackgroundColor(color)
	return m
}

// SetButtonTextColor sets the color of the button texts.
func (m *ViewModal) SetButtonTextColor(color tcell.Color) *ViewModal {
	m.form.SetButtonTextColor(color)
	return m
}

// SetHighlightColor sets the color of the highlighted text.
func (m *ViewModal) SetHighlightColor(color tcell.Color) *ViewModal {
	m.highlightColor = color
	return m
}

// SetDocumentColors sets the colors for document elements.
func (m *ViewModal) SetDocumentColors(keyColor, valueColor, bracketColor tcell.Color) *ViewModal {
	m.keyColor = keyColor
	m.valueColor = valueColor
	m.bracketColor = bracketColor
	return m
}

// SetButtonStyle sets the style of the buttons when they are not focused.
func (m *ViewModal) SetButtonStyle(style tcell.Style) *ViewModal {
	m.form.SetButtonStyle(style)
	return m
}

// SetButtonActivatedStyle sets the style of the buttons when they are focused.
func (m *ViewModal) SetButtonActivatedStyle(style tcell.Style) *ViewModal {
	m.form.SetButtonActivatedStyle(style)
	return m
}

// SetDoneFunc sets a handler which is called when one of the buttons was
// pressed. It receives the index of the button as well as its label text. The
// handler is also called when the user presses the Escape key. The index will
// then be negative and the label text an empty string.
func (m *ViewModal) SetDoneFunc(handler func(buttonIndex int, buttonLabel string)) *ViewModal {
	m.done = handler
	return m
}

// AddButtons adds buttons to the window. There must be at least one button and
// a "done" handler so the window can be closed again.
func (m *ViewModal) AddButtons(labels []string) *ViewModal {
	for index, label := range labels {
		func(i int, l string) {
			m.form.AddButton(label, func() {
				if m.done != nil {
					m.done(i, l)
				}
			})
			button := m.form.GetButton(m.form.GetButtonCount() - 1)
			button.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				switch event.Key() {
				case tcell.KeyRune:
					switch event.Rune() {
					case 'h':
						return tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone)
					case 'l':
						return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
					}
				}
				return event
			})
		}(index, label)
	}
	return m
}

// ClearButtons removes all buttons from the window.
func (m *ViewModal) ClearButtons() *ViewModal {
	m.form.ClearButtons()
	return m
}

// SetFocus shifts the focus to the button with the given index.
func (m *ViewModal) SetFocus(index int) *ViewModal {
	m.form.SetFocus(index)
	return m
}

// Focus is called when this primitive receives focus.
func (m *ViewModal) Focus(delegate func(p tview.Primitive)) {
	delegate(m.form)
}

// HasFocus returns whether or not this primitive has focus.
func (m *ViewModal) HasFocus() bool {
	return m.form.HasFocus()
}

// SetFullScreen toggles between full-screen and modal view
func (m *ViewModal) SetFullScreen(fullScreen bool) *ViewModal {
	m.isFullScreen = fullScreen
	return m
}

// IsFullScreen returns true if the modal is in full-screen mode
func (m *ViewModal) IsFullScreen() bool {
	return m.isFullScreen
}

func (m *ViewModal) Draw(screen tcell.Screen) {
	screenWidth, screenHeight := screen.Size()

	// Calculate the width and position of this modal
	var width, x, y int
	if m.isFullScreen {
		width = screenWidth
		x, y = 0, 0
	} else {
		width = screenWidth / 2
		x = (screenWidth - width) / 2
		y = (screenHeight - (screenHeight - m.marginTop - m.marginBottom)) / 2
	}

	// Reset the text and find out how wide it is
	m.frame.Clear()
	lines := tview.WordWrap(m.text.Content, width)

	maxLines := len(lines)
	if !m.isFullScreen {
		if maxLines > screenHeight-m.marginTop-m.marginBottom {
			maxLines = screenHeight - m.marginTop - m.marginBottom
		}
	} else {
		if maxLines > screenHeight {
			maxLines = screenHeight - m.marginBottom
		}
	}

	m.endPosition = maxLines

	// Calculate the total height and the starting line based on scroll position
	totalHeight := len(lines)
	startLine := m.scrollPosition
	if startLine > totalHeight-maxLines {
		startLine = totalHeight - maxLines
	}
	if startLine < 0 {
		startLine = 0
	}

	numNextLinesToHighlight := m.calculateNextLinesToHighlight(lines)
	for i := startLine; i < startLine+maxLines && i < totalHeight; i++ {
		lines[i] = m.formatAndColorizeLine(lines[i], i == startLine)

		if i-startLine == m.selectedLine {
			lines[i] = m.highlightLine(lines[i], true)
			absolutePosition := m.scrollPosition + m.selectedLine
			for j := absolutePosition + 1; j <= absolutePosition+numNextLinesToHighlight && j < len(lines); j++ {
				lines[j] = m.highlightLine(lines[j], false)
			}
		} else {
			lines[i] = " " + lines[i]
		}

		m.frame.AddText(lines[i], true, m.text.Align, m.text.Color)
	}

	height := maxLines + m.marginBottom

	// Set the rect for the modal
	m.SetRect(x, y, width, height)

	if m.isFullScreen {
		height = screenHeight
	}
	// Draw the frame
	m.frame.SetRect(x, y, width, height)
	m.frame.Draw(screen)
}

func calculateIndentation(line string) int {
	return len(line) - len(strings.TrimLeft(line, " \t"))
}

// calculateNextLinesToHighlight calculates the number of lines to highlight after the selected line.
// so total number of lines to highlight is numNextLinesToHighlight + 1 (plus the selected line)
// It's purly based on indentations and does not consider actual json structure.
func (m *ViewModal) calculateNextLinesToHighlight(lines []string) int {
	absolutePosition := m.selectedLine + m.scrollPosition
	if absolutePosition >= len(lines) {
		return 0
	}

	currentLine := lines[absolutePosition]
	currentIndent := calculateIndentation(currentLine)
	linesToHighlight := 0

	keyRegex := regexp.MustCompile(`^\s*"[^"]*":`)
	objInArrayStartRegex := regexp.MustCompile(`^\s*\{`)
	objInArrayEndRegex := regexp.MustCompile(`^\s*(\}|\},)\s*$`)

	// TODO: This should be probably so complicated
	switch {
	// first line and last line
	case absolutePosition == 0:
		return len(lines) - 1
	// last lines
	case absolutePosition == len(lines)-1:
		return linesToHighlight
	// if next identation is lesser, it means that current line is the last one in the block
	// so we need to highlight only current line
	case calculateIndentation(lines[absolutePosition+1]) <
		// only exception is wrapped text which has no indentation so we need to check for that
		currentIndent && calculateIndentation(lines[absolutePosition+1]) > 0:
		return 0
	case objInArrayStartRegex.MatchString(currentLine):
		for i := absolutePosition + 1; i < len(lines); i++ {
			if calculateIndentation(lines[i]) == currentIndent && objInArrayEndRegex.MatchString(lines[i]) {
				return linesToHighlight + 1
			}
			linesToHighlight++
		}
	case keyRegex.MatchString(currentLine):
		for i := absolutePosition + 1; i < len(lines); i++ {
			if i == len(lines)-1 {
				return i - absolutePosition - 1
			}
			if calculateIndentation(lines[i]) == currentIndent {
				if strings.HasSuffix(lines[i], "},") || strings.HasSuffix(lines[i], "}") {
					return linesToHighlight + 1
				}
				if strings.HasSuffix(lines[i], "],") || strings.HasSuffix(lines[i], "]") {
					return linesToHighlight + 1
				}
				return linesToHighlight
			}
			linesToHighlight++
		}

	default:
		return 0
	}

	return linesToHighlight
}

func (m *ViewModal) formatAndColorizeLine(line string, isFirstLine bool) string {
	if isFirstLine && strings.TrimSpace(line) == "{" {
		return fmt.Sprintf("[%s]{[%s]", m.bracketColor.CSS(), m.valueColor.CSS())
	}

	if strings.Contains(line, "{") || strings.Contains(line, "}") {
		line = strings.ReplaceAll(line, "{", fmt.Sprintf("[%s]{", m.bracketColor.CSS()))
		line = strings.ReplaceAll(line, "}", fmt.Sprintf("[%s]}[%s]", m.bracketColor.CSS(), m.valueColor.String()))
	}

	re := regexp.MustCompile(`"([^"]+)":(.*)`)
	line = re.ReplaceAllStringFunc(line, func(s string) string {
		matches := re.FindStringSubmatch(s)
		if len(matches) > 2 {
			key := matches[1]
			value := strings.TrimSpace(matches[2])
			return fmt.Sprintf("[%s]\"%s\"[:]: [%s]%s[-]", m.keyColor.CSS(), key, m.valueColor.CSS(), value)
		}
		return s
	})

	return line
}

func (m *ViewModal) highlightLine(line string, withMark bool) string {
	if withMark {
		return fmt.Sprintf("[-:%s:b]>%s[-:-:-]", m.highlightColor.CSS(), line)
	}
	return fmt.Sprintf("[-:%s:b]%s[-:-:-]", m.highlightColor.CSS(), line)
}

func (m *ViewModal) MoveUp() {
	if m.selectedLine > 0 {
		m.selectedLine--
	} else if m.scrollPosition > 0 {
		if m.scrollPosition == 0 {
			return
		}
		m.scrollPosition--
	}
}

func (m *ViewModal) MoveDown() {
	_, _, width, height := m.GetRect()
	maxLines := height - m.marginBottom
	totalLines := len(tview.WordWrap(m.text.Content, width))

	// sometimes totalLines are incorrect, to short (when key:value is multilines at the end),
	// to fix that we need to recalculate it based on the content
	if totalLines < maxLines {
		totalLines = maxLines
	}

	if m.selectedLine < maxLines-1 && m.selectedLine < totalLines-1 {
		m.selectedLine++
	} else if m.selectedLine < totalLines-1 && m.scrollPosition+m.selectedLine < totalLines-1 {
		m.scrollPosition++
	}
}

func (m *ViewModal) MoveToTop() {
	m.scrollPosition = 0
	m.selectedLine = 0
}

func (m *ViewModal) MoveToBottom() {
	_, _, width, height := m.GetRect()
	maxLines := height - m.marginBottom
	lines := tview.WordWrap(m.text.Content, width)
	totalLines := len(lines)

	// same as in MoveDown, but for bottom
	if totalLines < maxLines {
		totalLines = maxLines
	}

	if totalLines > maxLines {
		m.scrollPosition = totalLines - maxLines
		m.selectedLine = maxLines - 1
	} else {
		m.scrollPosition = 0
		m.selectedLine = totalLines - 1
	}
}

// TextAlignment sets the text alignment within the modal. This must be one of
func (m *ViewModal) SetText(text Text) *ViewModal {
	m.text = text
	return m
}
func (m *ViewModal) TextAlignment(align int) *ViewModal {
	m.text.Align = align
	return m
}

// MouseHandler returns the mouse handler for this primitive.
func (m *ViewModal) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return m.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		// Pass mouse events on to the form.
		consumed, capture = m.form.MouseHandler()(action, event, setFocus)
		if !consumed && action == tview.MouseLeftDown && m.InRect(event.Position()) {
			setFocus(m)
			consumed = true
		}
		return
	})
}

// SetKeyHandler sets a callback that handles key events for navigation.
// The handler should return nil if the event was consumed, or the event
// itself if it should be passed through to the default handling.
func (m *ViewModal) SetKeyHandler(handler func(event *tcell.EventKey) *tcell.EventKey) {
	m.keyHandler = handler
}

// InputHandler returns the handler for this primitive.
func (m *ViewModal) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return m.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if m.keyHandler != nil {
			if m.keyHandler(event) == nil {
				return
			}
		}

		if m.frame.HasFocus() {
			if handler := m.frame.InputHandler(); handler != nil {
				handler(event, setFocus)
				return
			}
		}
	})
}

func (m *ViewModal) SetScrollable(scrollable bool) *ViewModal {
	m.scrollable = scrollable
	return m
}

// CopySelectedLine copies the selected line to the clipboard.
// copyType can be "full" or "value". "full" will copy the entire highlighted lines,
// while "value" will copy only the value of the highlighted line.
func (m *ViewModal) CopySelectedLine(copyFunc func(text string) error, copyType string) error {
	_, _, width, _ := m.GetRect()
	lines := tview.WordWrap(m.text.Content, width)
	selectedLineIndex := m.scrollPosition + m.selectedLine

	if selectedLineIndex >= 0 && selectedLineIndex < len(lines) {
		numNextLinesToHighlight := m.calculateNextLinesToHighlight(lines)
		highlightedLines := lines[selectedLineIndex : selectedLineIndex+numNextLinesToHighlight+1]

		var textToCopy string
		switch copyType {
		case "full":
			textToCopy = strings.Join(highlightedLines, "\n")
			textToCopy = util.CleanJsonWhitespaces(textToCopy)
		case "value":
			// Join all highlighted lines
			textToCopy = strings.Join(highlightedLines, "\n")
			textToCopy = util.CleanJsonWhitespaces(textToCopy)
			// If it's an object inside we are just removing { }
			bracketRegex := regexp.MustCompile(`^\s*(\{.*?\})\s*$`)
			if bracketRegex.MatchString(strings.TrimSpace(textToCopy)) {
				textToCopy = strings.TrimSuffix(textToCopy, "}")
				textToCopy = strings.TrimPrefix(textToCopy, "{")
			} else {
				// Split by the first colon to separate key and value
				parts := strings.SplitN(textToCopy, ":", 2)
				if len(parts) > 1 {
					// Trim spaces
					textToCopy = strings.TrimSpace(parts[1])
				} else {
					textToCopy = strings.TrimSpace(textToCopy)
				}
			}

		default:
			textToCopy = strings.Join(highlightedLines, "\n")
		}

		textToCopy = strings.TrimSpace(textToCopy)
		return copyFunc(textToCopy)
	}
	return nil
}
