package tview

import (
	"github.com/rivo/tview"
)

type Flex struct {
	*tview.Flex
}

func NewFlex() *Flex {
	return &Flex{
		Flex: tview.NewFlex(),
	}
}

func (f *tview.Flex) AddItemAtIndex(index int, item tview.Primitive) *Flex {
	items := f.GetItemCount()
	if index > items {

	}
	f.AddItem(item, 0, 1, false)
	return f
}
