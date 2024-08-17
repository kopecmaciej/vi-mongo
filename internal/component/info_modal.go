package component

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/rs/zerolog/log"
)

const (
	InfoComponent = "Info"
)

func NewInfoModal(message string) *tview.Modal {
	log.Info().Msg(message)

	message = "[White::b] " + message + " [::]"

	infoModal := tview.NewModal()
	infoModal.SetTitle(" Info ")
	infoModal.SetBorderPadding(0, 0, 1, 1)
	infoModal.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	infoModal.SetTextColor(tcell.ColorGreen)
	infoModal.SetText(message)
	infoModal.AddButtons([]string{"Ok"})

	return infoModal
}

// ShowInfoModal shows a modal with an informational message
func ShowInfoModal(page *Root, message string) {
	infoModal := NewInfoModal(message)

	infoModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Ok" {
			page.RemovePage(InfoComponent)
		}
	})
	page.AddPage(InfoComponent, infoModal, true, true)
}

func ShowInfoModalAndFocus(page *Root, message string, setFocus func()) {
	infoModal := NewInfoModal(message)
	infoModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Ok" {
			page.RemovePage(InfoComponent)
			setFocus()
		}
	})
	page.AddPage(InfoComponent, infoModal, true, true)
}
