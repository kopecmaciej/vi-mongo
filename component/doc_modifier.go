package component

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/kopecmaciej/mongui/manager"
	"github.com/kopecmaciej/mongui/mongo"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	DocModifierComponent manager.Component = "DocModifier"
)

// DocModifier is a component that allows editing JSON documents
type DocModifier struct {
	*Component
}

func NewDocModifier() *DocModifier {
	return &DocModifier{
		Component: NewComponent(DocModifierComponent),
	}
}

func (d *DocModifier) Insert(ctx context.Context, db, coll string) (primitive.ObjectID, error) {
	createdDoc, err := d.openEditor("{}")
	if err != nil {
		log.Error().Err(err).Msg("Error opening editor")
		return primitive.NilObjectID, nil
	}
	if createdDoc == "{}" {
		log.Debug().Msgf("No document created")
		return primitive.NilObjectID, nil
	}

	document, err := mongo.JSONToMongo(createdDoc)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("error converting JSON to MongoDB document: %v", err)
	}

	rawID, err := d.dao.InsetDocument(ctx, db, coll, document)
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

	if updatedDocument == rawDocument {
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
	mongoDoc, err := mongo.JSONToMongo(rawDocument)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("error converting JSON to MongoDB document: %v", err)
	}

	delete(mongoDoc, "_id")

	jsonDoc, err := mongo.MongoToJSON(mongoDoc)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("error converting MongoDB document to JSON: %v", err)
	}

	duplicateDoc, err := d.openEditor(jsonDoc)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("error editing document: %v", err)
	}
	if duplicateDoc == "" {
		log.Debug().Msgf("Document not duplicated")
		return primitive.NilObjectID, nil
	}

	document, err := mongo.JSONToMongo(duplicateDoc)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("error converting JSON to MongoDB document: %v", err)
	}

	rawID, err := d.dao.InsetDocument(ctx, db, coll, document)
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

	document, err := mongo.JSONToMongo(rawDocument)
	if err != nil {
		return fmt.Errorf("error converting JSON to MongoDB document: %v", err)
	}

	id, ok := document["_id"].(primitive.ObjectID)
	if !ok {
		return fmt.Errorf("error getting _id from document")
	}
	delete(document, "_id")

	err = d.dao.UpdateDocument(ctx, db, coll, id, document)
	if err != nil {
		return fmt.Errorf("error updating document: %v", err)
	}

	return nil
}

// openEditor opens the editor with the document and returns the edited document
func (d *DocModifier) openEditor(rawDocument string) (string, error) {
	mongoDoc, err := mongo.JSONToMongo(rawDocument)
	if err != nil {
		return "", fmt.Errorf("error converting JSON to MongoDB document: %v", err)
	}

	jsonDoc, err := bson.MarshalExtJSON(mongoDoc, true, true)
	if err != nil {
		return "", fmt.Errorf("error marshaling MongoDB document to JSON: %v", err)
	}

	tmpFile, err := d.writeToTempFile(jsonDoc)
	if err != nil {
		return "", fmt.Errorf("error writing to temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	ed, err := d.app.Config.GetEditorCmd()
	if err != nil {
		return "", fmt.Errorf("error getting editor command: %v", err)
	}
	editor, err := exec.LookPath(ed)
	if err != nil {
		return "", fmt.Errorf("error looking for editor: %v", err)
	}

	updatedDocument := ""

	d.app.Suspend(func() {
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
func (d *DocModifier) writeToTempFile(bufferJson []byte) (*os.File, error) {
	tmpFile, err := os.CreateTemp("", "doc-*.json")
	if err != nil {
		return nil, fmt.Errorf("error creating temp file: %v", err)
	}

	_, err = tmpFile.Write(bufferJson)
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
