package mongo

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestExportDocumentsJSONArray(t *testing.T) {
	id := primitive.NewObjectID()
	documents := []primitive.M{
		{"_id": id, "createdAt": primitive.NewDateTimeFromTime(time.Unix(0, 0).UTC())},
		{"name": "second"},
	}

	var output bytes.Buffer
	require.NoError(t, ExportDocuments(&output, documents, true))
	require.Contains(t, output.String(), `"$oid": "`+id.Hex()+`"`)
	require.Contains(t, output.String(), `"$date": "1970-01-01T00:00:00Z"`)
	require.Contains(t, output.String(), "},\n  {")

	var decoded []any
	require.NoError(t, json.Unmarshal([]byte(output.String()), &decoded))
	require.Len(t, decoded, 2)
}

func TestExportDocumentsEmptyJSONArray(t *testing.T) {
	var output bytes.Buffer
	require.NoError(t, ExportDocuments(&output, nil, true))
	require.Equal(t, "[\n\n]\n", output.String())
}

func TestExportDocumentsCompactJSONArray(t *testing.T) {
	documents := []primitive.M{{"name": "first"}, {"name": "second"}}

	var output bytes.Buffer
	require.NoError(t, ExportDocuments(&output, documents, false))
	require.Equal(t, `[{"name":"first"},{"name":"second"}]`, output.String())
}

func TestExportDocumentsEmptyCompactJSONArray(t *testing.T) {
	var output bytes.Buffer
	require.NoError(t, ExportDocuments(&output, nil, false))
	require.Equal(t, "[]", output.String())
}

func TestExportDocumentsFileRequiresOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "export.json")
	require.NoError(t, os.WriteFile(path, []byte("original"), 0o600))

	err := ExportDocumentsFile(path, []primitive.M{{"name": "new"}}, false, true)
	require.True(t, errors.Is(err, ErrExportFileExists))
	content, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	require.Equal(t, "original", string(content))

	require.NoError(t, ExportDocumentsFile(path, []primitive.M{{"name": "new"}}, true, true))
	content, readErr = os.ReadFile(path)
	require.NoError(t, readErr)
	require.Contains(t, string(content), `"name": "new"`)
}

func TestExportDocumentsFileFromIterator(t *testing.T) {
	path := filepath.Join(t.TempDir(), "export.json")
	documents := []primitive.M{{"page": 1}, {"page": 2}, {"page": 3}}

	count, err := ExportDocumentsFileFromIterator(path, false, true, func(writeDocument func(primitive.M) error) error {
		for _, document := range documents {
			if err := writeDocument(document); err != nil {
				return err
			}
		}
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 3, count)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	var decoded []any
	require.NoError(t, json.Unmarshal(content, &decoded))
	require.Len(t, decoded, 3)
}
