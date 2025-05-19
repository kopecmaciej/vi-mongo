package component

import (
	"context"

	"github.com/kopecmaciej/vi-mongo/internal/manager"
	"github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/tui/modal"
	"github.com/kopecmaciej/vi-mongo/internal/tui/primitives"
	"github.com/rs/zerolog/log"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
)

const (
	PeekerId = "Peeker"
)

// Peeker is a view that provides a modal view for peeking at a document
type Peeker struct {
	*core.BaseElement
	*core.ViewModal

	docModifier *DocModifier
	currentDoc  string

	doneFunc func()
}

// NewPeeker creates a new Peeker view
func NewPeeker() *Peeker {
	p := &Peeker{
		BaseElement: core.NewBaseElement(),
		ViewModal:   core.NewViewModal(),
		docModifier: NewDocModifier(),
	}

	p.SetIdentifier(PeekerId)
	p.SetAfterInitFunc(p.init)

	return p
}

func (p *Peeker) init() error {
	p.setStyle()
	p.setLayout()
	p.setKeybindings()

	if err := p.docModifier.Init(p.App); err != nil {
		return err
	}

	p.handleEvents()

	return nil
}

func (p *Peeker) handleEvents() {
	go p.HandleEvents(PeekerId, func(event manager.EventMsg) {
		switch event.Message.Type {
		case manager.StyleChanged:
			p.setStyle()
		}
	})
}

func (p *Peeker) setLayout() {
	p.SetBorder(true)
	p.SetTitle("Document Details")
	p.SetTitleAlign(tview.AlignLeft)

	p.ViewModal.AddButtons([]string{"Edit", "Close"})
}

func (p *Peeker) setStyle() {
	style := &p.App.GetStyles().DocPeeker
	p.ViewModal.SetStyle(p.App.GetStyles())
	p.SetHighlightColor(style.HighlightColor.Color())
	p.SetDocumentColors(
		style.KeyColor.Color(),
		style.ValueColor.Color(),
		style.BracketColor.Color(),
	)
}

func (p *Peeker) setKeybindings() {
	k := p.App.GetKeys()
	p.ViewModal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case k.Contains(k.Peeker.MoveToTop, event.Name()):
			p.MoveToTop()
			return nil
		case k.Contains(k.Peeker.MoveToBottom, event.Name()):
			p.MoveToBottom()
			return nil
		case k.Contains(k.Peeker.CopyHighlight, event.Name()):
			if err := p.ViewModal.CopySelectedLine(clipboard.WriteAll, "full"); err != nil {
				log.Error().Err(err).Msg("Error copying full line")
				modal.ShowError(p.App.Pages, "Error copying full line", err)
			}
			return nil
		case k.Contains(k.Peeker.CopyValue, event.Name()):
			if err := p.ViewModal.CopySelectedLine(clipboard.WriteAll, "value"); err != nil {
				log.Error().Err(err).Msg("Error copying value")
				modal.ShowError(p.App.Pages, "Error copying value", err)
			}
			return nil
		case k.Contains(k.Peeker.ToggleFullScreen, event.Name()):
			p.ViewModal.SetFullScreen(!p.ViewModal.IsFullScreen())
			p.ViewModal.MoveToTop()
			return nil
		case k.Contains(k.Peeker.Exit, event.Name()):
			p.App.Pages.RemovePage(p.GetIdentifier())
			return nil
		}
		return event
	})
}

func (p *Peeker) MoveToTop() {
	p.ViewModal.MoveToTop()
}

func (p *Peeker) MoveToBottom() {
	p.ViewModal.MoveToBottom()
}

func (p *Peeker) SetDoneFunc(doneFunc func()) {
	p.doneFunc = doneFunc
}

func (p *Peeker) Render(ctx context.Context, state *mongo.CollectionState, _id any) error {
	p.MoveToTop()
	doc, err := state.GetJsonDocById(_id)
	if err != nil {
		return err
	}

	p.currentDoc = doc
	p.setText()

	p.App.Pages.AddPage(p.GetIdentifier(), p.ViewModal, true, true)
	p.ViewModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Edit" {
			updatedDoc, err := p.docModifier.Edit(ctx, state.Db, state.Coll, _id, p.currentDoc)
			if err != nil {
				modal.ShowError(p.App.Pages, "Error editing document", err)
				return
			}

			if updatedDoc != "" {
				state.UpdateRawDoc(updatedDoc)
				p.currentDoc = updatedDoc
				if p.doneFunc != nil {
					p.doneFunc()
				}
				p.setText()
			}
		} else if buttonLabel == "Close" || buttonLabel == "" {
			p.App.Pages.RemovePage(p.GetIdentifier())
		}
	})
	return nil
}

func (p *Peeker) setText() {
	p.ViewModal.SetText(primitives.Text{
		Content: p.currentDoc,
		Color:   p.App.GetStyles().DocPeeker.ValueColor.Color(),
		Align:   tview.AlignLeft,
	})
}
