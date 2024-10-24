package modal

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
)

type Delete struct {
	*core.BaseElement
	*core.Modal

	style *config.OthersStyle
}

func NewDeleteModal(id tview.Identifier) *Delete {
	dm := &Delete{
		BaseElement: core.NewBaseElement(),
		Modal:       core.NewModal(),
	}

	dm.SetIdentifier(id)
	dm.SetAfterInitFunc(dm.init)

	return dm
}

func (d *Delete) init() error {
	d.setLayout()
	d.setStyle()
	d.setKeybindings()

	d.handleEvents()

	return nil
}

func (d *Delete) setLayout() {
	d.AddButtons([]string{"Delete", "Cancel"})
	d.SetBorder(true)
	d.SetTitle(" Delete ")
	d.SetBorderPadding(0, 0, 1, 1)
}

func (d *Delete) setStyle() {
	d.SetStyle(d.App.GetStyles())
	d.style = &d.App.GetStyles().Others

	d.SetButtonActivatedStyle(tcell.StyleDefault.
		Background(d.style.DeleteButtonSelectedBackgroundColor.Color()))
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

func (d *Delete) handleEvents() {
	go d.HandleEvents(d.GetIdentifier(), func(event manager.EventMsg) {
		switch event.Message.Type {
		case manager.StyleChanged:
			d.setStyle()
		}
	})
}
