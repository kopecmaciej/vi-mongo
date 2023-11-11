package component

import "github.com/rivo/tview"

type Error struct {
	*tview.Modal
}

func NewError() *Error {
	return &Error{
		Modal: tview.NewModal(),
	}
}

func (e *Error) ShowErrorModal(page *tview.Pages, message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"Ok"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Ok" {
			//remove modal from page
			page.RemovePage("modal")
		}
	})
	page.AddPage("modal", modal, false, false)
	page.ShowPage("modal")
}
