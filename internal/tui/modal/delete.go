package modal

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
)

const (
	DeleteModal = "DeleteModal"
)

type Delete struct {
	*core.BaseElement
	*tview.Modal

	style *config.OthersStyle
}

func NewDeleteModal() *Delete {
	dm := &Delete{
		BaseElement: core.NewBaseElement(),
		Modal:       tview.NewModal(),
	}

	dm.SetIdentifier(DeleteModal)
	dm.SetAfterInitFunc(dm.init)

	return dm
}

func (d *Delete) init() error {
	d.setStyle()
	d.setKeybindings()

	d.AddButtons([]string{"[red]Delete", "Cancel"})

	return nil
}

func (d *Delete) setStyle() {
	d.style = &d.App.GetStyles().Others

	d.SetBorder(true)
	d.SetTitle(" Delete ")
	d.SetBorderPadding(0, 0, 1, 1)
	d.SetButtonTextColor(d.style.ButtonsTextColor.Color())
	d.SetButtonBackgroundColor(d.style.ButtonsBackgroundColor.Color())
}

func (d *Delete) setKeybindings() {
	d.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'h':
			return tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone)
		case 'l':
			return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
		}
		return event
	})
}
