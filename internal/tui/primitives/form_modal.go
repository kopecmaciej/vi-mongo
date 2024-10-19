package primitives

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
)

// FormModal is a modal window that contains a form.
type FormModal struct {
	*tview.Box
	form   *tview.Form
	frame  *tview.Frame
	done   func(buttonIndex int, buttonLabel string)
	cancel func()
}

// NewFormModal returns a new form modal.
func NewFormModal() *FormModal {
	m := &FormModal{
		Box: tview.NewBox(),
	}
	m.form = tview.NewForm().
		SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		SetButtonTextColor(tview.Styles.PrimaryTextColor)
	m.form.SetBackgroundColor(tview.Styles.ContrastBackgroundColor).SetBorderPadding(0, 0, 0, 0)
	m.frame = tview.NewFrame(m.form).SetBorders(0, 0, 1, 0, 0, 0)
	m.frame.SetBorder(true).
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor).
		SetBorderPadding(1, 1, 1, 1)

	return m
}

// GetForm returns the form.
func (m *FormModal) GetForm() *tview.Form {
	return m.form
}

// SetDoneFunc sets a handler which is called when one of the buttons was
// pressed. It receives the index of the button as well as its label text.
func (m *FormModal) SetDoneFunc(handler func(buttonIndex int, buttonLabel string)) *FormModal {
	m.done = handler
	return m
}

// SetCancelFunc sets a handler which is called when the user cancels the modal.
func (m *FormModal) SetCancelFunc(handler func()) *FormModal {
	m.cancel = handler
	return m
}

// Draw draws this primitive onto the screen.
func (m *FormModal) Draw(screen tcell.Screen) {
	screenWidth, screenHeight := screen.Size()
	width, height := screenWidth/2, screenHeight/2
	x := (screenWidth - width) / 2
	y := (screenHeight - height) / 2

	m.SetRect(x, y, width, height)
	m.Box.DrawForSubclass(screen, m)

	m.frame.SetRect(x, y, width, height)
	m.frame.Draw(screen)
}

// Focus is called when this primitive receives focus.
func (m *FormModal) Focus(delegate func(p tview.Primitive)) {
	delegate(m.form)
}

// HasFocus returns whether or not this primitive has focus.
func (m *FormModal) HasFocus() bool {
	return m.form.HasFocus()
}

// MouseHandler returns the mouse handler for this primitive.
func (m *FormModal) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return m.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if !m.InRect(event.Position()) {
			return false, nil
		}

		consumed, capture = m.form.MouseHandler()(action, event, setFocus)
		if consumed {
			setFocus(m)
		}
		return
	})
}

// InputHandler returns the handler for this primitive.
func (m *FormModal) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return m.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if event.Key() == tcell.KeyEscape {
			if m.cancel != nil {
				m.cancel()
			}
			return
		}
		if handler := m.form.InputHandler(); handler != nil {
			handler(event, setFocus)
		}
	})
}
