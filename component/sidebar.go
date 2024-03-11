package component

import (
	"context"
	"regexp"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/mongo"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
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
		Component: NewComponent("SideBar"),
		Flex:      tview.NewFlex(),
		dbTree:    NewDBTree(),
		filterBar: NewInputBar("Filter"),
		mutex:     sync.Mutex{},
	}

	s.SetAfterInitFunc(s.init)

	return s
}

func (s *SideBar) init() error {
	ctx := context.Background()
	s.setStyle()
	s.setKeybindings(ctx)

	if err := s.dbTree.Init(s.app); err != nil {
		return err
	}

	if err := s.render(ctx); err != nil {
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

func (s *SideBar) setKeybindings(ctx context.Context) {
	manager := s.app.Manager.SetKeyHandlerForComponent(s.GetIdentifier())
	manager(tcell.KeyRune, '/', "Enable filter bar", func(event *tcell.EventKey) *tcell.EventKey {
		s.filterBar.Enable()
		s.render(ctx)
		return nil
	})
	s.Flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return s.app.Manager.HandleKeyEvent(event, s.GetIdentifier())
	})
}

func (s *SideBar) render(ctx context.Context) error {
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
		s.render(ctx)
	}
	s.filterBar.DoneFuncHandler(accceptFunc, rejectFunc)
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
	s.dbTree.RenderTree(ctx, filtered, text)

	log.Debug().Msgf("Filtered: %v", filtered)
}

func (s *SideBar) renderWithFilter(ctx context.Context, filter string) error {
	if err := s.fetchDbsWithCollections(ctx, filter); err != nil {
		return err
	}
	s.dbTree.RenderTree(ctx, s.dbsWithColls, filter)

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
