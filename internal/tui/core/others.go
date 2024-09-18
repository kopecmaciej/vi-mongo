package core

// This file contains all other primitives that for now have only style set
// Once they get more complex, they should be moved to their own file

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/tui/primitives"
)

type (
	Flex struct {
		*tview.Flex
	}
	Form struct {
		*tview.Form
	}
	List struct {
		*tview.List
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
	ViewModal struct {
		*primitives.ViewModal
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

func NewForm() *Form {
	return &Form{
		Form: tview.NewForm(),
	}
}

func (f *Form) SetStyle(style *config.Styles) {
	f.Form.SetBackgroundColor(style.Global.BackgroundColor.Color())
	f.Form.SetBorderColor(style.Global.BorderColor.Color())
	f.Form.SetTitleColor(style.Global.TitleColor.Color())
	f.Form.SetFocusStyle(tcell.StyleDefault.Foreground(style.Global.FocusColor.Color()).Background(style.Global.BackgroundColor.Color()))

	if f.GetButtonCount() > 0 {
		f.SetButtonBackgroundColor(style.Others.ButtonsBackgroundColor.Color())
	}
}

func NewList() *List {
	return &List{
		List: tview.NewList(),
	}
}

func (l *List) SetStyle(style *config.Styles) {
	l.List.SetBackgroundColor(style.Global.BackgroundColor.Color())
	l.List.SetBorderColor(style.Global.BorderColor.Color())
	l.List.SetTitleColor(style.Global.TitleColor.Color())
	l.List.SetFocusStyle(tcell.StyleDefault.Foreground(style.Global.FocusColor.Color()).Background(style.Global.BackgroundColor.Color()))
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

func NewViewModal() *ViewModal {
	return &ViewModal{
		ViewModal: primitives.NewViewModal(),
	}
}

func (v *ViewModal) SetStyle(style *config.Styles) {
	v.ViewModal.SetBackgroundColor(style.Global.BackgroundColor.Color())
	v.ViewModal.SetBorderColor(style.Global.BorderColor.Color())
	v.ViewModal.SetTitleColor(style.Global.TitleColor.Color())
	v.ViewModal.SetFocusStyle(tcell.StyleDefault.Foreground(style.Global.FocusColor.Color()).Background(style.Global.BackgroundColor.Color()))
}
