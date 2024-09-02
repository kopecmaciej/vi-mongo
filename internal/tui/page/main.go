package page

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/kopecmaciej/vi-mongo/internal/tui/component"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
)

const (
	MainPage = "Main"
)

type Main struct {
	*core.BaseElement
	*tview.Flex

	innerFlex *tview.Flex
	style     *config.GlobalStyles
	header    *component.Header
	databases *component.Databases
	content   *component.Content
}

func NewMain() *Main {
	m := &Main{
		BaseElement: core.NewBaseElement(),
		Flex:        tview.NewFlex(),
		innerFlex:   tview.NewFlex(),
		header:      component.NewHeader(),
		databases:   component.NewDatabases(),
		content:     component.NewContent(),
	}

	m.SetIdentifier(MainPage)
	m.SetAfterInitFunc(m.init)

	return m
}

func (m *Main) init() error {
	m.setStyles()
	m.setKeybindings()

	return m.initComponents()
}

func (m *Main) Render() {
	m.content.Render(false)
	m.databases.Render()
	m.header.Render()

	m.databases.SetSelectFunc(m.content.HandleDatabaseSelection)

	m.render()
}

// UpdateDao updates the dao in the components
func (m *Main) UpdateDao(dao *mongo.Dao) {
	m.databases.UpdateDao(dao)
	m.header.UpdateDao(dao)
	m.content.UpdateDao(dao)
}

func (m *Main) initComponents() error {
	if err := m.header.Init(m.App); err != nil {
		return err
	}
	if err := m.databases.Init(m.App); err != nil {
		return err
	}
	if err := m.content.Init(m.App); err != nil {
		return err
	}
	return nil
}

func (m *Main) setStyles() {
	m.style = &m.App.GetStyles().Global
	m.SetBackgroundColor(m.style.BackgroundColor.Color())
	m.innerFlex.SetBackgroundColor(m.style.BackgroundColor.Color())
}

func (m *Main) render() error {
	m.Clear()
	m.innerFlex.Clear()

	m.innerFlex.SetBackgroundColor(m.style.BackgroundColor.Color())
	m.innerFlex.SetDirection(tview.FlexRow)

	m.AddItem(m.databases, 30, 0, true)
	m.AddItem(m.innerFlex, 0, 7, false)
	m.innerFlex.AddItem(m.header, 4, 0, false)
	m.innerFlex.AddItem(m.content, 0, 7, true)

	m.App.Pages.AddPage(m.GetIdentifier(), m, true, true)
	m.App.SetFocus(m)

	return nil
}

func (m *Main) setKeybindings() {
	k := m.App.GetKeys()
	m.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case k.Contains(k.Main.ToggleFocus, event.Name()):
			if m.App.GetFocus() == m.databases.DbTree {
				m.App.SetFocus(m.content)
			} else {
				m.App.SetFocus(m.databases)
			}
			return nil
		case k.Contains(k.Main.FocusDatabases, event.Name()):
			m.App.SetFocus(m.databases)
			return nil
		case k.Contains(k.Main.FocusContent, event.Name()):
			m.App.SetFocus(m.content)
			return nil
		case k.Contains(k.Main.HideDatabases, event.Name()):
			if _, ok := m.GetItem(0).(*component.Databases); ok {
				m.RemoveItem(m.databases)
				m.App.SetFocus(m.content)
			} else {
				m.Clear()
				m.render()
			}
			return nil
		}
		return event
	})
}
