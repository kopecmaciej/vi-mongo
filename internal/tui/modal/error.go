package modal

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/rs/zerolog/log"
)

const (
	ErrorModalId = "Error"
)

func NewError(message string, err error) *tview.Modal {
	taggedMessage := "[White::b] " + message + " [::]"

	if err != nil {
		log.Error().Err(err).Msg(message)

		errMsg := err.Error()
		if errMsg != "" {
			if len(errMsg) > 240 {
				errMsg = errMsg[:240] + " ..."
			}
			taggedMessage += "\n" + errMsg
		}
	}

	errModal := tview.NewModal()
	errModal.SetTitle(" Error ")
	errModal.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
	errModal.SetTextColor(tcell.ColorRed)
	errModal.SetText(taggedMessage)
	errModal.AddButtons([]string{"Ok"})

	return errModal
}

// ShowError shows a modal with an error message
// and logs the error if it's passed
func ShowError(page *core.Pages, message string, err error) {
	errModal := NewError(message, err)

	errModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Ok" {
			page.RemovePage(ErrorModalId)
		}
	})
	page.AddPage(ErrorModalId, errModal, true, true)
}

func ShowErrorAndSetFocus(page *core.Pages, message string, err error, setFocus func()) {
	errModal := NewError(message, err)
	errModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Ok" {
			page.RemovePage(ErrorModalId)
			setFocus()
		}
	})
	page.AddPage(ErrorModalId, errModal, true, true)
}
