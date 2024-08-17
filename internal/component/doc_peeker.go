package component

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/kopecmaciej/mongui/internal/config"
	"github.com/kopecmaciej/mongui/internal/manager"
	"github.com/kopecmaciej/mongui/internal/mongo"
	"github.com/kopecmaciej/mongui/internal/primitives"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/rs/zerolog/log"
)

const (
	DocPeekerComponent manager.Component = "DocPeeker"
)

// peekerState is used to store the state of the document being peeked at
type peekerState struct {
	mongo.CollectionState
	rawDocument string
}

// DocPeeker is a component that provides a modal view for peeking at a document
type DocPeeker struct {
	*Component
	*primitives.ModalView

	style       *config.DocPeekerStyle
	docModifier *DocModifier
	state       peekerState
}

// NewDocPeeker creates a new DocPeeker component
func NewDocPeeker() *DocPeeker {
	peekr := &DocPeeker{
		Component:   NewComponent(DocPeekerComponent),
		ModalView:   primitives.NewModalView(),
		docModifier: NewDocModifier(),
	}

	peekr.SetAfterInitFunc(peekr.init)

	return peekr
}

func (dc *DocPeeker) init() error {
	ctx := context.Background()

	dc.setStyle()
	dc.setKeybindings(ctx)

	if err := dc.docModifier.Init(dc.app); err != nil {
		return err
	}

	return nil
}

func (dc *DocPeeker) setStyle() {
	dc.style = &dc.app.Styles.DocPeeker
	dc.SetBorder(true)
	dc.SetTitle("Document Details")
	dc.SetTitleAlign(tview.AlignLeft)
	dc.SetTitleColor(dc.style.TitleColor.Color())
	dc.SetBackgroundColor(dc.style.BackgroundColor.Color())
	dc.SetBorderColor(dc.style.BorderColor.Color())
	dc.SetHighlightColor(dc.style.HighlightColor.Color())
	dc.SetDocumentColors(
		dc.style.KeyColor.Color(),
		dc.style.ValueColor.Color(),
		dc.style.BracketColor.Color(),
		dc.style.ArrayColor.Color(),
	)

	dc.ModalView.AddButtons([]string{"Edit", "Close"})
}

func (dc *DocPeeker) setKeybindings(ctx context.Context) {
	k := dc.app.Keys
	dc.ModalView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case k.Contains(k.DocPeeker.MoveToTop, event.Name()):
			dc.MoveToTop()
			return nil
		case k.Contains(k.DocPeeker.MoveToBottom, event.Name()):
			dc.MoveToBottom()
			return nil
		case k.Contains(k.DocPeeker.CopyFullLine, event.Name()):
			if err := dc.ModalView.CopySelectedLine(clipboard.WriteAll, "full"); err != nil {
				ShowErrorModal(dc.app.Root, "Error copying full line", err)
			}
			return nil
		case k.Contains(k.DocPeeker.CopyValue, event.Name()):
			if err := dc.ModalView.CopySelectedLine(clipboard.WriteAll, "value"); err != nil {
				ShowErrorModal(dc.app.Root, "Error copying value", err)
			}
			return nil
		case k.Contains(k.DocPeeker.Refresh, event.Name()):
			if err := dc.render(ctx); err != nil {
				log.Error().Err(err).Msg("Error refreshing document")
			}
			return nil
		}
		return event
	})
}

func (dc *DocPeeker) MoveToTop() {
	dc.ModalView.MoveToTop()
}

func (dc *DocPeeker) MoveToBottom() {
	dc.ModalView.MoveToBottom()
}

func (dc *DocPeeker) Peek(ctx context.Context, db, coll string, jsonString string) error {
	dc.state = peekerState{
		CollectionState: mongo.CollectionState{
			Db:   db,
			Coll: coll,
		},
		rawDocument: jsonString,
	}
	var prettyJson bytes.Buffer
	err := json.Indent(&prettyJson, []byte(jsonString), "", "  ")
	if err != nil {
		log.Error().Err(err).Msg("Error indenting JSON")
		return nil
	}
	text := prettyJson.String()

	dc.ModalView.SetText(primitives.Text{
		Content: text,
		Color:   dc.style.ValueColor.Color(),
		Align:   tview.AlignLeft,
	})

	root := dc.app.Root
	root.AddPage(dc.identifier, dc.ModalView, true, true)
	dc.ModalView.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Edit" {
			updatedDoc, err := dc.docModifier.Edit(ctx, db, coll, jsonString)
			if err != nil {
				ShowErrorModal(dc.app.Root, "Error editing document", err)
				return
			}
			dc.state.rawDocument = updatedDoc
			dc.render(ctx)
		} else if buttonLabel == "Close" || buttonLabel == "" {
			root.RemovePage(dc.identifier)
			dc.app.SetFocus(root.content)
		}
	})
	return nil
}

func (dc *DocPeeker) render(ctx context.Context) error {
	return dc.Peek(ctx, dc.state.Db, dc.state.Coll, dc.state.rawDocument)
}
