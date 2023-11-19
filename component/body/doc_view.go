package body

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mongo-ui/mongo"
	"mongo-ui/primitives"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DocView struct {
	app *tview.Application

	parent tview.Primitive
	dao    *mongo.Dao
}

func (d *DocView) DocView(ctx context.Context, db, coll string, rawDocument string) error {
	var prettyJson bytes.Buffer
	err := json.Indent(&prettyJson, []byte(rawDocument), "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return nil
	}
	text := string(prettyJson.Bytes())

	modal := primitives.NewModalView()
	modal.SetBackgroundColor(tcell.ColorDefault)
	modal.SetBorder(true)
	modal.SetTitle("Document Details")
	modal.SetTitleAlign(tview.AlignLeft)
	modal.SetTitleColor(tcell.ColorSteelBlue)
	modal.SetBorderColor(tcell.ColorSteelBlue)

	modal.SetText(primitives.Text{
		Content: text,
		Color:   tcell.ColorWhite,
		Align:   tview.AlignLeft,
	})

	modal.AddButtons([]string{"Edit", "Close"})

	root := ctx.Value("root").(*tview.Pages)
	root.AddPage("details", modal, true, true)
	d.app.SetFocus(modal)
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Edit" {
			d.editDocument(ctx, db, coll, rawDocument)
		} else {
			root.RemovePage("details")
			d.app.SetFocus(d.parent)
		}
	})
	return nil
}

func (d *DocView) editDocument(ctx context.Context, db, coll string, rawDocument string) error {
	var prettyJson bytes.Buffer
	err := json.Indent(&prettyJson, []byte(rawDocument), "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return nil
	}
	text := string(prettyJson.Bytes())

	id := string(prettyJson.Bytes()[8:40])

	editor := tview.NewTextArea()
	editor.SetBackgroundColor(tcell.ColorDefault)
	editor.SetBorder(true)
	editor.SetTitle(id)
	editor.SetTitleAlign(tview.AlignLeft)
	editor.SetTitleColor(tcell.ColorSteelBlue)

	editor.SetText(text, false)

	root := ctx.Value("root").(*tview.Pages)
	root.AddPage("editor", editor, true, true)
	d.app.SetFocus(editor)
	editor.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlS:
			err := d.saveDocument(ctx, db, coll, editor.GetText())
			if err != nil {
				log.Printf("Error saving document: %v", err)
			}
			root.RemovePage("editor")
			d.app.SetFocus(d.parent)
		case tcell.KeyCtrlC, tcell.KeyEsc:
			root.RemovePage("editor")
			d.app.SetFocus(d.parent)
		}
		return event
	})

	return nil
}

func (d *DocView) saveDocument(ctx context.Context, db, coll string, rawDocument string) error {
	var document map[string]interface{}
	err := json.Unmarshal([]byte(rawDocument), &document)
	if err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return nil
	}
	id := document["_id"].(string)

	if id == "" {
		return fmt.Errorf("Document must have an _id")
	}
	mongoId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("Invalid _id: %v", err)
	}

	err = d.dao.UpdateDocument(ctx, db, coll, mongoId, document)
	if err != nil {
		log.Printf("Error updating document: %v", err)
		return nil
	}

	return nil
}
