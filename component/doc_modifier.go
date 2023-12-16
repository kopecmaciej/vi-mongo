package component

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/kopecmaciej/mongui/mongo"
	"github.com/rs/zerolog/log"
)

// DocModifier is a component that allows editing JSON documents
type DocModifier struct {
	dao *mongo.Dao

	app *App
}

func NewDocModifier(dao *mongo.Dao) *DocModifier {
	return &DocModifier{
		dao: dao,
	}
}

func (d *DocModifier) Init(ctx context.Context) error {
	app, err := GetApp(ctx)
	if err != nil {
		return err
	}
	d.app = app
	return nil
}

func (d *DocModifier) Insert(ctx context.Context, db, coll string) error {
	createdDoc, err := d.openEditor(ctx, "{}")
	if err != nil {
		log.Printf("Error editing document: %v", err)
		return nil
	}
	if createdDoc == "{}" {
		log.Debug().Msgf("No document created")
		return nil
	}

	var document map[string]interface{}
	err = json.Unmarshal([]byte(createdDoc), &document)
	if err != nil {
		return fmt.Errorf("Error unmarshaling JSON: %v", err)
	}

	err = d.dao.InsetDocument(ctx, db, coll, document)
	if err != nil {
		return fmt.Errorf("Error inserting document: %v", err)
	}

	return nil
}

// Edit opens the editor with the document and saves it if it was changed
func (d *DocModifier) Edit(ctx context.Context, db, coll string, rawDocument string) (string, error) {
	updatedDocument, err := d.openEditor(ctx, rawDocument)
	if err != nil {
		return "", fmt.Errorf("Error editing document: %v", err)
	}

	if updatedDocument == rawDocument {
		log.Debug().Msgf("Edited JSON is the same as original")
		return "", nil
	}

	err = d.updateDocument(ctx, db, coll, updatedDocument)
	if err != nil {
		return "", fmt.Errorf("Error saving document: %v", err)
	}

	return updatedDocument, nil
}

// Duplicate opens the editor with the document and saves it as a new document
func (d *DocModifier) Duplicate(ctx context.Context, db, coll string, rawDocument string) error {
	replacedDoc, err := removeField(rawDocument, "_id")

	duplicateDoc, err := d.openEditor(ctx, replacedDoc)
	if err != nil {
		return fmt.Errorf("Error editing document: %v", err)
	}
	if duplicateDoc == "" {
		log.Debug().Msgf("Document not duplicated")
		return nil
	}

	var document map[string]interface{}
	err = json.Unmarshal([]byte(duplicateDoc), &document)
	if err != nil {
		return fmt.Errorf("Error unmarshaling JSON: %v", err)
	}

	delete(document, "_id")

	err = d.dao.InsetDocument(ctx, db, coll, document)
	if err != nil {
		return fmt.Errorf("Error inserting document: %v", err)
	}

	return nil
}

// updateDocument saves the document to the database
func (d *DocModifier) updateDocument(ctx context.Context, db, coll string, rawDocument string) error {
	if rawDocument == "" {
		return fmt.Errorf("Document cannot be empty")
	}
	var document map[string]interface{}
	err := json.Unmarshal([]byte(rawDocument), &document)
	if err != nil {
		log.Error().Msgf("Error unmarshaling JSON: %v", err)
		return nil
	}

	id, err := mongo.GetIDFromDocument(document)
	if err != nil {
		log.Error().Msgf("Error getting _id from document: %v", err)
		return nil
	}
	delete(document, "_id")

	err = d.dao.UpdateDocument(ctx, db, coll, id, document)
	if err != nil {
		log.Error().Msgf("Error updating document: %v", err)
		return nil
	}

	return nil
}

// openEditor opens the editor with the document and returns the edited document
func (d *DocModifier) openEditor(ctx context.Context, rawDocument string) (string, error) {
	prettyJsonBuffer, err := mongo.IndientJSON(rawDocument)
	if err != nil {
		return "", fmt.Errorf("Error indenting JSON: %v", err)
	}

	tmpFile, err := d.writeToTempFile(prettyJsonBuffer)
	if err != nil {
		return "", fmt.Errorf("Error writing to temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	editor, err := exec.LookPath(os.Getenv("EDITOR"))
	if err != nil {
		return "", fmt.Errorf("Error looking for editor: %v", err)
	}

	updatedDocument := ""

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
		if !json.Valid(editedBytes) {
			log.Printf("Edited JSON is not valid")
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
		return nil, fmt.Errorf("Error creating temp file: %v", err)
	}

	_, err = tmpFile.Write(bufferJson.Bytes())
	if err != nil {
		err = os.Remove(tmpFile.Name())
		if err != nil {
			return nil, fmt.Errorf("Error removing temp file: %v", err)
		}
		return nil, fmt.Errorf("Error writing to temp file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("Error closing temp file: %v", err)
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
