package component

import (
	"context"

	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/tui/modal"
	"github.com/kopecmaciej/vi-mongo/internal/tui/primitives"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
)

const (
	DocPeekerView = "DocPeeker"
)

// DocPeeker is a view that provides a modal view for peeking at a document
type DocPeeker struct {
	*core.BaseElement
	*primitives.ViewModal

	style       *config.DocPeekerStyle
	docModifier *DocModifier
	currentDoc  string

	doneFunc func()
}

// NewDocPeeker creates a new DocPeeker view
func NewDocPeeker() *DocPeeker {
	peekr := &DocPeeker{
		BaseElement: core.NewBaseElement(),
		ViewModal:   primitives.NewViewModal(),
		docModifier: NewDocModifier(),
	}

	peekr.SetAfterInitFunc(peekr.init)

	return peekr
}

func (dc *DocPeeker) init() error {
	dc.setStyle()
	dc.setKeybindings()

	if err := dc.docModifier.Init(dc.App); err != nil {
		return err
	}

	return nil
}

func (dc *DocPeeker) setStyle() {
	dc.style = &dc.App.GetStyles().DocPeeker
	dc.SetBorder(true)
	dc.SetTitle("Document Details")
	dc.SetTitleAlign(tview.AlignLeft)
	dc.SetHighlightColor(dc.style.HighlightColor.Color())
	dc.SetDocumentColors(
		dc.style.KeyColor.Color(),
		dc.style.ValueColor.Color(),
		dc.style.BracketColor.Color(),
		dc.style.ArrayColor.Color(),
	)

	dc.ViewModal.AddButtons([]string{"Edit", "Close"})
}

func (dc *DocPeeker) setKeybindings() {
	k := dc.App.GetKeys()
	dc.ViewModal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case k.Contains(k.DocPeeker.MoveToTop, event.Name()):
			dc.MoveToTop()
			return nil
		case k.Contains(k.DocPeeker.MoveToBottom, event.Name()):
			dc.MoveToBottom()
			return nil
		case k.Contains(k.DocPeeker.CopyFullObj, event.Name()):
			if err := dc.ViewModal.CopySelectedLine(clipboard.WriteAll, "full"); err != nil {
				modal.ShowError(dc.App.Pages, "Error copying full line", err)
			}
			return nil
		case k.Contains(k.DocPeeker.CopyValue, event.Name()):
			if err := dc.ViewModal.CopySelectedLine(clipboard.WriteAll, "value"); err != nil {
				modal.ShowError(dc.App.Pages, "Error copying value", err)
			}
			return nil
		case k.Contains(k.DocPeeker.Refresh, event.Name()):
			dc.render()
			return nil
		}
		return event
	})
}

func (dc *DocPeeker) MoveToTop() {
	dc.ViewModal.MoveToTop()
}

func (dc *DocPeeker) MoveToBottom() {
	dc.ViewModal.MoveToBottom()
}

func (dc *DocPeeker) SetDoneFunc(doneFunc func()) {
	dc.doneFunc = doneFunc
}

func (dc *DocPeeker) Peek(ctx context.Context, state *mongo.CollectionState, _id interface{}) error {
	doc, err := state.GetJsonDocById(_id)
	if err != nil {
		return err
	}

	dc.currentDoc = doc
	dc.render()

	root := dc.App.Pages
	root.AddPage(dc.GetIdentifier(), dc.ViewModal, true, true)
	dc.ViewModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Edit" {
			updatedDoc, err := dc.docModifier.Edit(ctx, state.Db, state.Coll, dc.currentDoc)
			if err != nil {
				modal.ShowError(dc.App.Pages, "Error editing document", err)
				return
			}

			state.UpdateRawDoc(updatedDoc)
			dc.currentDoc = updatedDoc
			if dc.doneFunc != nil {
				dc.doneFunc()
			}
			dc.render()
			dc.App.SetFocus(dc.ViewModal)
		} else if buttonLabel == "Close" || buttonLabel == "" {
			root.RemovePage(dc.GetIdentifier())
		}
	})
	return nil
}

func (dc *DocPeeker) render() {
	dc.ViewModal.SetText(primitives.Text{
		Content: dc.currentDoc,
		Color:   dc.style.ValueColor.Color(),
		Align:   tview.AlignLeft,
	})
}
