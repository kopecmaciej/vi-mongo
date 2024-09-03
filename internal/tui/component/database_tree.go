package component

import (
	"context"
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/tui/modal"
	"github.com/kopecmaciej/vi-mongo/internal/tui/primitives"
	"github.com/rs/zerolog/log"
)

const (
	DatabaseTreeView = "DatabaseTree"
	InputModalView   = "InputModal"
	ConfirmModalView = "ConfirmModal"
)

type DatabaseTree struct {
	*core.BaseElement
	*tview.TreeView

	addModal    *primitives.InputModal
	deleteModal *tview.Modal
	style       *config.DatabasesStyle

	nodeSelectFunc func(ctx context.Context, db string, coll string) error
}

func NewDatabaseTree() *DatabaseTree {
	d := &DatabaseTree{
		BaseElement: core.NewBaseElement(),
		TreeView:    tview.NewTreeView(),
		addModal:    primitives.NewInputModal(),
		deleteModal: tview.NewModal(),
	}

	d.SetIdentifier(DatabaseTreeView)
	d.SetIdentifierFunc(d.GetIdentifier)
	d.SetAfterInitFunc(d.init)

	return d
}

func (t *DatabaseTree) init() error {
	ctx := context.Background()

	t.setStyle()
	t.setKeybindings(ctx)

	return nil
}

func (t *DatabaseTree) setStyle() {
	t.style = &t.App.GetStyles().Databases
	t.SetBorder(true)
	t.SetTitle(" Databases ")
	t.SetBorderPadding(0, 0, 1, 1)
	t.SetGraphics(false)

	t.SetGraphicsColor(t.style.BranchColor.Color())
	t.SetSelectedFunc(func(node *tview.TreeNode) {
		t.SetCurrentNode(node)
	})

	t.addModal.SetInputLabel("Collection name: ")
	t.addModal.SetFieldBackgroundColor(tcell.ColorBlack)
	t.addModal.SetBorder(true)

	t.deleteModal.SetButtonTextColor(tcell.ColorWhite)
	t.deleteModal.AddButtons([]string{"[red]Delete", "Cancel"})
}

func (t *DatabaseTree) setKeybindings(ctx context.Context) {
	k := t.App.GetKeys()
	t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case k.Contains(k.DatabaseTree.ExpandAll, event.Name()):
			t.GetRoot().ExpandAll()
			t.GetRoot().Walk(func(node, parent *tview.TreeNode) bool {
				t.setNewSymbol(node, t.style.ClosedNodeSymbol.String(), t.style.OpenNodeSymbol.String())
				return true
			})
			return nil
		case k.Contains(k.DatabaseTree.CollapseAll, event.Name()):
			t.GetRoot().CollapseAll()
			t.GetRoot().SetExpanded(true)
			t.GetRoot().Walk(func(node, parent *tview.TreeNode) bool {
				t.setNewSymbol(node, t.style.OpenNodeSymbol.String(), t.style.ClosedNodeSymbol.String())
				return true
			})
			return nil
		case k.Contains(k.DatabaseTree.AddCollection, event.Name()):
			t.addCollection(ctx)
			return nil
		case k.Contains(k.DatabaseTree.DeleteCollection, event.Name()):
			t.deleteCollection(ctx)
			return nil
		}

		return event
	})
}

func (t *DatabaseTree) Render(ctx context.Context, dbsWitColls []mongo.DBsWithCollections, expand bool) {
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
			t.addChildNode(ctx, parent, child, false)
		}
	}

	t.SetCurrentNode(rootNode.GetChildren()[0])
	if expand {
		t.GetRoot().ExpandAll()
	}
}

func (t *DatabaseTree) addCollection(ctx context.Context) error {
	// get the current node's or parent node's
	level := t.GetCurrentNode().GetLevel()
	if level == 0 {
		return nil
	}
	var parent *tview.TreeNode
	if level == 1 {
		parent = t.GetCurrentNode()
	} else {
		parent = t.GetCurrentNode().GetReference().(*tview.TreeNode)
	}
	db := parent.GetText()

	t.addModal.SetLabel(fmt.Sprintf("Add collection name for [%s][::b]%s", t.style.NodeColor.Color(), db))

	t.addModal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			collectionName := t.addModal.GetText()
			if collectionName == "" {
				return event
			}
			db, collectionName = t.removeSymbols(db, collectionName)
			err := t.Dao.AddCollection(ctx, db, collectionName)
			if err != nil {
				log.Error().Err(err).Msg("Error adding collection")
				return nil
			}
			t.addChildNode(ctx, parent, collectionName, true)
			t.addModal.SetText("")
			t.App.Pages.RemovePage(InputModalView)
		}
		if event.Key() == tcell.KeyEscape {
			t.addModal.SetText("")
			t.App.Pages.RemovePage(InputModalView)
		}

		return event
	})

	t.App.Pages.AddPage(InputModalView, t.addModal, true, true)

	return nil
}

func (t *DatabaseTree) deleteCollection(ctx context.Context) error {
	level := t.GetCurrentNode().GetLevel()
	if level == 0 || level == 1 {
		return fmt.Errorf("cannot delete database")
	}
	parent := t.GetCurrentNode().GetReference().(*tview.TreeNode)
	db, coll := parent.GetText(), t.GetCurrentNode().GetText()
	text := fmt.Sprintf("Are you sure you want to delete [%s]%s [white]from [%s]%s", t.style.LeafColor.Color(), coll, t.style.NodeColor.Color(), db)
	t.deleteModal.SetText(text)
	db, coll = t.removeSymbols(db, coll)
	t.deleteModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		defer t.App.Pages.RemovePage(ConfirmModalView)
		if buttonIndex == 0 {
			err := t.Dao.DeleteCollection(ctx, db, coll)
			if err != nil {
				return
			}
			childCount := parent.GetChildren()
			var index int
			for i, child := range childCount {
				if child.GetText() == t.GetCurrentNode().GetText() {
					index = i
					break
				}
			}
			parent.RemoveChild(t.GetCurrentNode())
			if index == 0 && len(childCount) > 1 {
				t.SetCurrentNode(parent.GetChildren()[0])
			} else if index > 0 {
				t.SetCurrentNode(parent.GetChildren()[index-1])
			}
		}
	})

	t.App.Pages.AddPage(ConfirmModalView, t.deleteModal, true, true)

	return nil
}

func (t *DatabaseTree) SetSelectFunc(f func(ctx context.Context, db string, coll string) error) {
	t.nodeSelectFunc = f
}

func (t *DatabaseTree) addChildNode(ctx context.Context, parent *tview.TreeNode, collectionName string, expand bool) {
	collNode := t.collNode(collectionName)
	parent.AddChild(collNode).SetExpanded(expand)
	collNode.SetReference(parent)
	collNode.SetSelectedFunc(func() {
		db, coll := t.removeSymbols(parent.GetText(), collNode.GetText())
		err := t.nodeSelectFunc(ctx, db, coll)
		if err != nil {
			modal.ShowError(t.App.Pages, "Error selecting node", err)
		}
	})
}

func (t *DatabaseTree) rootNode() *tview.TreeNode {
	r := tview.NewTreeNode("")
	r.SetColor(t.style.NodeColor.Color())
	r.SetSelectable(false)
	r.SetExpanded(true)

	return r
}

func (t *DatabaseTree) dbNode(name string) *tview.TreeNode {
	r := tview.NewTreeNode(fmt.Sprintf("%s %s", t.style.ClosedNodeSymbol.String(), name))
	r.SetColor(t.style.NodeColor.Color())
	r.SetSelectable(true)
	r.SetExpanded(false)

	r.SetSelectedFunc(func() {
		if r.IsExpanded() {
			r.SetText(fmt.Sprintf("%s %s", t.style.ClosedNodeSymbol.String(), name))
		} else {
			r.SetText(fmt.Sprintf("%s %s", t.style.OpenNodeSymbol.String(), name))
		}
		r.SetExpanded(!r.IsExpanded())
	})

	return r
}

func (t *DatabaseTree) collNode(name string) *tview.TreeNode {
	ch := tview.NewTreeNode(fmt.Sprintf("%s %s", t.style.LeafSymbol.String(), name))
	ch.SetColor(t.style.LeafColor.Color())
	ch.SetSelectable(true)
	ch.SetExpanded(false)

	return ch
}

func (t *DatabaseTree) removeSymbols(db, coll string) (string, string) {
	db = strings.Replace(db, t.style.OpenNodeSymbol.String(), "", 1)
	coll = strings.Replace(coll, t.style.LeafSymbol.String(), "", 1)
	return strings.TrimSpace(db), strings.TrimSpace(coll)
}

func (t *DatabaseTree) setNewSymbol(node *tview.TreeNode, oldSymbol, newSymbol string) {
	text := node.GetText()

	node.SetText(strings.Replace(text, oldSymbol, newSymbol, 1))
}
