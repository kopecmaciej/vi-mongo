package component

import (
	"github.com/kopecmaciej/mongui/manager"
	"github.com/rivo/tview"
)

const (
	ErrorComponent manager.Component = "Error"
)

func ShowErrorModal(page *Root, message string) {
	errModal := tview.NewModal()
	errModal.SetTitle(" Error ")
	errModal.SetBorderPadding(0, 0, 1, 1)
	errModal.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	errModal.SetText(message)
	errModal.AddButtons([]string{"Ok"})

	errModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Ok" {
			page.RemovePage(ErrorComponent)
		}
	})
	page.AddPage(ErrorComponent, errModal, true, true)
}
