package component

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Event string

type SearchBar struct {
	*tview.InputField

	label   string
	KeyChan chan tcell.Key
}

func NewSearchBar(label string) *SearchBar {
	f := &SearchBar{
		InputField: tview.NewInputField(),
		label:      "searchBar",
		KeyChan:    make(chan tcell.Key),
	}

	f.setStyle()

	f.SetLabel(label + ": ")
	f.SetDoneFunc(func(key tcell.Key) {
		f.KeyChan <- key
	})

	return f
}

func (f *SearchBar) setStyle() {
	f.SetBackgroundColor(tcell.ColorGray)
	f.SetFieldBackgroundColor(tcell.ColorGray)
	f.SetFieldTextColor(tcell.ColorDefault)
	f.SetPlaceholderTextColor(tcell.ColorDefault)
}
