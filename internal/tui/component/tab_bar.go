package component

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
)

const (
	TabBarId = "TabBar"
)

type TabBarPrimitive interface {
	tview.Primitive
	Render()
}

type TabBarComponent struct {
	id        string
	component TabBarPrimitive
}

type TabBar struct {
	*core.BaseElement
	*core.Table

	active int
	tabs   []*TabBarComponent
}

func NewTabBar() *TabBar {
	t := &TabBar{
		BaseElement: core.NewBaseElement(),
		Table:       core.NewTable(),
		tabs:        []*TabBarComponent{},
	}

	t.SetIdentifier(TabBarId)
	t.SetAfterInitFunc(t.init)

	return t
}

func (t *TabBar) init() error {
	t.setStaticLayout()
	t.setStyle()

	t.handleEvents()
	return nil
}

func (t *TabBar) setStyle() {
	styles := t.App.GetStyles()
	t.SetStyle(styles)
}

func (t *TabBar) setStaticLayout() {
	t.SetBorderPadding(0, 0, 1, 0)
}

func (t *TabBar) handleEvents() {
	go t.HandleEvents(TabBarId, func(event manager.EventMsg) {
		switch event.Message.Type {
		case manager.StyleChanged:
			t.setStyle()
			t.Render()
		}
	})
}

func (t *TabBar) AddTab(name string, component TabBarPrimitive, defaultTab bool) {
	t.tabs = append(t.tabs, &TabBarComponent{
		id:        name,
		component: component,
	})
	if defaultTab {
		t.active = len(t.tabs) - 1
	}
	t.Render()
}

func (t *TabBar) NextTab() {
	t.active = (t.active + 1) % len(t.tabs)
	t.Render()
}

func (t *TabBar) PreviousTab() {
	t.active = (t.active - 1 + len(t.tabs)) % len(t.tabs)
	t.Render()
}

func (t *TabBar) Render() {
	styles := t.App.GetStyles()
	t.Clear()
	for i, tab := range t.tabs {
		cell := tview.NewTableCell(" " + tab.id + " ")
		if i == t.active {
			cell.SetTextColor(styles.TabBar.ActiveTextColor.Color())
			cell.SetAttributes(tcell.AttrBold)
			cell.SetBackgroundColor(styles.TabBar.ActiveBackgroundColor.Color())

		} else {
			cell.SetTextColor(styles.TabBar.TextColor.Color())
		}
		t.SetCell(0, i, cell)
	}
}

func (t *TabBar) GetActiveComponent() TabBarPrimitive {
	return t.tabs[t.active].component
}

func (t *TabBar) GetActiveTabIndex() int {
	return t.active
}
