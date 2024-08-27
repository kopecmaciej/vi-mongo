package component

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	DocModifierView = "DocModifier"
)

// DocModifier is a view that allows editing JSON documents
type DocModifier struct {
	*core.BaseElement
}

func NewDocModifier() *DocModifier {
	return &DocModifier{
		BaseElement: core.NewBaseElement(DocModifierView),
	}
}

func (d *DocModifier) Insert(ctx context.Context, db, coll string) (primitive.ObjectID, error) {
	createdDoc, err := d.openEditor("{}")
	if err != nil {
		log.Error().Err(err).Msg("Error opening editor")
		return primitive.NilObjectID, nil
	}
	if strings.ReplaceAll(createdDoc, " ", "") == "{}" {
		log.Debug().Msgf("No document created")
		return primitive.NilObjectID, nil
	}

	var document map[string]interface{}
	err = json.Unmarshal([]byte(createdDoc), &document)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	rawID, err := d.Dao.InsetDocument(ctx, db, coll, document)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("error inserting document: %v", err)
	}

	ID, ok := rawID.(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, fmt.Errorf("error converting _id to primitive.ObjectID")
	}

	return ID, nil
}

// Edit opens the editor with the document and saves it if it was changed
func (d *DocModifier) Edit(ctx context.Context, db, coll string, rawDocument string) (string, error) {
	updatedDocument, err := d.openEditor(rawDocument)
	if err != nil {
		return "", fmt.Errorf("error editing document: %v", err)
	}

	if strings.ReplaceAll(updatedDocument, " ", "") == strings.ReplaceAll(rawDocument, " ", "") {
		log.Debug().Msgf("Edited JSON is the same as original")
		return "", nil
	}

	err = d.updateDocument(ctx, db, coll, updatedDocument)
	if err != nil {
		return "", fmt.Errorf("error saving document: %v", err)
	}

	return updatedDocument, nil
}

// Duplicate opens the editor with the document and saves it as a new document
func (d *DocModifier) Duplicate(ctx context.Context, db, coll string, rawDocument string) (primitive.ObjectID, error) {
	replacedDoc, err := removeField(rawDocument, "_id")
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("error removing _id field: %v", err)
	}

	duplicateDoc, err := d.openEditor(replacedDoc)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("error editing document: %v", err)
	}
	if duplicateDoc == "" {
		log.Debug().Msgf("Document not duplicated")
		return primitive.NilObjectID, nil
	}

	parsedDoc, err := mongo.ParseJsonToBson(duplicateDoc)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("error parsing JSON: %v", err)
	}

	delete(parsedDoc, "_id")

	rawID, err := d.Dao.InsetDocument(ctx, db, coll, parsedDoc)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("error inserting document: %v", err)
	}

	ID, ok := rawID.(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, fmt.Errorf("error converting _id to primitive.ObjectID")
	}

	return ID, nil
}

// updateDocument saves the document to the database
func (d *DocModifier) updateDocument(ctx context.Context, db, coll string, rawDocument string) error {
	if rawDocument == "" {
		return fmt.Errorf("document cannot be empty")
	}

	parsedDoc, err := mongo.ParseJsonToBson(rawDocument)
	if err != nil {
		return fmt.Errorf("error parsing JSON: %v", err)
	}
	id, err := mongo.GetIDFromDocument(parsedDoc)
	if err != nil {
		return fmt.Errorf("error getting _id from document: %v", err)
	}
	err = d.Dao.UpdateDocument(ctx, db, coll, id, parsedDoc)
	if err != nil {
		log.Error().Msgf("error updating document: %v", err)
		return nil
	}

	return nil
}

// openEditor opens the editor with the document and returns the edited document
func (d *DocModifier) openEditor(rawDocument string) (string, error) {
	prettyJsonBuffer, err := mongo.IndentJson(rawDocument)
	if err != nil {
		return "", fmt.Errorf("error indenting JSON: %v", err)
	}

	tmpFile, err := d.writeToTempFile(prettyJsonBuffer)
	if err != nil {
		return "", fmt.Errorf("error writing to temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	ed, err := d.App.GetConfig().GetEditorCmd()
	if err != nil {
		return "", fmt.Errorf("error getting editor command: %v", err)
	}
	editor, err := exec.LookPath(ed)
	if err != nil {
		return "", fmt.Errorf("error looking for editor: %v", err)
	}

	updatedDocument := ""

	d.App.Suspend(func() {
		cmd := exec.Command(editor, tmpFile.Name())
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Error().Err(err).Msg("error running editor")
			return
		}

		editedBytes, err := os.ReadFile(tmpFile.Name())
		if err != nil {
			log.Error().Err(err).Msg("error reading edited file")
			return
		}
		if !json.Valid(editedBytes) {
			log.Error().Msg("Edited JSON is not valid")
			return
		}
		updatedDocument = string(editedBytes)
	})

	return updatedDocument, nil
}

// writeToTempFile writes the JSON to a temp file and returns the file
func (d *DocModifier) writeToTempFile(bufferJson bytes.Buffer) (*os.File, error) {
	tmpFile, err := os.CreateTemp("", "doc-*.json")
	if err != nil {
		return nil, fmt.Errorf("error creating temp file: %v", err)
	}

	_, err = tmpFile.Write(bufferJson.Bytes())
	if err != nil {
		err = os.Remove(tmpFile.Name())
		if err != nil {
			return nil, fmt.Errorf("error removing temp file: %v", err)
		}
		return nil, fmt.Errorf("error writing to temp file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("error closing temp file: %v", err)
	}

	return tmpFile, nil
}

// removeField removes the specified field from a JSON string.
func removeField(jsonStr, fieldToRemove string) (string, error) {
	// Unmarshal the JSON into a map
	var data map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return "", err
	}

	// Remove the specified field
	delete(data, fieldToRemove)

	// Marshal the map back into a JSON string
	modifiedJSON, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(modifiedJSON), nil
}
