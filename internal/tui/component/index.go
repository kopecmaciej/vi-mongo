package component

import (
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
)

const (
	IndexId = "Index"
)

type Index struct {
	*core.BaseElement
	*core.Flex

	textView *core.TextView
}

func NewIndex() *Index {
	i := &Index{
		BaseElement: core.NewBaseElement(),
		Flex:        core.NewFlex(),
		textView:    core.NewTextView(),
	}

	i.SetIdentifier(IndexId)
	i.SetAfterInitFunc(i.init)

	return i
}

func (i *Index) init() error {
	i.setStyle()

	return nil
}

func (i *Index) setStyle() {
	globalStyle := i.App.GetStyles()
	i.Flex.SetStyle(globalStyle)
	i.textView.SetStyle(globalStyle)
	i.textView.SetTextColor(globalStyle.TabBar.ActiveTextColor.Color())
	i.textView.SetBorder(true)

	i.SetBorderPadding(0, 0, 0, 0)
	i.SetBorder(true)
	i.SetTitle(" Indexes ")
	i.SetTitleAlign(tview.AlignCenter)
}

func (i *Index) Render() {
	i.textView.SetText("To be Implemented")

	i.Flex.AddItem(i.textView, 0, 1, false)
}
