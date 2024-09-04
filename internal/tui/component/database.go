package component

import (
	"context"
	"regexp"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/tui/modal"
)

const (
	DatabaseComponent = "Database"
	FilterBarView     = "FilterBar"
)

// Database is flex container for DatabaseTree and InputBar
type Database struct {
	*core.BaseElement
	*tview.Flex

	DbTree       *DatabaseTree
	filterBar    *InputBar
	mutex        sync.Mutex
	dbsWithColls []mongo.DBsWithCollections
}

func NewDatabase() *Database {
	s := &Database{
		BaseElement: core.NewBaseElement(),
		Flex:        tview.NewFlex(),
		DbTree:      NewDatabaseTree(),
		filterBar:   NewInputBar(FilterBarView, "Filter"),
		mutex:       sync.Mutex{},
	}

	s.SetIdentifier(DatabaseComponent)
	s.SetAfterInitFunc(s.init)

	return s
}

func (s *Database) init() error {
	ctx := context.Background()
	s.setStyle()
	s.setKeybindings()

	if err := s.DbTree.Init(s.App); err != nil {
		return err
	}

	if err := s.filterBar.Init(s.App); err != nil {
		return err
	}
	s.filterBarHandler(ctx)

	return nil
}

func (s *Database) setStyle() {
	s.Flex.SetDirection(tview.FlexRow)
}

func (s *Database) setKeybindings() {
	keys := s.App.GetKeys()
	s.Flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case keys.Contains(keys.Database.FilterBar, event.Name()):
			s.filterBar.Enable()
			s.Render()
			return nil
		}
		return event
	})
}

func (s *Database) Render() {
	s.Flex.Clear()

	var primitive tview.Primitive
	primitive = s.DbTree

	if s.filterBar.IsEnabled() {
		s.Flex.AddItem(s.filterBar, 3, 0, false)
		primitive = s.filterBar
	}
	defer s.App.SetFocus(primitive)

	if err := s.listDbsAndCollections(context.Background()); err != nil {
		modal.ShowError(s.App.Pages, "Failed to list databases and collections", err)
		return
	}

	s.DbTree.Render(context.Background(), s.dbsWithColls, false)

	s.Flex.AddItem(s.DbTree, 0, 1, true)
}

func (s *Database) filterBarHandler(ctx context.Context) {
	accceptFunc := func(text string) {
		s.filter(ctx, text)
	}
	rejectFunc := func() {
		s.Render()
	}
	s.filterBar.DoneFuncHandler(accceptFunc, rejectFunc)
}

func (s *Database) filter(ctx context.Context, text string) {
	defer s.Render()
	dbsWitColls := s.dbsWithColls
	expand := false
	filtered := []mongo.DBsWithCollections{}
	if text == "" {
		filtered = dbsWitColls
		expand = false
	} else {
		re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(text))
		for _, db := range dbsWitColls {
			matchedDB := re.MatchString(db.DB)
			matchedCollections := []string{}

			for _, coll := range db.Collections {
				if re.MatchString(coll) {
					matchedCollections = append(matchedCollections, coll)
				}
			}

			if matchedDB || len(matchedCollections) > 0 {
				filteredDB := mongo.DBsWithCollections{
					DB:          db.DB,
					Collections: matchedCollections,
				}
				if matchedDB {
					filteredDB.Collections = db.Collections
				}
				filtered = append(filtered, filteredDB)
				expand = expand || len(matchedCollections) > 0
			}
		}
	}
	s.DbTree.Render(ctx, filtered, expand)
}

func (s *Database) listDbsAndCollections(ctx context.Context) error {
	dbsWitColls, err := s.Dao.ListDbsWithCollections(ctx, "")
	if err != nil {
		return err
	}
	s.dbsWithColls = dbsWitColls

	return nil
}

func (s *Database) SetSelectFunc(f func(ctx context.Context, db string, coll string) error) {
	s.DbTree.SetSelectFunc(f)
}
