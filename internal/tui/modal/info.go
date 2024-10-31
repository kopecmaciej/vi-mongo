package modal

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
)

const (
	InfoModalId = "Info"
)

func NewInfo(message string) *tview.Modal {
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

// ShowInfo shows a modal with an informational message
func ShowInfo(page *core.Pages, message string) {
	infoModal := NewInfo(message)

	infoModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Ok" {
			page.RemovePage(InfoModalId)
		}
	})
	page.AddPage(InfoModalId, infoModal, true, true)
}

func ShowInfoModalAndFocus(page *core.Pages, message string, setFocus func()) {
	infoModal := NewInfo(message)
	infoModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Ok" {
			page.RemovePage(InfoModalId)
			setFocus()
		}
	})
	page.AddPage(InfoModalId, infoModal, true, true)
}
