package component

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Event string

const (
	Rune Event = "rune"
	Text Event = "text"
	Done Event = "done"
)

type SearchBar struct {
	*tview.InputField

	InputTextChan chan string
}

// NewSearchBar creates a new filterBar
func NewSearchBar(label string) *SearchBar {
	f := &SearchBar{
		InputField:    tview.NewInputField(),
		InputTextChan: make(chan string),
	}

	f.SetLabel(label)
	f.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			f.InputTextChan <- f.GetText()
		}
	})

	return f
}

func (f *SearchBar) SendRune(r rune) {
	f.InputTextChan <- f.GetText()
}
