package view

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/internal/config"
	"github.com/kopecmaciej/mongui/internal/views/core"
	"github.com/kopecmaciej/tview"
)

const (
	DeleteModalView = "DeleteModal"
)

type DeleteModal struct {
	*core.BaseView
	*tview.Modal

	style *config.OthersStyle
}

func NewDeleteModal() *DeleteModal {
	dm := &DeleteModal{
		BaseView: core.NewBaseView(DeleteModalView),
		Modal:    tview.NewModal(),
	}

	dm.SetAfterInitFunc(dm.init)

	return dm
}

func (d *DeleteModal) init() error {
	d.setStyle()
	d.setKeybindings()

	d.AddButtons([]string{"[red]Delete", "Cancel"})

	return nil
}

func (d *DeleteModal) setStyle() {
	d.style = &d.App.GetStyles().Others

	d.SetBorder(true)
	d.SetTitle(" Delete ")
	d.SetBorderPadding(0, 0, 1, 1)
	d.SetButtonTextColor(d.style.ButtonsTextColor.Color())
	d.SetButtonBackgroundColor(d.style.ButtonsBackgroundColor.Color())
}

func (d *DeleteModal) setKeybindings() {
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
