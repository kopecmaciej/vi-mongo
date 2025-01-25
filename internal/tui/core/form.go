package core

import (
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
)

type Form struct {
	*tview.Form
}

func NewForm() *Form {
	return &Form{
		Form: tview.NewForm(),
	}
}

func (f *Form) SetStyle(style *config.Styles) {
	SetCommonStyle(f.Form, style)
	if f.GetButtonCount() > 0 {
		f.SetButtonBackgroundColor(style.Others.ButtonsBackgroundColor.Color())
		f.SetButtonTextColor(style.Others.ButtonsTextColor.Color())
	}
}

// This function will not include buttons, so if there are any
// should be added separatly
func (f *Form) InsertFormItem(pos int, item tview.FormItem) *Form {
	count := f.GetFormItemCount()
	if pos < 0 || pos > count {
		pos = count
	}

	existingItems := make([]tview.FormItem, count)
	for i := 0; i < count; i++ {
		existingItems[i] = f.GetFormItem(i)
	}

	f.Clear(true)
	for i := 0; i < pos; i++ {
		f.AddFormItem(existingItems[i])
	}
	f.AddFormItem(item)
	for i := pos; i < count; i++ {
		f.AddFormItem(existingItems[i])
	}

	return f
}
