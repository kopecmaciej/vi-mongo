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
	filterBar   *InputBar
	nodeSelectF func(a string, b string, filter map[string]interface{}) error
	mutex       sync.Mutex
	label       string
}

func NewSideBar(dao *mongo.Dao) *SideBar {
	flex := tview.NewFlex()
	return &SideBar{
		Flex:      flex,
		Tree:      tview.NewTreeView(),
		filterBar: NewInputBar("Filter"),
		label:     "sideBar",
		dao:       dao,
		mutex:     sync.Mutex{},
	}
}

func (s *SideBar) Init(ctx context.Context) error {
	s.app = GetApp(ctx)

	s.setStyle()
	s.setShortcuts(ctx)

	rootNode := s.dbNode("Databases")
	s.Tree.SetRoot(rootNode)

	if err := s.render(ctx); err != nil {
		return err
	}
	if err := s.renderTree(ctx, ""); err != nil {
		return err
	}
	if err := s.filterBar.Init(ctx); err != nil {
		return err
	}
	go s.filterBarListener(ctx)

	return nil
}

func (s *SideBar) setStyle() {
	s.Tree.SetBackgroundColor(tcell.ColorDefault)
	s.Tree.SetBorderPadding(1, 1, 1, 1)
	s.Tree.SetBorder(true)
	s.Tree.SetBorderColor(tcell.ColorDimGray)
	s.Tree.SetTitle(" Databases ")
	s.Tree.SetGraphicsColor(tcell.ColorDefault)

	s.Flex.SetDirection(tview.FlexRow)
	s.Flex.SetBackgroundColor(tcell.ColorDefault)
}

func (s *SideBar) setShortcuts(ctx context.Context) {
	s.Flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			if event.Rune() == '/' {
				s.toogleFilterBar(ctx)
				go s.filterBar.SetText("")
        return nil
			}
		}
		return event
	})
}

func (s *SideBar) render(ctx context.Context) error {
	s.Flex.Clear()

	if s.filterBar.IsEnabled() {
		s.Flex.AddItem(s.filterBar, 3, 0, false)
		defer s.app.SetFocus(s.filterBar)
	} else {
		defer s.app.SetFocus(s.Tree)
	}

	s.Flex.AddItem(s.Tree, 0, 1, true)

	return nil
}

func (s *SideBar) filterBarListener(ctx context.Context) {
	eventChan := s.filterBar.EventChan

	for {
		key := <-eventChan
		if _, ok := key.(tcell.Key); !ok {
			continue
		}
		switch key {
		case tcell.KeyEsc:
			s.app.QueueUpdateDraw(func() {
				s.renderTree(context.Background(), "")
				s.toogleFilterBar(ctx)
			})
		case tcell.KeyEnter:
			s.app.QueueUpdateDraw(func() {
				s.renderTree(context.Background(), s.filterBar.GetText())
				s.toogleFilterBar(ctx)
			})
		}
	}
}

func (s *SideBar) getPrimitiveByLabel(label string) tview.Primitive {
	switch label {
	case "filter":
		return s.filterBar
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
		emptyNode.SetColor(tcell.ColorDefault)
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
				s.nodeSelectF(item.DB, child.GetText(), nil)
			})
		}
	}

	s.Tree.SetCurrentNode(rootNode.GetChildren()[0])

	return nil
}

func (s *SideBar) toogleFilterBar(ctx context.Context) {
	s.filterBar.Toggle()
	s.render(ctx)
}

func (s *SideBar) rootNode() *tview.TreeNode {
	r := tview.NewTreeNode("Databases")
	r.SetColor(tcell.ColorRed)
	r.SetSelectable(false)
	r.SetExpanded(true)

	collNode := tview.NewTreeNode("Collections")
	r.AddChild(collNode)
	collNode.SetColor(tcell.ColorGreen)
	collNode.SetSelectable(false)
	collNode.SetExpanded(true)

	return r
}

func (s *SideBar) dbNode(name string) *tview.TreeNode {
	r := tview.NewTreeNode(name)
	r.SetColor(tcell.ColorIndianRed.TrueColor())
	r.SetSelectable(true)
	r.SetExpanded(false)

	r.SetSelectedFunc(func() {
		r.SetExpanded(!r.IsExpanded())
	})

	return r
}

func (s *SideBar) collNode(name string) *tview.TreeNode {
	ch := tview.NewTreeNode(name)
	ch.SetColor(tcell.ColorGreen.TrueColor())
	ch.SetSelectable(true)
	ch.SetExpanded(false)

	return ch
}
