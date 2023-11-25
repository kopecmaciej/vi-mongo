package component

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Event string

type SearchBar struct {
	*tview.InputField

	label         string
	InputTextChan chan string
}

// NewSearchBar creates a new filterBar
func NewSearchBar(label string) *SearchBar {
	f := &SearchBar{
		InputField:    tview.NewInputField(),
		label:         "searchBar",
		InputTextChan: make(chan string),
	}

	f.setStyle()

	f.SetLabel(label)
	f.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter || key == tcell.KeyRune {
			f.InputTextChan <- f.GetText()
		}
		if key == tcell.KeyEsc {
			f.InputTextChan <- ""
		}
	})

	return f
}

func (f *SearchBar) setStyle() {
	f.SetBackgroundColor(tcell.ColorDefault)
	f.SetFieldBackgroundColor(tcell.ColorDefault)
	f.SetFieldTextColor(tcell.ColorDefault)
	f.SetLabelColor(tcell.ColorDefault)
}

func (f *SearchBar) SendRune(r rune) {
	f.InputTextChan <- f.GetText()
}
