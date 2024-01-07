package component

import (
	"context"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/manager"
	"github.com/rivo/tview"
)

const (
	DeleteModalComponent manager.Component = "DeleteModal"
)

type DeleteModal struct {
	*tview.Modal

	app   *App
}

func NewDeleteModal() *DeleteModal {
	return &DeleteModal{
		Modal: tview.NewModal(),
	}
}

func (d *DeleteModal) Init(ctx context.Context) error {
	app, err := GetApp(ctx)
	if err != nil {
		return err
	}
	d.app = app

	d.setStyle()
	d.setShortcuts()

	d.AddButtons([]string{"[red]Delete", "Cancel"})

	return nil
}

func (d *DeleteModal) setStyle() {
	d.SetBorder(true)
	d.SetTitle(" Delete ")
	d.SetBorderPadding(0, 0, 1, 1)
}

func (d *DeleteModal) setShortcuts() {
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
