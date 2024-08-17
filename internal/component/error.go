package component

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/rs/zerolog/log"
)

const (
	ErrorComponent = "Error"
)

func NewErrorModal(message string, err error) *tview.Modal {
	if err != nil {
		log.Error().Err(err).Msg(message)
	}

	message = "[White::b] " + message + " [::]"

	errModal := tview.NewModal()
	errModal.SetTitle(" Error ")
	errModal.SetBorderPadding(0, 0, 1, 1)
	errModal.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	errModal.SetTextColor(tcell.ColorRed)
	errModal.SetText(message)
	errModal.AddButtons([]string{"Ok"})

	return errModal
}

// ShowErrorModal shows a modal with an error message
// and logs the error if it's passed
func ShowErrorModal(page *Root, message string, err error) {
	errModal := NewErrorModal(message, err)

	errModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Ok" {
			page.RemovePage(ErrorComponent)
		}
	})
	page.AddPage(ErrorComponent, errModal, true, true)
}

func ShowErrorModalAndFocus(page *Root, message string, err error, setFocus func()) {
	errModal := NewErrorModal(message, err)
	errModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Ok" {
			page.RemovePage(ErrorComponent)
			setFocus()
		}
	})
	page.AddPage(ErrorComponent, errModal, true, true)
}
