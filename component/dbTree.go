package component

import (
	"context"
	"mongo-ui/mongo"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type DBTree struct {
	*tview.TreeView

	NodeSelectF func(a string, b string, filter map[string]interface{}) error
}

func NewDBTree() *DBTree {
	return &DBTree{
		TreeView: tview.NewTreeView(),
	}
}

func (t *DBTree) Init() {
	t.setStyle()
	t.setShortcuts()

	rootNode := t.dbNode("Databases")
	t.SetRoot(rootNode)

}

func (t *DBTree) setStyle() {
	t.SetBorder(true)
	t.SetTitle(" Databases ")
	t.SetBorderPadding(0, 0, 1, 1)
	t.SetSelectedFunc(func(node *tview.TreeNode) {
		t.SetCurrentNode(node)
	})
}

func (t *DBTree) setShortcuts() {
	t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			if event.Rune() == 'E' {
				t.GetRoot().ExpandAll()
				return nil
			}
			if event.Rune() == 'W' {
				t.GetRoot().CollapseAll()
				t.GetRoot().SetExpanded(true)
				return nil
			}
			if event.Rune() == 'o' {
				t.GetCurrentNode().SetExpanded(!t.GetCurrentNode().IsExpanded())
				return nil
			}
		}
		return event
	})
}

func (t *DBTree) RenderTree(ctx context.Context, dbsWitColls []mongo.DBsWithCollections, filter string) {
	rootNode := t.rootNode()
	t.SetRoot(rootNode)

	if len(dbsWitColls) == 0 {
		emptyNode := tview.NewTreeNode("No databases found")
		emptyNode.SetSelectable(false)

		rootNode.AddChild(emptyNode)
	}

	for _, item := range dbsWitColls {
		parent := t.dbNode(item.DB)
		rootNode.AddChild(parent)

		for _, child := range item.Collections {
			child := t.collNode(child)
			parent.AddChild(child)

			child.SetSelectedFunc(func() {
				t.NodeSelectF(parent.GetText(), child.GetText(), nil)
			})
		}
	}

	t.SetCurrentNode(rootNode.GetChildren()[0])
}

func (t *DBTree) rootNode() *tview.TreeNode {
	r := tview.NewTreeNode("Databases")
	r.SetColor(tcell.NewRGBColor(56, 125, 68))
	r.SetSelectable(false)
	r.SetExpanded(true)

	collNode := tview.NewTreeNode("Collections")
	collNode.SetColor(tcell.NewRGBColor(22, 54, 148))
	collNode.SetSelectable(false)
	collNode.SetExpanded(true)

	r.AddChild(collNode)

	return r
}

func (t *DBTree) dbNode(name string) *tview.TreeNode {
	r := tview.NewTreeNode(name)
	r.SetColor(tcell.NewRGBColor(56, 125, 68))
	r.SetSelectable(true)
	r.SetExpanded(false)

	r.SetSelectedFunc(func() {
		r.SetExpanded(!r.IsExpanded())
	})

	return r
}

func (t *DBTree) collNode(name string) *tview.TreeNode {
	ch := tview.NewTreeNode(name)
	ch.SetColor(tcell.NewRGBColor(22, 54, 148))
	ch.SetSelectable(true)
	ch.SetExpanded(false)

	return ch
}
