package primitives

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
)

// ListModal is a simple list primitive that is displayed as a modal
type ListModal struct {
	*tview.Box

	list *tview.List
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

	// Calculate the default width of the modal
	width, height := screenWidth/2, screenHeight/2

	// Calculate the position of the popup (centered)
	x := (screenWidth-width)/2 + (screenWidth-width)/16
	y := (screenHeight - height) / 2

	lm.SetRect(x, y, width, height)

	lm.Box.DrawForSubclass(screen, lm)

	// add padding to the list
	x, y, width, height = x+1, y+1, width-2, height-2

	// Set the list's dimensions and position and draw it
	lm.list.SetRect(x, y, width, height)
	lm.list.Draw(screen)
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

// GetCurrentItem returns the index of the currently selected item
func (lm *ListModal) GetCurrentItem() int {
	return lm.list.GetCurrentItem()
}

// Clear removes all items from the list
func (lm *ListModal) Clear() *ListModal {
	lm.list.Clear()
	return lm
}

// ShowSecondaryText sets whether or not secondary text is shown
func (lm *ListModal) ShowSecondaryText(show bool) *ListModal {
	lm.list.ShowSecondaryText(show)
	return lm
}

// SetMainTextStyle sets the text style of main text.
func (lm *ListModal) SetMainTextStyle(style tcell.Style) *ListModal {
	lm.list.SetMainTextStyle(style)
	return lm
}

// SetSecondaryTextStyle sets the text style of secondary text.
func (lm *ListModal) SetSecondaryTextStyle(style tcell.Style) *ListModal {
	lm.list.SetSecondaryTextStyle(style)
	return lm
}

// SetSelectedTextColor sets the color of the selected item's text.
func (lm *ListModal) SetSelectedStyle(style tcell.Style) *ListModal {
	lm.list.SetSelectedStyle(style)
	return lm
}

// InputHandler returns the handler for this primitive.
func (lm *ListModal) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return lm.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		lm.list.InputHandler()(event, setFocus)
	})
}
