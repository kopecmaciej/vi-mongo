package component

import (
	"github.com/kopecmaciej/mongui/manager"
	"github.com/rivo/tview"
)

const (
	ErrorComponent manager.Component = "Error"
)

type Error struct {
	*tview.Modal
}

func NewError() *Error {
	return &Error{
		Modal: tview.NewModal(),
	}
}

func (e *Error) ShowErrorModal(page *Root, message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"Ok"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Ok" {
			page.RemovePage(ErrorComponent)
		}
	})
	page.AddPage(ErrorComponent, modal, true, true)
}
