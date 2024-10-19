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
	MainPageId = "Main"
)

type Main struct {
	*core.BaseElement
	*core.Flex

	innerFlex *core.Flex
	style     *config.GlobalStyles
	header    *component.Header
	tabBar    *component.TabBar
	databases *component.Database
	content   *component.Content
	index     *component.Index
	aiPrompt  *component.AIPrompt
}

func NewMain() *Main {
	m := &Main{
		BaseElement: core.NewBaseElement(),
		Flex:        core.NewFlex(),
		innerFlex:   core.NewFlex(),
		header:      component.NewHeader(),
		tabBar:      component.NewTabBar(),
		databases:   component.NewDatabase(),
		content:     component.NewContent(),
		index:       component.NewIndex(),
		aiPrompt:    component.NewAIPrompt(),
	}

	m.SetIdentifier(MainPageId)
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
	go m.HandleEvents(MainPageId, func(event manager.EventMsg) {
		switch event.Message.Type {
		case manager.StyleChanged:
			m.setStyles()
		}
	})
}

func (m *Main) initComponents() error {
	if err := m.header.Init(m.App); err != nil {
		return err
	}

	if err := m.tabBar.Init(m.App); err != nil {
		return err
	}

	if err := m.databases.Init(m.App); err != nil {
		return err
	}
	if err := m.content.Init(m.App); err != nil {
		return err
	}

	if err := m.index.Init(m.App); err != nil {
		return err
	}

	if err := m.aiPrompt.Init(m.App); err != nil {
		return err
	}

	m.tabBar.AddTab("Content", m.content, true)
	m.tabBar.AddTab("Indexes", m.index, false)
	m.tabBar.AddTab("AI", m.aiPrompt, false)

	return nil
}

func (m *Main) Render() {
	m.databases.Render()
	m.header.Render()
	m.tabBar.Render()

	m.databases.SetSelectFunc(func(ctx context.Context, db, coll string) error {
		err := m.content.HandleDatabaseSelection(ctx, db, coll)
		if err != nil {
			return err
		}
		return m.index.HandleDatabaseSelection(ctx, db, coll)
	})

	m.render()
}

// UpdateDao updates the dao in the components
func (m *Main) UpdateDao(dao *mongo.Dao) {
	m.databases.UpdateDao(dao)
	m.header.UpdateDao(dao)
	m.content.UpdateDao(dao)
	m.index.UpdateDao(dao)
}

func (m *Main) render() error {
	m.Clear()
	m.innerFlex.Clear()

	m.AddItem(m.databases, 30, 0, true)
	m.AddItem(m.innerFlex, 0, 7, false)
	m.innerFlex.AddItem(m.header, 4, 0, false)
	m.innerFlex.AddItem(m.tabBar, 1, 0, false)
	m.innerFlex.AddItem(m.tabBar.GetActiveComponentAndRender(), 0, 7, true)

	m.App.Pages.AddPage(m.GetIdentifier(), m, true, true)
	m.App.SetFocus(m)

	return nil
}

func (m *Main) setKeybindings() {
	k := m.App.GetKeys()
	m.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case k.Contains(k.Main.FocusNext, event.Name()):
			// TODO: figure out how to handle key priorities
			if m.index.IsAddFormFocused() || m.aiPrompt.IsAIPromptFocused() {
				return event
			}
			if m.databases.IsFocused() {
				m.App.SetFocus(m.tabBar.GetActiveComponent())
			} else {
				m.innerFlex.RemoveItem(m.tabBar.GetActiveComponent())
				m.tabBar.NextTab()
				m.innerFlex.AddItem(m.tabBar.GetActiveComponentAndRender(), 0, 7, true)

				m.App.SetFocus(m.tabBar.GetActiveComponent())
			}
			return nil
		case k.Contains(k.Main.FocusPrevious, event.Name()):
			if m.index.IsAddFormFocused() || m.aiPrompt.IsAIPromptFocused() {
				return event
			}
			if m.tabBar.GetActiveTabIndex() == 0 {
				m.App.SetFocus(m.databases)
			} else {
				m.innerFlex.RemoveItem(m.tabBar.GetActiveComponent())
				m.tabBar.PreviousTab()
				m.innerFlex.AddItem(m.tabBar.GetActiveComponentAndRender(), 0, 7, true)
				m.App.SetFocus(m.tabBar.GetActiveComponent())
			}
			return nil
		case k.Contains(k.Main.HideDatabase, event.Name()):
			if _, ok := m.GetItem(0).(*component.Database); ok {
				m.RemoveItem(m.databases)
				m.App.SetFocus(m.tabBar.GetActiveComponent())
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

	m.App.Pages.AddPage(modal.ServerInfoModalId, serverInfoModal, true, true)
}
