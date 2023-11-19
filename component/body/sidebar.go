package body

import (
	"context"
	"mongo-ui/mongo"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type SideBar struct {
	*tview.TreeView

	label     string
	app       *tview.Application
	dao       *mongo.Dao
	eventChan chan interface{}
}

func NewSideBar(dao *mongo.Dao) *SideBar {
	return &SideBar{
		TreeView: tview.NewTreeView(),

		label:     "sideBar",
		dao:       dao,
		eventChan: make(chan interface{}, 1),
	}
}

func (s *SideBar) Init(ctx context.Context) error {
	s.setStyle()
	s.app = ctx.Value("app").(*tview.Application)

	rootNode := s.dbNode("Databases")
	s.SetRoot(rootNode)

	return nil
}

func (s *SideBar) RenderTree(ctx context.Context, nodeSelectF func(a string, b string) error) error {
	rootNode := s.rootNode()
	s.SetRoot(rootNode)

	dbsWitColls, err := s.dao.ListDbsWithCollections(ctx)
	if err != nil {
		return err
	}

	for _, item := range dbsWitColls {
		parent := s.dbNode(item.DB)
		rootNode.AddChild(parent)

		for _, child := range item.Collections {
			child := s.collNode(child)
			parent.AddChild(child)

			child.SetSelectedFunc(func() {
				nodeSelectF(item.DB, child.GetText())
			})
		}
	}

	s.SetCurrentNode(rootNode.GetChildren()[0])

	return nil
}

func (s *SideBar) setStyle() {
	s.SetBackgroundColor(tcell.ColorDefault)
	s.SetBorderPadding(0, 0, 3, 3)
	s.SetBorder(true)
	s.SetTitle("Databases")
}

func (s *SideBar) rootNode() *tview.TreeNode {
	r := tview.NewTreeNode("Databases")
	r.SetColor(tcell.ColorWhite)
	r.SetSelectable(false)
	r.SetExpanded(true)

	return r
}

func (s *SideBar) dbNode(name string) *tview.TreeNode {
	r := tview.NewTreeNode(name)
	r.SetColor(tcell.ColorRed)
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
