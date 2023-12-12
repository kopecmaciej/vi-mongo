package component

import (
	"context"
	"regexp"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/manager"
	"github.com/kopecmaciej/mongui/mongo"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
)

const (
	SideBarComponent manager.Component = "SideBar"
)

type SideBar struct {
	*tview.Flex

	DBTree       *DBTree
	filterBar    *InputBar
	app          *App
	dao          *mongo.Dao
	mutex        sync.Mutex
	label        string
	dbsWithColls []mongo.DBsWithCollections
}

func NewSideBar(dao *mongo.Dao) *SideBar {
	flex := tview.NewFlex()
	return &SideBar{
		Flex:      flex,
		DBTree:    NewDBTree(dao),
		filterBar: NewInputBar("Filter"),
		label:     "sideBar",
		dao:       dao,
		mutex:     sync.Mutex{},
	}
}

func (s *SideBar) Init(ctx context.Context) error {
	app, err := GetApp(ctx)
	if err != nil {
		return err
	}
	s.app = app

	s.setStyle()
	s.setShortcuts(ctx)

	if err := s.DBTree.Init(ctx); err != nil {
		return err
	}

	if err := s.render(ctx); err != nil {
		return err
	}
	if err := s.fetchAndRender(ctx, ""); err != nil {
		return err
	}
	if err := s.filterBar.Init(ctx); err != nil {
		return err
	}
	s.filterBarListener(ctx)

	return nil
}

func (s *SideBar) setStyle() {
	s.Flex.SetDirection(tview.FlexRow)
	s.Flex.SetBackgroundColor(tcell.ColorDefault)
}

func (s *SideBar) setShortcuts(ctx context.Context) {
	s.Flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			if event.Rune() == '/' {
				s.filterBar.Toggle()
				s.render(ctx)
				return nil
			}
		}
		return event
	})
}

func (s *SideBar) render(ctx context.Context) error {
	s.Flex.Clear()

	var primitive tview.Primitive
	primitive = s.DBTree

	if s.filterBar.IsEnabled() {
		s.Flex.AddItem(s.filterBar, 3, 0, false)
		primitive = s.filterBar
	}
	defer s.app.SetFocus(primitive)

	s.Flex.AddItem(s.DBTree, 0, 1, true)

	return nil
}

func (s *SideBar) filterBarListener(ctx context.Context) {
	accceptFunc := func(text string) {
		s.filter(ctx, text)
	}
	rejectFunc := func() {
		s.render(ctx)
	}
	go s.filterBar.EventListener(accceptFunc, rejectFunc)
}

func (s *SideBar) filter(ctx context.Context, text string) {
	defer s.render(ctx)
	dbsWitColls := s.dbsWithColls
	filtered := []mongo.DBsWithCollections{}
	if text == "" {
		return
	}
	for _, db := range dbsWitColls {
		re := regexp.MustCompile(`(?i)` + text)
		if re.MatchString(db.DB) {
			filtered = append(filtered, db)
		}
		//TODO: tree should expand on found coll
		for _, coll := range db.Collections {
			if re.MatchString(coll) {
				filtered = append(filtered, db)
			}
		}
	}
	s.DBTree.RenderTree(ctx, filtered, text)

	log.Debug().Msgf("Filtered: %v", filtered)
}

func (s *SideBar) fetchAndRender(ctx context.Context, filter string) error {
	if err := s.fetchDbsWithCollections(ctx, filter); err != nil {
		return err
	}
	s.DBTree.RenderTree(ctx, s.dbsWithColls, filter)

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
