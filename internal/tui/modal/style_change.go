package modal

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
)

const (
	StyleChangeModal = "StyleChangeModal"
)

type StyleChange struct {
	*core.BaseElement
	*core.ListModal

	applyStyle func(styleName string) error
}

func NewStyleChangeModal() *StyleChange {
	sc := &StyleChange{
		BaseElement: core.NewBaseElement(),
		ListModal:   core.NewListModal(),
	}

	sc.SetIdentifier(StyleChangeModal)
	sc.SetAfterInitFunc(sc.init)

	return sc
}

func (sc *StyleChange) init() error {
	sc.setStaticLayout()
	sc.setStyle()
	sc.setKeybindings()

	sc.setContent()

	return nil
}

func (sc *StyleChange) setStaticLayout() {
	sc.SetBorder(true)
	sc.SetTitle(" Change Style ")
	sc.SetBorderPadding(0, 0, 1, 1)
}

func (sc *StyleChange) setStyle() {
	sc.SetStyle(sc.App.GetStyles())
}

func (sc *StyleChange) setKeybindings() {
	sc.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			sc.App.Pages.RemovePage(StyleChangeModal)
			return nil
		case tcell.KeyEnter:
			sc.App.Pages.RemovePage(StyleChangeModal)
			sc.applyStyle(sc.GetText())
			sc.setStyle()
			return nil
		}
		switch event.Rune() {
		case 'l':
			return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
		case 'h':
			return tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone)
		}
		return event
	})
}

func (sc *StyleChange) setContent() {
	allStyles, err := config.GetAllStyles()
	if err != nil {
		ShowError(sc.App.Pages, "Failed to load styles", err)
		return
	}

	for _, style := range allStyles {
		sc.AddItem(style, "", 0, nil)
	}

}

func (sc *StyleChange) SetApplyStyle(applyStyle func(styleName string) error) {
	sc.applyStyle = applyStyle
}

func (sc *StyleChange) Render() {
	sc.App.Pages.AddPage(StyleChangeModal, sc, true, true)
}
