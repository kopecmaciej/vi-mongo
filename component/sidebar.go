package component

import (
	"context"
	"mongo-ui/mongo"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type SideBar struct {
	*tview.Flex

	app         *App
	dao         *mongo.Dao
	Tree        *tview.TreeView
	searchBar   *SearchBar
	nodeSelectF func(a string, b string) error
	flexStack   []flexStack
	mutex       sync.Mutex
	label       string
}

func NewSideBar(dao *mongo.Dao) *SideBar {
	return &SideBar{
		Flex:      tview.NewFlex(),
		Tree:      tview.NewTreeView(),
		searchBar: NewSearchBar("Search"),
		label:     "sideBar",
		dao:       dao,
		mutex:     sync.Mutex{},
	}
}

func (s *SideBar) Init(ctx context.Context) error {
	s.setStyle()
	s.setShortcuts(ctx)

	s.app = GetApp(ctx)

	rootNode := s.dbNode("Databases")
	s.Tree.SetRoot(rootNode)

	s.SetDirection(tview.FlexRow)

	s.flexStack = []flexStack{
		{s.searchBar.label, 1, 0, false, false},
		{s.label, 0, 1, true, true},
	}

	s.render(ctx)

	s.renderTree(ctx, "")

	go s.searchBarListener(ctx)

	return nil
}

func (s *SideBar) setStyle() {
	s.Tree.SetBackgroundColor(tcell.ColorDefault)
	s.Tree.SetBorderPadding(1, 1, 1, 1)
	s.Tree.SetBorder(true)
	s.Tree.SetBorderColor(tcell.ColorDimGray)
	s.Tree.SetTitle(" Databases ")

	s.Flex.SetBackgroundColor(tcell.ColorDefault)
}

func (s *SideBar) setShortcuts(ctx context.Context) {
	s.Flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			if event.Rune() == '/' {
				s.seachBarToggle()
				s.render(ctx)
				go s.searchBar.SetText("")
			}
		}
		return event
	})
}

func (s *SideBar) render(ctx context.Context) error {
	s.Flex.Clear()

	for _, item := range s.flexStack {
		if item.enabled {
			primitive := s.getPrimitiveByLabel(item.label)
			s.AddItem(primitive, item.fixed, item.prop, item.focus)
		}
	}

	return nil
}

func (s *SideBar) searchBarListener(ctx context.Context) {
	keyChan := s.searchBar.KeyChan

	for {
		key := <-keyChan
		switch key {
		case tcell.KeyEsc:
			s.app.QueueUpdateDraw(func() {
				s.Flex.RemoveItem(s.searchBar)
				s.seachBarToggle()
				s.renderTree(context.Background(), "")
			})
		case tcell.KeyEnter:
			s.app.QueueUpdateDraw(func() {
				s.Flex.RemoveItem(s.searchBar)
				s.renderTree(context.Background(), s.searchBar.GetText())
				s.seachBarToggle()
			})
		}
	}
}

func (s *SideBar) getPrimitiveByLabel(label string) tview.Primitive {
	switch label {
	case "searchBar":
		return s.searchBar
	case "sideBar":
		return s.Tree
	default:
		return nil
	}
}

func (s *SideBar) renderTree(ctx context.Context, filter string) error {
	rootNode := s.rootNode()
	s.Tree.SetRoot(rootNode)

	dbsWitColls, err := s.dao.ListDbsWithCollections(ctx, filter)
	if err != nil {
		return err
	}

	if len(dbsWitColls) == 0 {
		emptyNode := tview.NewTreeNode("No databases found")
		emptyNode.SetColor(tcell.ColorRed)
		emptyNode.SetSelectable(false)

		rootNode.AddChild(emptyNode)
		return nil
	}

	for _, item := range dbsWitColls {
		parent := s.dbNode(item.DB)
		rootNode.AddChild(parent)

		for _, child := range item.Collections {
			child := s.collNode(child)
			parent.AddChild(child)

			child.SetSelectedFunc(func() {
				s.nodeSelectF(item.DB, child.GetText())
			})
		}
	}

	s.Tree.SetCurrentNode(rootNode.GetChildren()[0])

	return nil
}

func (s *SideBar) seachBarToggle() {
	s.mutex.Lock()
	if s.flexStack[0].enabled {
		s.flexStack[0].enabled = false
		s.app.SetFocus(s.Tree)
	} else {
		s.flexStack[0].enabled = true
		s.app.SetFocus(s.searchBar)
	}
	s.mutex.Unlock()
}

func (s *SideBar) rootNode() *tview.TreeNode {
	r := tview.NewTreeNode("Databases")
	r.SetColor(tcell.ColorRed)
	r.SetSelectable(false)
	r.SetExpanded(true)

	return r
}

func (s *SideBar) dbNode(name string) *tview.TreeNode {
	r := tview.NewTreeNode(name)
	r.SetColor(tcell.ColorGreen)
	r.SetSelectable(true)
	r.SetExpanded(false)

	r.SetSelectedFunc(func() {
		r.SetExpanded(!r.IsExpanded())
	})

	return r
}

func (s *SideBar) collNode(name string) *tview.TreeNode {
	ch := tview.NewTreeNode(name)
	ch.SetColor(tcell.ColorWhite)
	ch.SetSelectable(true)
	ch.SetExpanded(false)

	return ch
}
