package component

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/kopecmaciej/mongui/manager"
	"github.com/kopecmaciej/mongui/mongo"
	"github.com/kopecmaciej/mongui/primitives"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
)

const (
	TextPeekerComponent manager.Component = "TextPeeker"
)

type peekerState struct {
	mongo.CollectionState
	rawDocument string
	id          primitive.ObjectID
}

type DocPeeker struct {
	*primitives.ModalView

	eventChan   chan interface{}
	docModifier *DocModifier
	app         *App
	dao         *mongo.Dao
	state       peekerState
	manager     *manager.ComponentManager
}

func NewDocPeeker(dao *mongo.Dao) *DocPeeker {
	return &DocPeeker{
		ModalView:   primitives.NewModalView(),
		docModifier: NewDocModifier(dao),
		dao:         dao,
	}
}

func (jp *DocPeeker) Init(ctx context.Context) error {
	app, err := GetApp(ctx)
	if err != nil {
		return err
	}
	jp.app = app

	jp.setStyle()
	jp.setShortcuts(ctx)

	jp.manager = jp.app.ComponentManager

	if err := jp.docModifier.Init(ctx); err != nil {
		return err
	}
	jp.docModifier.Render = func() error {
		return jp.render(ctx)
	}

	return nil
}

func (jp *DocPeeker) setStyle() {
	jp.SetBorder(true)
	jp.SetTitle("Document Details")
	jp.SetTitleAlign(tview.AlignLeft)
	jp.SetTitleColor(tcell.ColorSteelBlue)

	jp.ModalView.AddButtons([]string{"Edit", "Close"})
}

func (jp *DocPeeker) setShortcuts(ctx context.Context) {
	jp.ModalView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlR {
			if err := jp.render(ctx); err != nil {
				log.Error().Err(err).Msg("Error refreshing document")
			}
			return nil
		}
		return event
	})
}

func (jp *DocPeeker) Peek(ctx context.Context, db, coll string, jsonString string) error {
	jp.state = peekerState{
		CollectionState: mongo.CollectionState{
			Db:   db,
			Coll: coll,
		},
		rawDocument: jsonString,
	}
	var prettyJson bytes.Buffer
	err := json.Indent(&prettyJson, []byte(jsonString), "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return nil
	}
	text := string(prettyJson.Bytes())

	jp.ModalView.SetText(primitives.Text{
		Content: text,
		Color:   tcell.ColorWhite,
		Align:   tview.AlignLeft,
	})

	root := jp.app.Root
	root.AddPage(TextPeekerComponent, jp.ModalView, true, true)
	jp.ModalView.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Edit" {
			updatedDoc, err := jp.docModifier.Edit(ctx, db, coll, jsonString)
			if err != nil {
				log.Error().Err(err).Msg("Error editing document")
			}
			jp.state.rawDocument = updatedDoc
			jp.render(ctx)
		} else if buttonLabel == "Close" || buttonLabel == "" {
			root.RemovePage(TextPeekerComponent)
		}
	})
	return nil
}

func (jp *DocPeeker) render(ctx context.Context) error {
	return jp.Peek(ctx, jp.state.Db, jp.state.Coll, jp.state.rawDocument)
}
