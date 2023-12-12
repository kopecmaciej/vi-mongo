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
	// func that will be called after document is saved
	// and should Render the desired view
	Render func() error
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

func (d *DocModifier) Add(ctx context.Context, db, coll string) error {
	createdDoc, err := d.openEditor(ctx, "{}")
	if err != nil {
		log.Printf("Error editing document: %v", err)
		return nil
	}
	if createdDoc == "" {
		log.Debug().Msgf("Document not created")
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
	if d.Render == nil {
		return "", fmt.Errorf("Render function not set")
	}
	var prettyJson bytes.Buffer
	err := json.Indent(&prettyJson, []byte(rawDocument), "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return "", nil
	}

	tmpFile, err := os.CreateTemp("", "doc-*.json")
	if err != nil {
		return "", fmt.Errorf("Error creating temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write(prettyJson.Bytes())
	if err != nil {
		return "", fmt.Errorf("Error writing to temp file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("Error closing temp file: %v", err)
	}

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
		if updatedDocument == string(prettyJson.Bytes()) {
			log.Debug().Msgf("Edited JSON is the same as original")
			return
		}

		err = d.saveDocument(ctx, db, coll, updatedDocument)
		if err != nil {
			log.Printf("Error saving edited document: %v", err)
			return
		}
	})

	return updatedDocument, nil
}

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

func (d *DocModifier) openEditor(ctx context.Context, rawDocument string) (string, error) {
	updatedDocument, err := d.openEditor(ctx, rawDocument)
	if err != nil {
		return "", fmt.Errorf("Error editing document: %v", err)
	}

	if updatedDocument == rawDocument {
		log.Debug().Msgf("Document not changed")
		return "", nil
	}

	return updatedDocument, nil
}

// saveDocument saves the document to the database
func (d *DocModifier) saveDocument(ctx context.Context, db, coll string, rawDocument string) error {
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
