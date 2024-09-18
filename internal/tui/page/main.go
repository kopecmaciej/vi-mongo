package page

import (
	"context"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
	"github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/kopecmaciej/vi-mongo/internal/tui/component"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/tui/modal"
	"github.com/rs/zerolog/log"
)

const (
	MainPage = "Main"
)

type Main struct {
	*core.BaseElement
	*core.Flex

	innerFlex *core.Flex
	style     *config.GlobalStyles
	header    *component.Header
	databases *component.Database
	content   *component.Content
}

func NewMain() *Main {
	m := &Main{
		BaseElement: core.NewBaseElement(),
		Flex:        core.NewFlex(),
		innerFlex:   core.NewFlex(),
		header:      component.NewHeader(),
		databases:   component.NewDatabase(),
		content:     component.NewContent(),
	}

	m.SetIdentifier(MainPage)
	m.SetAfterInitFunc(m.init)

	return m
}

func (m *Main) init() error {
	m.setStyles()
	m.setKeybindings()

	m.handleEvents()

	return m.initComponents()
}

func (m *Main) setStyles() {
	m.SetStyle(m.App.GetStyles())
	m.innerFlex.SetStyle(m.App.GetStyles())
	m.innerFlex.SetDirection(tview.FlexRow)
}

func (m *Main) handleEvents() {
	go m.HandleEvents(MainPage, func(event manager.EventMsg) {
		switch event.Message.Type {
		case manager.StyleChanged:
			m.setStyles()
		}
	})
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

func (m *Main) render() error {
	m.Clear()
	m.innerFlex.Clear()

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
		case k.Contains(k.Main.FocusDatabase, event.Name()):
			m.App.SetFocus(m.databases)
			return nil
		case k.Contains(k.Main.FocusContent, event.Name()):
			m.App.SetFocus(m.content)
			return nil
		case k.Contains(k.Main.HideDatabase, event.Name()):
			if _, ok := m.GetItem(0).(*component.Database); ok {
				m.RemoveItem(m.databases)
				m.App.SetFocus(m.content)
			} else {
				m.Clear()
				m.render()
			}
			return nil
		case k.Contains(k.Main.ShowServerInfo, event.Name()):
			m.ShowServerInfoModal()
			return nil
		}
		return event
	})
}

func (m *Main) ShowServerInfoModal() {
	serverInfoModal := modal.NewServerInfoModal(m.Dao)
	if err := serverInfoModal.Init(m.App); err != nil {
		log.Error().Err(err).Msg("Failed to initialize server info modal")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := serverInfoModal.Render(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to render server info modal")
		return
	}

	m.App.Pages.AddPage(modal.ServerInfoModalView, serverInfoModal, true, true)
}
