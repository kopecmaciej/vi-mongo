package primitives

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ModalInput is a simple input field primitive that is displayed as a modal
type ModalInput struct {
	*tview.Box

	input *tview.InputField
	label string
}

// NewModalInput returns a new input field.
func NewModalInput() *ModalInput {
	mi := &ModalInput{
		Box:   tview.NewBox(),
		input: tview.NewInputField(),
	}

	return mi
}

func (mi *ModalInput) Draw(screen tcell.Screen) {
	screenWidth, screenHeight := screen.Size()

	// Calculate the width and height of the popup
	width, height := screenWidth/5, screenHeight/6

	// Calculate the position of the popup (centered)
	x, y := (screenWidth-width)/2, (screenHeight-height)/2

	// Set the position and size of the ModalInput
	mi.Box.SetRect(x, y, width, height)

	// Draw the box for the ModalInput
	mi.Box.DrawForSubclass(screen, mi.input)

	// Adjust the position and size of the input field within the box
	inputX, inputY, inputWidth, _ := mi.GetInnerRect()

	tview.Print(screen, mi.label, inputX, inputY, inputWidth, tview.AlignCenter, tcell.ColorYellow)

  inputY += 3
  inputX = inputX + 2
  inputWidth = inputWidth - 4
	mi.input.SetRect(inputX, inputY, inputWidth, 1)
	mi.input.Draw(screen)
}

func (mi *ModalInput) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return mi.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		mi.input.InputHandler()(event, setFocus)
	})
}

func (mi *ModalInput) SetText(text string) *ModalInput {
	mi.input.SetText(text)
	return mi
}

func (mi *ModalInput) GetText() string {
	return mi.input.GetText()
}

func (mi *ModalInput) SetLabel(label string) *ModalInput {
  mi.label = label
	return mi
}

func (mi *ModalInput) SetInputLabel(label string) *ModalInput {
  mi.input.SetLabel(label)
  return mi
}

func (mi *ModalInput) SetLabelColor(color tcell.Color) *ModalInput {
	mi.input.SetLabelColor(color)
	return mi
}

func (mi *ModalInput) SetFieldBackgroundColor(color tcell.Color) *ModalInput {
	mi.input.SetFieldBackgroundColor(color)
	return mi
}

func (mi *ModalInput) SetFieldTextColor(color tcell.Color) *ModalInput {
	mi.input.SetFieldTextColor(color)
	return mi
}

func (mi *ModalInput) SetBackgroundColor(color tcell.Color) *ModalInput {
	mi.Box.SetBackgroundColor(color)
	return mi
}
