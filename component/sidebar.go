package component

import (
	"context"
	"regexp"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/manager"
	"github.com/kopecmaciej/mongui/mongo"
	"github.com/kopecmaciej/tview"
	"github.com/rs/zerolog/log"
)

const (
	SideBarComponent   manager.Component = "SideBar"
	FilterBarComponent manager.Component = "FilterBar"
)

type SideBar struct {
	*Component
	*tview.Flex

	dbTree       *DBTree
	filterBar    *InputBar
	mutex        sync.Mutex
	dbsWithColls []mongo.DBsWithCollections
}

func NewSideBar() *SideBar {
	s := &SideBar{
		Component: NewComponent(SideBarComponent),
		Flex:      tview.NewFlex(),
		dbTree:    NewDBTree(),
		filterBar: NewInputBar(string(FilterBarComponent)),
		mutex:     sync.Mutex{},
	}

	s.SetAfterInitFunc(s.init)

	return s
}

func (s *SideBar) init() error {
	ctx := context.Background()
	s.setStyle()
	s.setKeybindings()

	if err := s.dbTree.Init(s.app); err != nil {
		return err
	}

	if err := s.render(); err != nil {
		return err
	}
	if err := s.renderWithFilter(ctx, ""); err != nil {
		return err
	}
	if err := s.filterBar.Init(s.app); err != nil {
		return err
	}
	s.filterBarListener(ctx)

	return nil
}

func (s *SideBar) setStyle() {
	s.Flex.SetDirection(tview.FlexRow)
}

func (s *SideBar) setKeybindings() {
	keys := s.app.Keys
	s.Flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case keys.Contains(keys.Root.Sidebar.FilterBar, event.Name()):
			s.filterBar.Enable()
			s.render()
			return nil
		}
		return event
	})
}

func (s *SideBar) render() error {
	s.Flex.Clear()

	var primitive tview.Primitive
	primitive = s.dbTree

	if s.filterBar.IsEnabled() {
		s.Flex.AddItem(s.filterBar, 3, 0, false)
		primitive = s.filterBar
	}
	defer s.app.SetFocus(primitive)

	s.Flex.AddItem(s.dbTree, 0, 1, true)

	return nil
}

func (s *SideBar) filterBarListener(ctx context.Context) {
	accceptFunc := func(text string) {
		s.filter(ctx, text)
	}
	rejectFunc := func() {
		s.render()
	}
	s.filterBar.DoneFuncHandler(accceptFunc, rejectFunc)
}

func (s *SideBar) filter(ctx context.Context, text string) {
	defer s.render()
	dbsWitColls := s.dbsWithColls
	expand := false
	filtered := []mongo.DBsWithCollections{}
	if text == "" {
		return
	}
	for _, db := range dbsWitColls {
		re := regexp.MustCompile(`(?i)` + text)
		if re.MatchString(db.DB) {
			contain := false
			for _, f := range filtered {
				if f.DB == db.DB {
					contain = true
					break
				}
			}
			if !contain {
				filtered = append(filtered, db)
			}
		}
		for _, coll := range db.Collections {
			if re.MatchString(coll) {
				contain := false
				for _, f := range filtered {
					if f.DB == db.DB {
						f.Collections = append(f.Collections, coll)
						contain = true
						break
					}
				}
				if !contain {
					expand = true
					filtered = append(filtered, mongo.DBsWithCollections{
						DB:          db.DB,
						Collections: []string{coll},
					})
				}
			}
		}
	}
	s.dbTree.Render(ctx, filtered, expand)

	log.Debug().Msgf("Filtered: %v", filtered)
}

func (s *SideBar) renderWithFilter(ctx context.Context, filter string) error {
	if err := s.fetchDbsWithCollections(ctx, filter); err != nil {
		return err
	}
	s.dbTree.Render(ctx, s.dbsWithColls, false)

	return nil
}

func (s *SideBar) fetchDbsWithCollections(ctx context.Context, filter string) error {
	dbsWitColls, err := s.dao.ListDbsWithCollections(ctx, filter)
	if err != nil {
		return err
	}
	s.dbsWithColls = dbsWitColls

	return nil
}
