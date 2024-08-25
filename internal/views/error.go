package view

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/internal/views/core"
	"github.com/kopecmaciej/tview"
	"github.com/rs/zerolog/log"
)

const (
	ErrorView = "Error"
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
func ShowErrorModal(page *core.Pages, message string, err error) {
	errModal := NewErrorModal(message, err)

	errModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Ok" {
			page.RemovePage(ErrorView)
		}
	})
	page.AddPage(ErrorView, errModal, true, true)
}

func ShowErrorModalAndFocus(page *core.Pages, message string, err error, setFocus func()) {
	errModal := NewErrorModal(message, err)
	errModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Ok" {
			page.RemovePage(ErrorView)
			setFocus()
		}
	})
	page.AddPage(ErrorView, errModal, true, true)
}
