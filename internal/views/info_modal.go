package view

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/internal/views/core"
	"github.com/kopecmaciej/tview"
	"github.com/rs/zerolog/log"
)

const (
	InfoView = "Info"
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
func ShowInfoModal(page *core.Pages, message string) {
	infoModal := NewInfoModal(message)

	infoModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Ok" {
			page.RemovePage(InfoView)
		}
	})
	page.AddPage(InfoView, infoModal, true, true)
}

func ShowInfoModalAndFocus(page *core.Pages, message string, setFocus func()) {
	infoModal := NewInfoModal(message)
	infoModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Ok" {
			page.RemovePage(InfoView)
			setFocus()
		}
	})
	page.AddPage(InfoView, infoModal, true, true)
}
