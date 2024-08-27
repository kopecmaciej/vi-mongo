package tui

import (
	"context"
	"regexp"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
)

const (
	DatabasesView = "Databases"
	FilterBarView = "FilterBar"
)

// Databases is flex container for DatabaseTree and InputBar
type Databases struct {
	*core.BaseView
	*tview.Flex

	dbTree       *DatabaseTree
	filterBar    *InputBar
	mutex        sync.Mutex
	dbsWithColls []mongo.DBsWithCollections
}

func NewDatabases() *Databases {
	s := &Databases{
		BaseView:  core.NewBaseView(DatabasesView),
		Flex:      tview.NewFlex(),
		dbTree:    NewDatabaseTree(),
		filterBar: NewInputBar(FilterBarView, "Filter"),
		mutex:     sync.Mutex{},
	}

	s.SetAfterInitFunc(s.init)

	return s
}

func (s *Databases) init() error {
	ctx := context.Background()
	s.setStyle()
	s.setKeybindings()

	if err := s.dbTree.Init(s.App); err != nil {
		return err
	}

	if err := s.render(); err != nil {
		return err
	}
	if err := s.renderWithFilter(ctx, ""); err != nil {
		return err
	}
	if err := s.filterBar.Init(s.App); err != nil {
		return err
	}
	s.filterBarListener(ctx)

	return nil
}

func (s *Databases) setStyle() {
	s.Flex.SetDirection(tview.FlexRow)
}

func (s *Databases) setKeybindings() {
	keys := s.App.GetKeys()
	s.Flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case keys.Contains(keys.Root.Databases.FilterBar, event.Name()):
			s.filterBar.Enable()
			s.render()
			return nil
		}
		return event
	})
}

func (s *Databases) render() error {
	s.Flex.Clear()

	var primitive tview.Primitive
	primitive = s.dbTree

	if s.filterBar.IsEnabled() {
		s.Flex.AddItem(s.filterBar, 3, 0, false)
		primitive = s.filterBar
	}
	defer s.App.SetFocus(primitive)

	s.Flex.AddItem(s.dbTree, 0, 1, true)

	return nil
}

func (s *Databases) filterBarListener(ctx context.Context) {
	accceptFunc := func(text string) {
		s.filter(ctx, text)
	}
	rejectFunc := func() {
		s.render()
	}
	s.filterBar.DoneFuncHandler(accceptFunc, rejectFunc)
}

func (s *Databases) filter(ctx context.Context, text string) {
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
}

func (s *Databases) renderWithFilter(ctx context.Context, filter string) error {
	if err := s.fetchDbsWithCollections(ctx, filter); err != nil {
		return err
	}
	s.dbTree.Render(ctx, s.dbsWithColls, false)

	return nil
}

func (s *Databases) fetchDbsWithCollections(ctx context.Context, filter string) error {
	dbsWitColls, err := s.Dao.ListDbsWithCollections(ctx, filter)
	if err != nil {
		return err
	}
	s.dbsWithColls = dbsWitColls

	return nil
}
