package component

import (
	"context"
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/config"
	"github.com/kopecmaciej/mongui/manager"
	"github.com/kopecmaciej/mongui/mongo"
	"github.com/kopecmaciej/mongui/primitives"
	"github.com/kopecmaciej/tview"
	"github.com/rs/zerolog/log"
)

const (
	DBTreeComponent       manager.Component = "DBTree"
	InputModalComponent   manager.Component = "InputModal"
	ConfirmModalComponent manager.Component = "ConfirmModal"
)

type DBTree struct {
	*Component
	*tview.TreeView

	inputModal *primitives.InputModal
	style      *config.SidebarStyle

	NodeSelectFunc func(ctx context.Context, db string, coll string, filter map[string]interface{}) error
}

func NewDBTree() *DBTree {
	d := &DBTree{
		Component:  NewComponent(DBTreeComponent),
		TreeView:   tview.NewTreeView(),
		inputModal: primitives.NewInputModal(),
	}

	d.SetAfterInitFunc(d.init)

	return d
}

func (t *DBTree) init() error {
	ctx := context.Background()

	t.setStyle()
	t.setKeybindings(ctx)

	return nil
}

func (t *DBTree) setStyle() {
	t.style = &t.app.Styles.Sidebar
	t.SetBorder(true)
	t.SetTitle(" Databases ")
	t.SetBorderPadding(0, 0, 1, 1)
	t.SetGraphics(false)

	t.SetBackgroundColor(t.style.BackgroundColor.Color())
	t.SetBorderColor(t.style.BorderColor.Color())
	t.SetGraphicsColor(t.style.BranchColor.Color())
	t.SetSelectedFunc(func(node *tview.TreeNode) {
		t.SetCurrentNode(node)
	})
}

func (t *DBTree) setKeybindings(ctx context.Context) {
	k := t.app.Keys
	t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case k.Contains(k.Root.Sidebar.DBTree.ExpandAll, event.Name()):
			t.GetRoot().ExpandAll()
			return nil
		case k.Contains(k.Root.Sidebar.DBTree.CollapseAll, event.Name()):
			t.GetRoot().CollapseAll()
			t.GetRoot().SetExpanded(true)
			return nil
		case k.Contains(k.Root.Sidebar.DBTree.AddCollection, event.Name()):
			t.addCollection(ctx)
			return nil
		case k.Contains(k.Root.Sidebar.DBTree.DeleteCollection, event.Name()):
			t.deleteCollection(ctx)
			return nil
		case k.Contains(k.Root.Sidebar.DBTree.ToggleExpand, event.Name()):
			t.GetCurrentNode().SetExpanded(!t.GetCurrentNode().IsExpanded())
			return nil
		}

		return event
	})
}

func (t *DBTree) RenderTree(ctx context.Context, dbsWitColls []mongo.DBsWithCollections, expand bool) {
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

func (t *DBTree) addCollection(ctx context.Context) error {
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

	label := fmt.Sprintf("Add collection name for db %s", db)
	t.inputModal.SetLabel(label)
	t.inputModal.SetInputLabel("Collection name: ")
	t.inputModal.SetFieldBackgroundColor(tcell.ColorBlack)
	t.inputModal.SetBorder(true)

	t.inputModal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			collectionName := t.inputModal.GetText()
			if collectionName == "" {
				return event
			}
			db, collectionName = t.removeSymbols(db, collectionName)
			log.Info().Msgf("Adding collection %s to db %s", collectionName, db)
			err := t.dao.AddCollection(ctx, db, collectionName)
			if err != nil {
				log.Error().Err(err).Msg("Error adding collection")
				return nil
			}
			t.addChildNode(ctx, parent, collectionName, true)
			t.inputModal.SetText("")
			t.app.Root.RemovePage(InputModalComponent)
		}
		if event.Key() == tcell.KeyEscape {
			t.inputModal.SetText("")
			t.app.Root.RemovePage(InputModalComponent)
		}

		return event
	})

	t.app.Root.AddPage(InputModalComponent, t.inputModal, true, true)

	return nil
}

func (t *DBTree) deleteCollection(ctx context.Context) error {
	level := t.GetCurrentNode().GetLevel()
	if level == 0 || level == 1 {
		return fmt.Errorf("Cannot delete database")
	}
	parent := t.GetCurrentNode().GetReference().(*tview.TreeNode)
	db, coll := t.removeSymbols(parent.GetText(), t.GetCurrentNode().GetText())

	confirmModal := tview.NewModal()
	confirmModal.SetButtonTextColor(tcell.ColorWhite)
	text := fmt.Sprintf("Are you sure you want to delete collection %s from db %s", coll, db)
	confirmModal.SetText(text).
		AddButtons([]string{"OK", "Cancel"})
	confirmModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		defer t.app.Root.RemovePage(ConfirmModalComponent)
		if buttonLabel == "OK" {
			err := t.dao.DeleteCollection(ctx, db, coll)
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

	t.app.Root.AddPage(ConfirmModalComponent, confirmModal, true, true)

	return nil
}

func (t *DBTree) addChildNode(ctx context.Context, parent *tview.TreeNode, collectionName string, expand bool) {
	collNode := t.collNode(collectionName)
	parent.AddChild(collNode).SetExpanded(expand)
	collNode.SetReference(parent)
	collNode.SetSelectedFunc(func() {
		db, coll := t.removeSymbols(parent.GetText(), collNode.GetText())
		t.NodeSelectFunc(ctx, db, coll, nil)
	})
}

func (t *DBTree) rootNode() *tview.TreeNode {
	r := tview.NewTreeNode("")
	r.SetColor(t.style.NodeColor.Color())
	r.SetSelectable(false)
	r.SetExpanded(true)

	return r
}

func (t *DBTree) dbNode(name string) *tview.TreeNode {
	r := tview.NewTreeNode(fmt.Sprintf("%s %s", t.style.NodeSymbol.String(), name))
	r.SetColor(t.style.NodeColor.Color())
	r.SetSelectable(true)
	r.SetExpanded(false)

	r.SetSelectedFunc(func() {
		r.SetExpanded(!r.IsExpanded())
	})

	return r
}

func (t *DBTree) collNode(name string) *tview.TreeNode {
	ch := tview.NewTreeNode(fmt.Sprintf("%s %s", t.style.LeafSymbol.String(), name))
	ch.SetColor(t.style.LeafColor.Color())
	ch.SetSelectable(true)
	ch.SetExpanded(false)

	return ch
}

func (t *DBTree) removeSymbols(db, coll string) (string, string) {
	db = strings.Replace(db, t.style.NodeSymbol.String(), "", 1)
	coll = strings.Replace(coll, t.style.LeafSymbol.String(), "", 1)
	return strings.TrimSpace(db), strings.TrimSpace(coll)
}
