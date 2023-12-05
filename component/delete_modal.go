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

	app     *App
	manager *manager.ComponentManager
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
	d.manager = d.app.ComponentManager

	d.setStyle()
	d.setShortcuts()

	d.AddButtons([]string{"Delete", "Cancel"})

	return nil
}

func (d *DeleteModal) setStyle() {
	d.SetBorder(true)
	d.SetTitle(" Delete ")
	d.SetBorderPadding(0, 0, 1, 1)
}

func (d *DeleteModal) SetText(text string) {
	d.SetText(text)
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

func (d *DeleteModal) SetDoneFunc(handler func(buttonIndex int, buttonLabel string)) {
	d.SetDoneFunc(handler)

	d.app.Root.RemovePage(DeleteModalComponent)
}
