package primitives

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
)

// Text is the text to be displayed in the modal.
type Text struct {
	Content string
	Color   tcell.Color
	Align   int
}

// ModalView is a centered message window used to inform the user or prompt them
// for an immediate decision. It needs to have at least one button (added via
// [ModalView.AddButtons]) or it will never disappear.
type ModalView struct {
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
	keyColor, valueColor, bracketColor tcell.Color
}

// NewModalView returns a new modal message window.
func NewModalView() *ModalView {
	m := &ModalView{
		Box: tview.NewBox(),
		text: Text{
			Color: tview.Styles.PrimaryTextColor,
			Align: tview.AlignLeft,
		},
		scrollable:     true,
		scrollPosition: 0,
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
func (m *ModalView) SetBackgroundColor(color tcell.Color) *ModalView {
	m.form.SetBackgroundColor(color)
	m.frame.SetBackgroundColor(color)
	return m
}

// SetTextColor sets the color of the message text.
func (m *ModalView) SetTextColor(color tcell.Color) *ModalView {
	m.text.Color = color
	return m
}

// SetButtonBackgroundColor sets the background color of the buttons.
func (m *ModalView) SetButtonBackgroundColor(color tcell.Color) *ModalView {
	m.form.SetButtonBackgroundColor(color)
	return m
}

// SetButtonTextColor sets the color of the button texts.
func (m *ModalView) SetButtonTextColor(color tcell.Color) *ModalView {
	m.form.SetButtonTextColor(color)
	return m
}

// SetButtonStyle sets the style of the buttons when they are not focused.
func (m *ModalView) SetButtonStyle(style tcell.Style) *ModalView {
	m.form.SetButtonStyle(style)
	return m
}

// SetButtonActivatedStyle sets the style of the buttons when they are focused.
func (m *ModalView) SetButtonActivatedStyle(style tcell.Style) *ModalView {
	m.form.SetButtonActivatedStyle(style)
	return m
}

// SetDoneFunc sets a handler which is called when one of the buttons was
// pressed. It receives the index of the button as well as its label text. The
// handler is also called when the user presses the Escape key. The index will
// then be negative and the label text an empty string.
func (m *ModalView) SetDoneFunc(handler func(buttonIndex int, buttonLabel string)) *ModalView {
	m.done = handler
	return m
}

// AddButtons adds buttons to the window. There must be at least one button and
// a "done" handler so the window can be closed again.
func (m *ModalView) AddButtons(labels []string) *ModalView {
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
				case tcell.KeyDown, tcell.KeyRight:
					return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
				case tcell.KeyUp, tcell.KeyLeft:
					return tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone)
				}
				return event
			})
		}(index, label)
	}
	return m
}

// ClearButtons removes all buttons from the window.
func (m *ModalView) ClearButtons() *ModalView {
	m.form.ClearButtons()
	return m
}

// SetFocus shifts the focus to the button with the given index.
func (m *ModalView) SetFocus(index int) *ModalView {
	m.form.SetFocus(index)
	return m
}

// Focus is called when this primitive receives focus.
func (m *ModalView) Focus(delegate func(p tview.Primitive)) {
	delegate(m.form)
}

// HasFocus returns whether or not this primitive has focus.
func (m *ModalView) HasFocus() bool {
	return m.form.HasFocus()
}

func (m *ModalView) Draw(screen tcell.Screen) {
	// Calculate the width of this modal.
	screenWidth, screenHeight := screen.Size()
	width := screenWidth / 3

	// Reset the text and find out how wide it is.
	m.frame.Clear()
	lines := tview.WordWrap(m.text.Content, width)

	// Variables for scrolling
	maxLines := screenHeight - 12

	m.endPosition = len(lines) - maxLines

	// Calculate the total height and the starting line based on scroll position
	totalHeight := len(lines)
	startLine := m.scrollPosition
	if startLine > totalHeight-maxLines {
		startLine = totalHeight - maxLines
	}
	if startLine < 0 {
		startLine = 0
	}

	numLinesToHighlight := m.calculateNextLinesToHighlight(lines)

	for i := startLine; i < startLine+maxLines && i < totalHeight; i++ {
		lines[i] = m.formatLine(lines[i], i == startLine)

		if i-startLine == m.selectedLine {
			lines[i] = m.highlightLine(lines[i], true)
			for j := m.selectedLine + 1; j <= m.selectedLine+numLinesToHighlight; j++ {
				lines[j] = m.highlightLine(lines[j], false)
			}
		} else {
			lines[i] = " " + lines[i]
		}

		m.frame.AddText(lines[i], true, m.text.Align, m.text.Color)
	}

	// Set the modal's position and size.
	height := maxLines + 6
	width += 4
	x := (screenWidth - width) / 2
	y := (screenHeight - height) / 2
	m.SetRect(x, y, width, height)

	// Draw the frame.
	m.frame.SetRect(x, y, width, height)
	m.frame.Draw(screen)
}

func calculateIndent(line string) int {
	return len(line) - len(strings.TrimLeft(line, " \t"))
}

func (m *ModalView) calculateNextLinesToHighlight(lines []string) int {
	currentIndent := calculateIndent(lines[m.selectedLine])
	linesToHighlight := 0

	// TODO: this is not clear way to handle highlighting, but for now it works
	for i := m.selectedLine + 1; i < len(lines); i++ {
		nextIndent := calculateIndent(lines[i])

		if lines[i] == "}" && nextIndent == 0 {
			return linesToHighlight
		}

		// highlight till the end if first line
		if m.selectedLine == 0 {
			return len(lines) - 1
		}

		// if current indent is 0, highlight only given line
		if currentIndent == 0 {
			return linesToHighlight
		}

		// Case 1: Same indent, new key:value pair
		if nextIndent == currentIndent {
			return linesToHighlight
			// Case 2: Wrapped line, continue highlighting
		} else if nextIndent > 0 && nextIndent < currentIndent {
			return linesToHighlight
		} else if nextIndent == 0 {
			linesToHighlight++
			// Case 3: Object or array, continue until we find matching indent
		} else if nextIndent > currentIndent {
			linesToHighlight++
			for j := i + 1; j < len(lines); j++ {
				if calculateIndent(lines[j]) == currentIndent {
					return j - m.selectedLine
				}
			}
		}
	}

	return linesToHighlight
}

func (m *ModalView) formatLine(line string, isFirstLine bool) string {
	if isFirstLine && strings.TrimSpace(line) == "{" {
		return fmt.Sprintf("[%s]{[%s]", m.bracketColor.String(), m.valueColor.String())
	}

	if strings.Contains(line, "{") || strings.Contains(line, "}") {
		line = strings.ReplaceAll(line, "{", fmt.Sprintf("[%s]{", m.bracketColor.String()))
		line = strings.ReplaceAll(line, "}", fmt.Sprintf("[%s]}[%s]", m.bracketColor.String(), m.valueColor.String()))
	}

	re := regexp.MustCompile(`"([^"]+)":(.*)`)
	line = re.ReplaceAllStringFunc(line, func(s string) string {
		matches := re.FindStringSubmatch(s)
		if len(matches) > 2 {
			key := matches[1]
			value := strings.TrimSpace(matches[2])
			return fmt.Sprintf("[%s]\"%s\"[:]: [%s]%s[-]", m.keyColor.String(), key, m.valueColor.String(), value)
		}
		return s
	})

	return line
}

func (m *ModalView) highlightLine(line string, withMark bool) string {
	if withMark {
		return fmt.Sprintf("[-:%s:b]>%s[-:-:-]", m.highlightColor.String(), line)
	}
	return fmt.Sprintf("[-:%s:b]%s[-:-:-]", m.highlightColor.String(), line)
}

// Additional methods to handle scrolling
func (m *ModalView) ScrollUp() {
	if m.scrollPosition == 0 {
		return
	}
	m.scrollPosition--
}

func (m *ModalView) ScrollDown() {
	if m.scrollPosition == m.endPosition {
		return
	}
	m.scrollPosition++
}

// TextAlignment sets the text alignment within the modal. This must be one of
func (m *ModalView) SetText(text Text) *ModalView {
	m.text = text
	return m
}
func (m *ModalView) TextAlignment(align int) *ModalView {
	m.text.Align = align
	return m
}

// MouseHandler returns the mouse handler for this primitive.
func (m *ModalView) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
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

// InputHandler returns the handler for this primitive.
func (m *ModalView) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return m.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		key := event.Key()

		switch key {
		case tcell.KeyDown:
			m.MoveDown()
		case tcell.KeyUp:
			m.MoveUp()
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				m.MoveDown()
			case 'k':
				m.MoveUp()
			case 'g':
				m.scrollPosition = 0
				m.selectedLine = 0
			case 'G':
				m.scrollPosition = m.endPosition
				_, _, width, _ := m.GetRect()
				m.selectedLine = len(tview.WordWrap(m.text.Content, width-6))
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

func (m *ModalView) SetScrollable(scrollable bool) *ModalView {
	m.scrollable = scrollable
	return m
}

func (m *ModalView) MoveUp() {
	if m.selectedLine > 0 {
		m.selectedLine--
	} else if m.scrollPosition > 0 {
		m.ScrollUp()
	}
}

func (m *ModalView) MoveDown() {
	_, _, width, height := m.GetRect()
	maxLines := height - 6
	totalLines := len(tview.WordWrap(m.text.Content, width-4))

	if m.selectedLine < maxLines-1 && m.selectedLine < totalLines-1 {
		m.selectedLine++
	} else if m.scrollPosition < m.endPosition {
		m.ScrollDown()
	}
}

func (m *ModalView) CopySelectedLine(copyFunc func(text string) error, copyType string) error {
	_, _, width, _ := m.GetRect()
	lines := tview.WordWrap(m.text.Content, width-4)
	selectedLineIndex := m.scrollPosition + m.selectedLine

	if selectedLineIndex >= 0 && selectedLineIndex < len(lines) {
		numLinesToHighlight := m.calculateNextLinesToHighlight(lines)
		highlightedLines := lines[selectedLineIndex : selectedLineIndex+numLinesToHighlight+1]

		var textToCopy string
		switch copyType {
		case "full":
			textToCopy = strings.Join(highlightedLines, "\n")
			textToCopy = strings.TrimSuffix(textToCopy, ",")
			textToCopy = strings.ReplaceAll(textToCopy, " ", "")
			textToCopy = strings.ReplaceAll(textToCopy, "{", "{ ")
			textToCopy = strings.ReplaceAll(textToCopy, "}", " }")
		case "value":
			for _, line := range highlightedLines {
				if parts := strings.SplitN(line, ":", 2); len(parts) > 1 {
					textToCopy += strings.TrimSpace(parts[1]) + "\n"
				} else {
					textToCopy += strings.TrimSpace(line) + "\n"
				}
			}
		default:
			textToCopy = strings.Join(highlightedLines, "\n")
		}

		return copyFunc(strings.TrimSpace(textToCopy))
	}
	return nil
}

// SetHighlightColor sets the color of the highlighted text.
func (m *ModalView) SetHighlightColor(color tcell.Color) *ModalView {
	m.highlightColor = color
	return m
}

// SetDocumentColors sets the colors for document elements.
func (m *ModalView) SetDocumentColors(keyColor, valueColor, bracketColor tcell.Color) *ModalView {
	m.keyColor = keyColor
	m.valueColor = valueColor
	m.bracketColor = bracketColor
	return m
}
