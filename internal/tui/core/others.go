package core

// This file contains all other primitives that for now have only style set
// Once they get more complex, they should be moved to their own file

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
)

type (
	Flex struct {
		*tview.Flex
	}
	TextView struct {
		*tview.TextView
	}
	TreeView struct {
		*tview.TreeView
	}
	InputField struct {
		*tview.InputField
	}
)

func NewFlex() *Flex {
	return &Flex{
		Flex: tview.NewFlex(),
	}
}

func (f *Flex) SetStyle(style *config.Styles) {
	f.Flex.SetBackgroundColor(style.Global.BackgroundColor.Color())
	f.Flex.SetBorderColor(style.Global.BorderColor.Color())
	f.Flex.SetTitleColor(style.Global.TitleColor.Color())
	f.Flex.SetFocusStyle(tcell.StyleDefault.Foreground(style.Global.FocusColor.Color()).Background(style.Global.BackgroundColor.Color()))
}

func NewTextView() *TextView {
	return &TextView{
		TextView: tview.NewTextView(),
	}
}

func (t *TextView) SetStyle(style *config.Styles) {
	t.TextView.SetBackgroundColor(style.Global.BackgroundColor.Color())
	t.TextView.SetBorderColor(style.Global.BorderColor.Color())
	t.TextView.SetTitleColor(style.Global.TitleColor.Color())
	t.TextView.SetFocusStyle(tcell.StyleDefault.Foreground(style.Global.FocusColor.Color()).Background(style.Global.BackgroundColor.Color()))
}

func NewTreeView() *TreeView {
	return &TreeView{
		TreeView: tview.NewTreeView(),
	}
}

func (t *TreeView) SetStyle(style *config.Styles) {
	t.TreeView.SetBackgroundColor(style.Global.BackgroundColor.Color())
	t.TreeView.SetBorderColor(style.Global.BorderColor.Color())
	t.TreeView.SetTitleColor(style.Global.TitleColor.Color())
	t.TreeView.SetFocusStyle(tcell.StyleDefault.Foreground(style.Global.FocusColor.Color()).Background(style.Global.BackgroundColor.Color()))
}

func NewInputField() *InputField {
	return &InputField{
		InputField: tview.NewInputField(),
	}
}

func (i *InputField) SetStyle(style *config.Styles) {
	i.InputField.SetBackgroundColor(style.Global.BackgroundColor.Color())
	i.InputField.SetBorderColor(style.Global.BorderColor.Color())
	i.InputField.SetTitleColor(style.Global.TitleColor.Color())
	i.InputField.SetFocusStyle(tcell.StyleDefault.Foreground(style.Global.FocusColor.Color()).Background(style.Global.BackgroundColor.Color()))
}
