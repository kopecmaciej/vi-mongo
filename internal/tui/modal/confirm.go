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

func (c *Confirm) init() error {
	c.setLayout()
	c.setStyle()
	c.setKeybindings()

	c.handleEvents()

	return nil
}

func (c *Confirm) setLayout() {
	c.AddButtons([]string{c.confirmLabel, "Cancel"})
	c.SetBorder(true)
	c.SetTitle(" " + c.confirmLabel + " ")
	c.SetBorderPadding(0, 0, 1, 1)
}

func (c *Confirm) setStyle() {
	c.SetStyle(c.App.GetStyles())
	c.style = &c.App.GetStyles().Others

	c.SetButtonActivatedStyle(tcell.StyleDefault.
		Background(c.style.DeleteButtonSelectedBackgroundColor.Color()))
}

func (c *Confirm) setKeybindings() {
	c.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'h':
			return tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone)
		case 'l':
			return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
		}
		return event
	})
}

func (c *Confirm) handleEvents() {
	go c.HandleEvents(c.GetIdentifier(), func(event manager.EventMsg) {
		switch event.Message.Type {
		case manager.StyleChanged:
			c.setStyle()
		}
	})
}

func (c *Confirm) SetConfirmButtonLabel(label string) {
	c.confirmLabel = label
	c.ClearButtons()
	c.AddButtons([]string{label, "Cancel"})
}
