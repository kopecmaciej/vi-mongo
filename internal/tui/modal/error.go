package modal

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/rs/zerolog/log"
)

const (
	ErrorModal = "Error"
)

func NewError(message string, err error) *tview.Modal {
	if err != nil {
		log.Error().Err(err).Msg(message)
	}

	message = "[White::b] " + message + " [::]"
	errMsg := err.Error()
	if errMsg != "" {
		message += "\n" + errMsg
	}

	errModal := tview.NewModal()
	errModal.SetTitle(" Error ")
	errModal.SetBorderPadding(0, 0, 1, 1)
	errModal.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	errModal.SetTextColor(tcell.ColorRed)
	errModal.SetText(message)
	errModal.AddButtons([]string{"Ok"})

	return errModal
}

// ShowError shows a modal with an error message
// and logs the error if it's passed
func ShowError(page *core.Pages, message string, err error) {
	errModal := NewError(message, err)

	errModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Ok" {
			page.RemovePage(ErrorModal)
		}
	})
	page.AddPage(ErrorModal, errModal, true, true)
}

func ShowErrorAndSetFocus(page *core.Pages, message string, err error, setFocus func()) {
	errModal := NewError(message, err)
	errModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Ok" {
			page.RemovePage(ErrorModal)
			setFocus()
		}
	})
	page.AddPage(ErrorModal, errModal, true, true)
}
