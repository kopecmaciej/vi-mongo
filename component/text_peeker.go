package component

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/kopecmaciej/mongui/manager"
	"github.com/kopecmaciej/mongui/mongo"
	"github.com/kopecmaciej/mongui/primitives"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	TextPeekerComponent manager.Component = "TextPeeker"
)

type docState struct {
	contentState
	rawDocument string
}

type TextPeeker struct {
	app     *App
	dao     *mongo.Dao
	state   docState
	parent  tview.Primitive
	manager *manager.ComponentManager
}

func NewTextPeeker(dao *mongo.Dao) *TextPeeker {
	return &TextPeeker{
		dao: dao,
	}
}

func (d *TextPeeker) Init(ctx context.Context, parent tview.Primitive) error {
	app, err := GetApp(ctx)
	if err != nil {
		return err
	}
	d.app = app

	d.manager = d.app.ComponentManager
	d.parent = parent
	return nil
}

func (d *TextPeeker) PeekJson(ctx context.Context, db, coll string, jsonString string) error {
	d.state = docState{
		contentState: contentState{
			db:   db,
			coll: coll,
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

	modal := primitives.NewModalView()
	modal.SetBorder(true)
	modal.SetTitle("Document Details")
	modal.SetTitleAlign(tview.AlignLeft)
	modal.SetTitleColor(tcell.ColorSteelBlue)

	modal.SetText(primitives.Text{
		Content: text,
		Color:   tcell.ColorWhite,
		Align:   tview.AlignLeft,
	})

	modal.AddButtons([]string{"Edit", "Close"})

	root := d.app.Root
	root.AddPage(TextPeekerComponent, modal, true, true)
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Edit" {
			d.EditJson(ctx, db, coll, jsonString, d.refresh)
		} else {
			root.RemovePage(TextPeekerComponent)
		}
	})
	return nil
}

func (d *TextPeeker) refresh(ctx context.Context) error {
	return d.PeekJson(ctx, d.state.db, d.state.coll, d.state.rawDocument)
}

// EditJson opens the editor with the document and saves it if it was changed
func (tp *TextPeeker) EditJson(ctx context.Context, db, coll string, rawDocument string, render func(ctx context.Context) error) error {
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

	tp.app.Suspend(func() {
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
		if !json.Valid(editedBytes) {
			log.Printf("Edited JSON is not valid")
			return
		}
		if string(editedBytes) == string(prettyJson.Bytes()) {
			log.Debug().Msgf("Edited JSON is the same as original")
			return
		}

		err = tp.saveDocument(ctx, db, coll, string(editedBytes))
		if err != nil {
			log.Printf("Error saving edited document: %v", err)
			return
		} else {
			log.Debug().Msg("Document saved")
			err := render(ctx)
			if err != nil {
				// TODO: show modal with error
				log.Printf("Error rendering: %v", err)
				return
			}
		}

	})

	return nil
}

func (tp *TextPeeker) saveDocument(ctx context.Context, db, coll string, rawDocument string) error {
	if rawDocument == "" {
		return fmt.Errorf("Document cannot be empty")
	}
	var document map[string]interface{}
	err := json.Unmarshal([]byte(rawDocument), &document)
	if err != nil {
		log.Error().Msgf("Error unmarshaling JSON: %v", err)
		return nil
	}
	
	id := document["_id"].(map[string]interface{})["$oid"].(string)
	delete(document, "_id")

	if id == "" {
		log.Error().Msgf("Document must have an _id")
	}
	mongoId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Error().Msgf("Invalid _id: %v", err)
	}

	err = tp.dao.UpdateDocument(ctx, db, coll, mongoId, document)
	if err != nil {
		log.Error().Msgf("Error updating document: %v", err)
		return nil
	}

	return nil
}
