package mongo

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrExportFileExists = errors.New("export file already exists")

type jsonArrayWriter struct {
	writer *bufio.Writer
	count  int
}

func newJSONArrayWriter(writer io.Writer) (*jsonArrayWriter, error) {
	arrayWriter := &jsonArrayWriter{writer: bufio.NewWriter(writer)}
	if _, err := arrayWriter.writer.WriteString("[\n"); err != nil {
		return nil, fmt.Errorf("write JSON array start: %w", err)
	}
	return arrayWriter, nil
}

func (w *jsonArrayWriter) WriteDocument(document primitive.M) error {
	data, err := bson.MarshalExtJSON(document, false, false)
	if err != nil {
		return fmt.Errorf("marshal document %d: %w", w.count+1, err)
	}

	var indented bytes.Buffer
	if err := json.Indent(&indented, data, "  ", "  "); err != nil {
		return fmt.Errorf("indent document %d: %w", w.count+1, err)
	}
	if w.count > 0 {
		if _, err := w.writer.WriteString(",\n"); err != nil {
			return fmt.Errorf("write document separator: %w", err)
		}
	}
	if _, err := w.writer.WriteString("  "); err != nil {
		return fmt.Errorf("write document indentation: %w", err)
	}
	if _, err := w.writer.Write(indented.Bytes()); err != nil {
		return fmt.Errorf("write document %d: %w", w.count+1, err)
	}
	w.count++
	return nil
}

func (w *jsonArrayWriter) Close() error {
	if _, err := w.writer.WriteString("\n]\n"); err != nil {
		return fmt.Errorf("write JSON array end: %w", err)
	}
	if err := w.writer.Flush(); err != nil {
		return fmt.Errorf("flush export: %w", err)
	}
	return nil
}

// ExportDocuments writes documents as a relaxed MongoDB Extended JSON array.
func ExportDocuments(writer io.Writer, documents []primitive.M) error {
	arrayWriter, err := newJSONArrayWriter(writer)
	if err != nil {
		return err
	}
	for _, document := range documents {
		if err := arrayWriter.WriteDocument(document); err != nil {
			return err
		}
	}
	return arrayWriter.Close()
}

// ExportDocumentsFile atomically replaces path after the complete export has
// been written successfully. Existing files require overwrite to be true.
func ExportDocumentsFile(path string, documents []primitive.M, overwrite bool) error {
	_, err := exportDocumentsFile(path, overwrite, func(writeDocument func(primitive.M) error) error {
		for _, document := range documents {
			if err := writeDocument(document); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// ExportDocumentsFileFromIterator streams documents from iterate into an
// atomic JSON array export and returns the number of exported documents.
func ExportDocumentsFileFromIterator(
	path string,
	overwrite bool,
	iterate func(func(primitive.M) error) error,
) (int, error) {
	return exportDocumentsFile(path, overwrite, iterate)
}

func exportDocumentsFile(
	path string,
	overwrite bool,
	iterate func(func(primitive.M) error) error,
) (int, error) {
	if path == "" {
		return 0, errors.New("export path cannot be empty")
	}
	if _, err := os.Stat(path); err == nil {
		if !overwrite {
			return 0, ErrExportFileExists
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return 0, fmt.Errorf("inspect export path: %w", err)
	}

	temporary, err := os.CreateTemp(filepath.Dir(path), ".vi-mongo-export-*")
	if err != nil {
		return 0, fmt.Errorf("create temporary export file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()

	if err := temporary.Chmod(0o644); err != nil {
		_ = temporary.Close()
		return 0, fmt.Errorf("set export permissions: %w", err)
	}
	arrayWriter, err := newJSONArrayWriter(temporary)
	if err != nil {
		_ = temporary.Close()
		return 0, err
	}
	if err := iterate(arrayWriter.WriteDocument); err != nil {
		_ = temporary.Close()
		return 0, err
	}
	if err := arrayWriter.Close(); err != nil {
		_ = temporary.Close()
		return 0, err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return 0, fmt.Errorf("sync export file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return 0, fmt.Errorf("close export file: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return 0, fmt.Errorf("replace export file: %w", err)
	}
	return arrayWriter.count, nil
}
