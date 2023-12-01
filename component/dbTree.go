package component

import (
	"context"
	"fmt"
	"mongo-ui/mongo"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type DBTree struct {
	*tview.TreeView

	NodeSelectF func(a string, b string, filter map[string]interface{}) error
	app         *App
	mongo       *mongo.Dao
}

func NewDBTree(mongo *mongo.Dao) *DBTree {
	return &DBTree{
		TreeView: tview.NewTreeView(),
		mongo:    mongo,
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
				t.NodeSelectF(parent.GetText(), child.GetText(), nil)
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

	inputModal := tview.NewModal()
	inputModal.SetButtonTextColor(tcell.ColorWhite)
	text := fmt.Sprintf("Enter collection name for db %s", db)
	collectionName := ""
	inputModal.SetText(text).
		AddButtons([]string{"OK", "Cancel"})

	inputModal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			collectionName += string(event.Rune())
			inputModal.SetText(fmt.Sprintf("%s\n%s", text, collectionName))
		}
		if event.Key() == tcell.KeyBackspace2 {
			if len(collectionName) > 0 {
				collectionName = collectionName[:len(collectionName)-1]
				inputModal.SetText(fmt.Sprintf("%s\n%s", text, collectionName))
			}
		}

		return event
	})

	inputModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		defer t.app.SetFocus(t)
		if buttonLabel == "OK" {
			if collectionName != "" {
				err := t.mongo.AddCollection(ctx, db, collectionName)
				if err != nil {
					return
				}
				collNode := t.collNode(collectionName)
				parent.AddChild(collNode)
				collNode.SetReference(parent)
				defer t.app.Root.RemovePage("inputModal")
			}
		} else if buttonLabel == "Cancel" || buttonLabel == "" {
			t.app.Root.RemovePage("inputModal")
		}
		return
	})

	t.app.Root.AddPage("inputModal", inputModal, true, true)
	t.app.SetFocus(inputModal)

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
			t.app.Root.RemovePage("confirmModal")
			t.app.SetFocus(t)
		} else if buttonLabel == "Cancel" || buttonLabel == "" {
			t.app.Root.RemovePage("confirmModal")
			t.app.SetFocus(t)
		}
	})

	t.app.Root.AddPage("confirmModal", confirmModal, true, true)
	t.app.SetFocus(confirmModal)

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
