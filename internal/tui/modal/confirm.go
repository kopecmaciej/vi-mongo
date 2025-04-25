package modal

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
)

type Confirm struct {
	*core.BaseElement
	*core.Modal

	confirmLabel string
	style        *config.OthersStyle
}

func NewConfirm(id tview.Identifier) *Confirm {
	dm := &Confirm{
		BaseElement:  core.NewBaseElement(),
		Modal:        core.NewModal(),
		confirmLabel: "Confirm",
	}

	dm.SetIdentifier(id)
	dm.SetAfterInitFunc(dm.init)

	return dm
}

func (d *Confirm) init() error {
	d.setLayout()
	d.setStyle()
	d.setKeybindings()

	d.handleEvents()

	return nil
}

func (d *Confirm) setLayout() {
	d.AddButtons([]string{d.confirmLabel, "Cancel"})
	d.SetBorder(true)
	d.SetTitle(" " + d.confirmLabel + " ")
	d.SetBorderPadding(0, 0, 1, 1)
}

func (d *Confirm) setStyle() {
	d.SetStyle(d.App.GetStyles())
	d.style = &d.App.GetStyles().Others

	d.SetButtonActivatedStyle(tcell.StyleDefault.
		Background(d.style.DeleteButtonSelectedBackgroundColor.Color()))
}

func (d *Confirm) setKeybindings() {
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

func (d *Confirm) handleEvents() {
	go d.HandleEvents(d.GetIdentifier(), func(event manager.EventMsg) {
		switch event.Message.Type {
		case manager.StyleChanged:
			d.setStyle()
		}
	})
}

func (d *Confirm) SetConfirmButtonLabel(label string) {
	d.confirmLabel = label
	d.ClearButtons()
	d.AddButtons([]string{label, "Cancel"})
}
