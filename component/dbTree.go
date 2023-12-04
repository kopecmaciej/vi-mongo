package component

import (
	"context"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/manager"
	"github.com/kopecmaciej/mongui/mongo"
	"github.com/kopecmaciej/mongui/primitives"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
)

const (
	InputModalComponent   manager.Component = "InputModal"
	ConfirmModalComponent manager.Component = "ConfirmModal"
)

type DBTree struct {
	*tview.TreeView

	inputModal  *primitives.ModalInput
	NodeSelectF func(ctx context.Context, a string, b string, filter map[string]interface{}) error
	app         *App
	mongo       *mongo.Dao
}

func NewDBTree(mongo *mongo.Dao) *DBTree {
	return &DBTree{
		TreeView:   tview.NewTreeView(),
		inputModal: primitives.NewModalInput(),
		mongo:      mongo,
	}
}

func (t *DBTree) Init(ctx context.Context) {
	t.app = GetApp(ctx)
	t.setStyle()
	t.setShortcuts(ctx)

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

func (t *DBTree) setShortcuts(ctx context.Context) {
	t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlD:
			t.deleteCollection(ctx)
			return nil
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
			if event.Rune() == 'A' {
				t.addCollection(ctx)
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
			child.SetReference(parent)
			parent.AddChild(child)

			child.SetSelectedFunc(func() {
				t.NodeSelectF(ctx, parent.GetText(), child.GetText(), nil)
			})
		}
	}

	t.SetCurrentNode(rootNode.GetChildren()[0])
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
			err := t.mongo.AddCollection(ctx, db, collectionName)
			if err != nil {
				log.Error().Err(err).Msg("Error adding collection")
				return nil
			}
			collNode := t.collNode(collectionName)
			parent.AddChild(collNode)
			collNode.SetReference(parent)
			t.app.Root.RemovePage(InputModalComponent)
		}
		if event.Key() == tcell.KeyEscape {
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
	db := parent.GetText()
	collection := t.GetCurrentNode().GetText()

	confirmModal := tview.NewModal()
	confirmModal.SetButtonTextColor(tcell.ColorWhite)
	text := fmt.Sprintf("Are you sure you want to delete collection %s from db %s", collection, db)
	confirmModal.SetText(text).
		AddButtons([]string{"OK", "Cancel"})
	confirmModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "OK" {
			err := t.mongo.DeleteCollection(ctx, db, collection)
			if err != nil {
				return
			}
			parent.RemoveChild(t.GetCurrentNode())
			t.app.Root.RemovePage(ConfirmModalComponent)
			t.app.SetFocus(t)
		} else if buttonLabel == "Cancel" || buttonLabel == "" {
			t.app.Root.RemovePage(ConfirmModalComponent)
			t.app.SetFocus(t)
		}
	})

	t.app.Root.AddPage(ConfirmModalComponent, confirmModal, true, true)

	return nil
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
