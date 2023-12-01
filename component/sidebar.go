package component

import (
	"context"
	"mongo-ui/mongo"
	"regexp"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type SideBar struct {
	*tview.Flex

	DBTree       *DBTree
	FilterBar    *InputBar
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
		DBTree:    NewDBTree(),
		FilterBar: NewInputBar("Filter"),
		label:     "sideBar",
		dao:       dao,
		mutex:     sync.Mutex{},
	}
}

func (s *SideBar) Init(ctx context.Context) error {
	s.app = GetApp(ctx)

	s.setStyle()
	s.setShortcuts(ctx)

	s.DBTree.Init()

	if err := s.render(ctx); err != nil {
		return err
	}
	if err := s.fetchAndRender(ctx, ""); err != nil {
		return err
	}
	if err := s.FilterBar.Init(ctx); err != nil {
		return err
	}
	go s.filterBarListener(ctx)

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
				s.toogleFilterBar(ctx)
				go s.FilterBar.SetText("")
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

	if s.FilterBar.IsEnabled() {
		s.Flex.AddItem(s.FilterBar, 3, 0, false)
		primitive = s.FilterBar
	}
	defer s.app.SetFocus(primitive)

	s.Flex.AddItem(s.DBTree, 0, 1, true)

	return nil
}

func (s *SideBar) filterBarListener(ctx context.Context) {
	eventChan := s.FilterBar.EventChan

	for {
		key := <-eventChan
		if _, ok := key.(tcell.Key); !ok {
			continue
		}
		switch key {
		case tcell.KeyEsc:
			s.app.QueueUpdateDraw(func() {
				s.toogleFilterBar(ctx)
			})
		case tcell.KeyEnter:
			s.app.QueueUpdateDraw(func() {
				s.filter(ctx)
			})
		}
	}
}

func (s *SideBar) filter(ctx context.Context) {
	dbsWitColls := s.dbsWithColls
	filtered := []mongo.DBsWithCollections{}
	text := s.FilterBar.GetText()
	if text == "" {
		s.toogleFilterBar(ctx)
		return
	}
	for _, item := range dbsWitColls {
		re := regexp.MustCompile(`(?i)` + text)
		if re.MatchString(item.DB) {
			filtered = append(filtered, item)
		}
		for _, child := range item.Collections {
			if re.MatchString(child) {
				filtered = append(filtered, item)
			}
		}
	}
	s.toogleFilterBar(ctx)
	s.DBTree.RenderTree(ctx, filtered, text)
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

func (s *SideBar) toogleFilterBar(ctx context.Context) {
	s.FilterBar.Toggle()
	s.render(ctx)
}
