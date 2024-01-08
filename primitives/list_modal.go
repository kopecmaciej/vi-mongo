package primitives

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ListModal is a simple list primitive that is displayed as a modal
type ListModal struct {
	*tview.Box

	list  *tview.List
	label string
}

func NewListModal() *ListModal {
	return &ListModal{
		Box:  tview.NewBox(),
		list: tview.NewList(),
	}
}

// Draw draws this primitive onto the screen.
func (lm *ListModal) Draw(screen tcell.Screen) {
	screenWidth, screenHeight := screen.Size()

	width, height := screenWidth/5, screenHeight/2

	// Calculate the position of the popup (centered)
	x, y := (screenWidth-width)/2, (screenHeight-height)/2

	lm.SetRect(x, y, width, height)

	lm.Box.DrawForSubclass(screen, lm)

	// add padding to the list
	x, y, width, height = x+1, y+1, width-2, height-2

	// Set the list's dimensions and position and draw it
	lm.list.SetRect(x, y, width, height)
	lm.list.Draw(screen)
}

func (lm *ListModal) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return lm.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		lm.list.InputHandler()(event, setFocus)
	})
}

// GetText returns text of the selected item
func (lm *ListModal) GetText() string {
	selected := lm.list.GetCurrentItem()
	mainText, _ := lm.list.GetItemText(selected)
	return mainText
}

// AddItem adds item to the list
func (lm *ListModal) AddItem(text string, secondaryText string, shortcut rune, selected func()) *ListModal {
	lm.list.AddItem(text, secondaryText, shortcut, selected)
	return lm
}

// RemoveItem removes item from the list
func (lm *ListModal) RemoveItem(index int) *ListModal {
	lm.list.RemoveItem(index)
	return lm
}
