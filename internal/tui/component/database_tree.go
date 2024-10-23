package component

import (
	"context"
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
	"github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/tui/modal"
	"github.com/kopecmaciej/vi-mongo/internal/tui/primitives"
	"github.com/rs/zerolog/log"
)

const (
	InputModalId          = "InputModal"
	ConfirmModalId        = "ConfirmModal"
	DatabaseTreeId        = "DatabaseTree"
	DatabaseDeleteModalId = "DatabaseDeleteModal"
)

type DatabaseTree struct {
	*core.BaseElement
	*core.TreeView

	inputModal  *primitives.InputModal
	deleteModal *modal.Delete
	style       *config.DatabasesStyle

	nodeSelectFunc func(ctx context.Context, db string, coll string) error
}

func NewDatabaseTree() *DatabaseTree {
	d := &DatabaseTree{
		BaseElement: core.NewBaseElement(),
		TreeView:    core.NewTreeView(),
		inputModal:  primitives.NewInputModal(),
		deleteModal: modal.NewDeleteModal(DatabaseDeleteModalId),
	}

	d.SetIdentifier(DatabaseTreeId)
	d.SetAfterInitFunc(d.init)

	return d
}

func (t *DatabaseTree) init() error {
	ctx := context.Background()

	t.setStyle()
	t.setLayout()
	t.setKeybindings(ctx)
	t.SetSelectedFunc(func(node *tview.TreeNode) {
		t.SetCurrentNode(node)
	})

	if err := t.deleteModal.Init(t.App); err != nil {
		return err
	}

	t.handleEvents()

	return nil
}

func (t *DatabaseTree) setLayout() {
	t.SetBorder(true)
	t.SetTitle(" Databases ")
	t.SetBorderPadding(0, 0, 1, 1)
	t.SetGraphics(false)

	t.inputModal.SetBorder(true)
	t.inputModal.SetTitle("Add collection")
}

func (t *DatabaseTree) setStyle() {
	globalStyle := t.App.GetStyles()
	t.TreeView.SetStyle(globalStyle)
	t.style = &globalStyle.Databases

	t.inputModal.SetBorderColor(globalStyle.Global.BorderColor.Color())
	t.inputModal.SetBackgroundColor(globalStyle.Global.BackgroundColor.Color())
	t.inputModal.SetFieldTextColor(globalStyle.Others.ModalTextColor.Color())
	t.inputModal.SetFieldBackgroundColor(globalStyle.Global.ContrastBackgroundColor.Color())
}

func (t *DatabaseTree) setKeybindings(ctx context.Context) {
	closedNodeSymbol := config.SymbolWithColor(t.style.ClosedNodeSymbol, t.style.NodeSymbolColor)
	openNodeSymbol := config.SymbolWithColor(t.style.OpenNodeSymbol, t.style.NodeSymbolColor)
	k := t.App.GetKeys()
	t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case k.Contains(k.Database.ExpandAll, event.Name()):
			t.expandAllNodes(closedNodeSymbol, openNodeSymbol)
			return nil
		case k.Contains(k.Database.CollapseAll, event.Name()):
			t.collapseAllNodes(openNodeSymbol, closedNodeSymbol)
			return nil
		case k.Contains(k.Database.AddCollection, event.Name()):
			t.showAddCollectionModal(ctx)
			return nil
		case k.Contains(k.Database.DeleteCollection, event.Name()):
			t.showDeleteCollectionModal(ctx)
			return nil
		case k.Contains(k.Database.RenameCollection, event.Name()):
			t.showRenameCollectionModal(ctx)
			return nil
		}
		return event
	})
}

func (t *DatabaseTree) expandAllNodes(closedSymbol, openSymbol string) {
	t.GetRoot().ExpandAll()
	t.GetRoot().Walk(func(node, parent *tview.TreeNode) bool {
		t.setNodeSymbol(node, closedSymbol, openSymbol)
		return true
	})
}

func (t *DatabaseTree) collapseAllNodes(openSymbol, closedSymbol string) {
	t.GetRoot().CollapseAll()
	t.GetRoot().SetExpanded(true)
	t.GetRoot().Walk(func(node, parent *tview.TreeNode) bool {
		t.setNodeSymbol(node, openSymbol, closedSymbol)
		return true
	})
}

func (t *DatabaseTree) handleEvents() {
	go t.HandleEvents(DatabaseTreeId, func(event manager.EventMsg) {
		switch event.Message.Type {
		case manager.StyleChanged:
			t.setStyle()
			t.RefreshStyle()
		}
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

func (t *DatabaseTree) RefreshStyle() {
	t.GetRoot().Walk(func(node, parent *tview.TreeNode) bool {
		if parent != nil {
			t.updateNodeSymbol(parent)
		}
		t.updateLeafSymbol(node)
		return true
	})
}

func (t *DatabaseTree) showAddCollectionModal(ctx context.Context) error {
	parent := t.getParentNode()
	if parent == nil {
		return nil
	}
	db := parent.GetText()

	t.inputModal.SetLabel(fmt.Sprintf("Add collection name for [%s][::b]%s", t.style.NodeTextColor.Color(), db))
	t.inputModal.SetInputCapture(t.createAddCollectionInputCapture(ctx, parent, db))
	t.App.Pages.AddPage(InputModalId, t.inputModal, true, true)
	return nil
}

func (t *DatabaseTree) createAddCollectionInputCapture(ctx context.Context, parent *tview.TreeNode, db string) func(*tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			t.handleAddCollection(ctx, parent, db)
		case tcell.KeyEscape:
			t.closeAddModal()
		}
		return event
	}
}

func (t *DatabaseTree) handleAddCollection(ctx context.Context, parent *tview.TreeNode, db string) {
	collectionName := t.inputModal.GetText()
	if collectionName == "" {
		return
	}
	db, collectionName = t.removeSymbols(db, collectionName)
	err := t.Dao.AddCollection(ctx, db, collectionName)
	if err != nil {
		log.Error().Err(err).Msg("Error adding collection")
		return
	}
	t.addChildNode(ctx, parent, collectionName, true)
	t.closeAddModal()
}

func (t *DatabaseTree) closeAddModal() {
	t.inputModal.SetText("")
	t.App.Pages.RemovePage(InputModalId)
}

func (t *DatabaseTree) showDeleteCollectionModal(ctx context.Context) error {
	if t.GetCurrentNode().GetLevel() < 2 {
		return fmt.Errorf("cannot delete database")
	}
	parent := t.GetCurrentNode().GetReference().(*tview.TreeNode)
	db, coll := parent.GetText(), t.GetCurrentNode().GetText()
	t.deleteModal.SetText(t.getDeleteConfirmationText(db, coll))
	db, coll = t.removeSymbols(db, coll)
	t.deleteModal.SetDoneFunc(t.createDeleteCollectionDoneFunc(ctx, db, coll, parent))
	t.App.Pages.AddPage(ConfirmModalId, t.deleteModal, true, true)
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
	r.SetColor(t.style.NodeTextColor.Color())
	r.SetSelectable(false)
	r.SetExpanded(true)

	return r
}

func (t *DatabaseTree) dbNode(name string) *tview.TreeNode {
	openNodeSymbol := config.SymbolWithColor(t.style.OpenNodeSymbol, t.style.NodeSymbolColor)
	closedNodeSymbol := config.SymbolWithColor(t.style.ClosedNodeSymbol, t.style.NodeSymbolColor)
	r := tview.NewTreeNode(fmt.Sprintf("%s %s", closedNodeSymbol, name))
	r.SetColor(t.style.NodeTextColor.Color())
	r.SetSelectable(true)
	r.SetExpanded(false)

	r.SetSelectedFunc(func() {
		if r.IsExpanded() {
			r.SetText(fmt.Sprintf("%s %s", closedNodeSymbol, name))
		} else {
			r.SetText(fmt.Sprintf("%s %s", openNodeSymbol, name))
		}
		r.SetExpanded(!r.IsExpanded())
	})

	return r
}

func (t *DatabaseTree) collNode(name string) *tview.TreeNode {
	leafSymbol := config.SymbolWithColor(t.style.LeafSymbol, t.style.LeafSymbolColor)
	ch := tview.NewTreeNode(fmt.Sprintf("%s %s", leafSymbol, name))
	ch.SetColor(t.style.LeafTextColor.Color())
	ch.SetSelectable(true)
	ch.SetExpanded(false)

	return ch
}

func (t *DatabaseTree) removeSymbols(db, coll string) (string, string) {
	openNodeSymbol := config.SymbolWithColor(t.style.OpenNodeSymbol, t.style.NodeSymbolColor)
	closedNodeSymbol := config.SymbolWithColor(t.style.ClosedNodeSymbol, t.style.NodeSymbolColor)
	leafSymbol := config.SymbolWithColor(t.style.LeafSymbol, t.style.LeafSymbolColor)
	symbolsToRemove := []string{
		openNodeSymbol,
		closedNodeSymbol,
		leafSymbol,
	}

	for _, symbol := range symbolsToRemove {
		db = strings.ReplaceAll(db, symbol, "")
		coll = strings.ReplaceAll(coll, symbol, "")
	}

	return strings.TrimSpace(db), strings.TrimSpace(coll)
}

func (t *DatabaseTree) setNodeSymbol(node *tview.TreeNode, oldSymbol, newSymbol string) {
	text := node.GetText()
	node.SetText(strings.Replace(text, oldSymbol, newSymbol, 1))
}

func (t *DatabaseTree) getParentNode() *tview.TreeNode {
	level := t.GetCurrentNode().GetLevel()
	if level == 0 {
		return nil
	}
	if level == 1 {
		return t.GetCurrentNode()
	}
	return t.GetCurrentNode().GetReference().(*tview.TreeNode)
}

func (t *DatabaseTree) getDeleteConfirmationText(db, coll string) string {
	return fmt.Sprintf("Are you sure you want to delete [%s]%s[-:-:-] [white]from [%s]%s[-:-:-]",
		t.style.LeafTextColor.Color(), coll, t.style.NodeTextColor.Color(), db)
}

func (t *DatabaseTree) createDeleteCollectionDoneFunc(ctx context.Context, db, coll string, parent *tview.TreeNode) func(int, string) {
	return func(buttonIndex int, buttonLabel string) {
		defer t.App.Pages.RemovePage(ConfirmModalId)
		if buttonIndex == 0 {
			t.handleDeleteCollection(ctx, db, coll, parent)
		}
	}
}

func (t *DatabaseTree) handleDeleteCollection(ctx context.Context, db, coll string, parent *tview.TreeNode) {
	err := t.Dao.DeleteCollection(ctx, db, coll)
	if err != nil {
		return
	}
	t.removeCollectionNode(parent)
}

func (t *DatabaseTree) removeCollectionNode(parent *tview.TreeNode) {
	currentNode := t.GetCurrentNode()
	childCount := parent.GetChildren()
	index := t.findNodeIndex(childCount, currentNode)
	parent.RemoveChild(currentNode)
	t.updateCurrentNode(parent, childCount, index)
}

func (t *DatabaseTree) findNodeIndex(children []*tview.TreeNode, node *tview.TreeNode) int {
	for i, child := range children {
		if child.GetText() == node.GetText() {
			return i
		}
	}
	return -1
}

func (t *DatabaseTree) updateCurrentNode(parent *tview.TreeNode, childCount []*tview.TreeNode, index int) {
	if index == 0 && len(childCount) > 1 {
		t.SetCurrentNode(parent.GetChildren()[0])
	} else if index > 0 {
		t.SetCurrentNode(parent.GetChildren()[index-1])
	}
}

func (t *DatabaseTree) updateNodeSymbol(node *tview.TreeNode) {
	node.SetColor(t.style.NodeTextColor.Color())
	openNodeSymbol := config.SymbolWithColor(t.style.OpenNodeSymbol, t.style.NodeSymbolColor)
	closedNodeSymbol := config.SymbolWithColor(t.style.ClosedNodeSymbol, t.style.NodeSymbolColor)
	currText := strings.Split(node.GetText(), " ")
	if len(currText) < 2 {
		return
	}
	if node.IsExpanded() {
		node.SetText(fmt.Sprintf("%s %s", openNodeSymbol, currText[1]))
	} else {
		node.SetText(fmt.Sprintf("%s %s", closedNodeSymbol, currText[1]))
	}

	node.SetSelectedFunc(func() {
		if node.IsExpanded() {
			node.SetText(fmt.Sprintf("%s %s", closedNodeSymbol, currText[1]))
		} else {
			node.SetText(fmt.Sprintf("%s %s", openNodeSymbol, currText[1]))
		}
		node.SetExpanded(!node.IsExpanded())
	})
}

func (t *DatabaseTree) updateLeafSymbol(node *tview.TreeNode) {
	node.SetColor(t.style.LeafTextColor.Color())
	leafSymbol := config.SymbolWithColor(t.style.LeafSymbol, t.style.LeafSymbolColor)
	currText := strings.Split(node.GetText(), " ")
	if len(currText) < 2 {
		return
	}
	node.SetText(fmt.Sprintf("%s %s", leafSymbol, currText[1]))
}
func (t *DatabaseTree) showRenameCollectionModal(ctx context.Context) error {
	if t.GetCurrentNode().GetLevel() < 2 {
		return fmt.Errorf("cannot rename database")
	}
	db, coll := t.GetCurrentNode().GetReference().(*tview.TreeNode).GetText(), t.GetCurrentNode().GetText()
	t.inputModal.SetLabel(fmt.Sprintf("Rename collection name for [%s][::b]%s", t.style.NodeTextColor.Color(), db))
	t.inputModal.SetInputCapture(t.createRenameCollectionInputCapture(ctx, db, coll))
	t.App.Pages.AddPage(InputModalId, t.inputModal, true, true)
	return nil
}

func (t *DatabaseTree) createRenameCollectionInputCapture(ctx context.Context, db, coll string) func(*tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			t.handleRenameCollection(ctx, db, coll)
		case tcell.KeyEscape:
			t.closeAddModal()
		}
		return event
	}
}

func (t *DatabaseTree) handleRenameCollection(ctx context.Context, db, coll string) {
	newCollectionName := t.inputModal.GetText()
	if newCollectionName == "" {
		return
	}
	db, coll = t.removeSymbols(db, coll)
	err := t.Dao.RenameCollection(ctx, db, coll, newCollectionName)
	if err != nil {
		log.Error().Err(err).Msg("Error renaming collection")
		modal.ShowError(t.App.Pages, "Error renaming collection", err)
		return
	}
	t.renameCollectionNode(newCollectionName)
	t.closeAddModal()
}

func (t *DatabaseTree) renameCollectionNode(newName string) {
	currentNode := t.GetCurrentNode()
	leafSymbol := config.SymbolWithColor(t.style.LeafSymbol, t.style.LeafSymbolColor)
	newText := fmt.Sprintf("%s %s", leafSymbol, newName)
	currentNode.SetText(newText)
}
