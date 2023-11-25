package component

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mongo-ui/mongo"
	"mongo-ui/primitives"
	"os"
	"os/exec"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type docState struct {
	contentState
	rawDocument string
}

type DocViewer struct {
	app    *App
	dao    *mongo.Dao
	state  docState
	parent tview.Primitive
}

func NewDocView(dao *mongo.Dao) *DocViewer {
	return &DocViewer{
		dao: dao,
	}
}

func (d *DocViewer) Init(ctx context.Context, parent tview.Primitive) error {
	d.app = GetApp(ctx)
	d.parent = parent
	return nil
}

func (d *DocViewer) DocViewer(ctx context.Context, db, coll string, rawDocument string) error {
	d.state = docState{
		contentState: contentState{
			db:   db,
			coll: coll,
		},
		rawDocument: rawDocument,
	}
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

	root := GetApp(ctx).Root
	root.AddPage("details", modal, true, true)
	d.app.SetFocus(modal)
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Edit" {
			d.DocEdit(ctx, db, coll, rawDocument, d.refresh)
		} else {
			root.RemovePage("details")
			d.app.SetFocus(d.parent)
		}
	})
	return nil
}

func (d *DocViewer) refresh() {
	d.DocViewer(context.Background(), d.state.db, d.state.coll, d.state.rawDocument)
}

func (d *DocViewer) DocEdit(ctx context.Context, db, coll string, rawDocument string, fun func()) error {
	var prettyJson bytes.Buffer
	err := json.Indent(&prettyJson, []byte(rawDocument), "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return err
	}

	tmpFile, err := os.CreateTemp("", "doc-*.json")
	if err != nil {
		return fmt.Errorf("Error creating temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write(prettyJson.Bytes())
	if err != nil {
		return fmt.Errorf("Error writing to temp file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("Error closing temp file: %v", err)
	}

	editor, err := exec.LookPath(os.Getenv("EDITOR"))
	if err != nil {
		return fmt.Errorf("Error finding editor: %v", err)
	}

	d.app.Suspend(func() {
		cmd := exec.Command(editor, tmpFile.Name())
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Printf("Error running editor: %v", err)
			return
		}

		editedBytes, err := os.ReadFile(tmpFile.Name())
		if err != nil {
			log.Printf("Error reading edited file: %v", err)
			return
		}

		err = d.saveDocument(ctx, db, coll, string(editedBytes))
		if err != nil {
			log.Printf("Error saving edited document: %v", err)
			return
		} else {
			fun()
		}
	})

	return nil
}

func (d *DocViewer) saveDocument(ctx context.Context, db, coll string, rawDocument string) error {
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
