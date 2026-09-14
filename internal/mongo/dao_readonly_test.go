package mongo

import (
	"context"
	"errors"
	"testing"

	"github.com/kopecmaciej/vi-mongo/internal/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// newReadOnlyDao builds a Dao with a nil client, so any write reaching the
// driver would panic instead of silently passing the test
func newReadOnlyDao(readOnly bool) *Dao {
	return NewDao(nil, &config.MongoConfig{
		Options: config.MongoOptions{ReadOnly: &readOnly},
	})
}

func TestWritesBlockedWhenReadOnly(t *testing.T) {
	d := newReadOnlyDao(true)
	ctx := context.Background()

	tests := map[string]func() error{
		"InsetDocument":    func() error { _, err := d.InsetDocument(ctx, "db", "coll", bson.M{}); return err },
		"UpdateDocument":   func() error { return d.UpdateDocument(ctx, "db", "coll", "id", bson.M{}, bson.M{"a": 1}) },
		"DeleteDocument":   func() error { return d.DeleteDocument(ctx, "db", "coll", "id") },
		"AddCollection":    func() error { return d.AddCollection(ctx, "db", "coll") },
		"DeleteCollection": func() error { return d.DeleteCollection(ctx, "db", "coll") },
		"RenameCollection": func() error { return d.RenameCollection(ctx, "db", "old", "new") },
		"CreateIndex":      func() error { return d.CreateIndex(ctx, "db", "coll", mongo.IndexModel{}) },
		"DropIndex":        func() error { return d.DropIndex(ctx, "db", "coll", "idx") },
		"AggregateOut": func() error {
			_, err := d.AggregateDocuments(ctx, "db", "coll", mongo.Pipeline{bson.D{{Key: "$out", Value: "other"}}})
			return err
		},
		"AggregateMerge": func() error {
			_, err := d.AggregateDocuments(ctx, "db", "coll", mongo.Pipeline{bson.D{{Key: "$merge", Value: "other"}}})
			return err
		},
	}

	for name, call := range tests {
		t.Run(name, func(t *testing.T) {
			if err := call(); !errors.Is(err, ErrReadOnly) {
				t.Errorf("expected ErrReadOnly, got %v", err)
			}
		})
	}
}

func TestReadOnlyDefaultsToFalse(t *testing.T) {
	d := NewDao(nil, &config.MongoConfig{})
	if err := d.ensureWritable(); err != nil {
		t.Errorf("expected writes allowed by default, got %v", err)
	}
}

func TestFindWriteStage(t *testing.T) {
	tests := []struct {
		name     string
		pipeline mongo.Pipeline
		want     string
	}{
		{"read only pipeline", mongo.Pipeline{bson.D{{Key: "$match", Value: bson.M{}}}}, ""},
		{"out at the end", mongo.Pipeline{bson.D{{Key: "$match", Value: bson.M{}}}, bson.D{{Key: "$out", Value: "c"}}}, "$out"},
		{"merge stage", mongo.Pipeline{bson.D{{Key: "$merge", Value: bson.M{}}}}, "$merge"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stage, found := findWriteStage(tt.pipeline)
			if found != (tt.want != "") || stage != tt.want {
				t.Errorf("got (%q, %v), want %q", stage, found, tt.want)
			}
		})
	}
}
