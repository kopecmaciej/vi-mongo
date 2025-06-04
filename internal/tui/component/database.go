package component

import (
	"context"
	"fmt"
	"regexp"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
	"github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/tui/modal"
)

const (
	DatabaseId  = "Database"
	FilterBarId = "FilterBar"
)

// Database is flex container for DatabaseTree and InputBar
type Database struct {
	*core.BaseElement
	*core.Flex

	DbTree       *DatabaseTree
	filterBar    *InputBar
	mutex        sync.Mutex
	dbsWithColls []mongo.DBsWithCollections
}

func NewDatabase() *Database {
	d := &Database{
		BaseElement: core.NewBaseElement(),
		Flex:        core.NewFlex(),
		DbTree:      NewDatabaseTree(),
		filterBar:   NewInputBar(FilterBarId, "Filter"),
		mutex:       sync.Mutex{},
	}

	d.SetIdentifier(DatabaseId)
	d.SetAfterInitFunc(d.init)

	return d
}

func (d *Database) init() error {
	ctx := context.Background()
	d.setStyle()
	d.setKeybindings()

	if err := d.DbTree.Init(d.App); err != nil {
		return err
	}

	if err := d.filterBar.Init(d.App); err != nil {
		return err
	}
	d.filterBarHandler(ctx)

	d.handleEvents()

	return nil
}

func (d *Database) setStyle() {
	d.Flex.SetStyle(d.App.GetStyles())
	d.DbTree.SetStyle(d.App.GetStyles())
	d.filterBar.SetStyle(d.App.GetStyles())
	d.Flex.SetDirection(tview.FlexRow)
}

func (d *Database) setKeybindings() {
	keys := d.App.GetKeys()
	d.Flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case keys.Contains(keys.Database.FilterBar, event.Name()):
			d.filterBar.Enable()
			d.Render()
			return nil
		}
		return event
	})
}

func (d *Database) handleEvents() {
	go d.HandleEvents(DatabaseId, func(event manager.EventMsg) {
		switch event.Message.Type {
		case manager.StyleChanged:
			d.setStyle()
			d.DbTree.RefreshStyle()
		}
	})
}

func (d *Database) Render() {
	ctx := context.Background()
	d.Flex.Clear()

	var primitive tview.Primitive
	primitive = d.DbTree

	if d.filterBar.IsEnabled() {
		d.Flex.AddItem(d.filterBar, 3, 0, false)
		primitive = d.filterBar
	}
	defer d.App.SetFocus(primitive)

	if err := d.listDbsAndCollections(ctx); err != nil {
		// TODO: refactor how rendering is handled as this error will not be shown
		modal.ShowError(d.App.Pages, "Failed to list databases and collections", nil)
		d.dbsWithColls = []mongo.DBsWithCollections{}
	}

	d.DbTree.Render(ctx, d.dbsWithColls, false)

	d.Flex.AddItem(d.DbTree, 0, 1, true)
}

func (d *Database) IsFocused() bool {
	return d.App.GetFocus().GetIdentifier() == d.GetIdentifier() ||
		d.App.GetFocus().GetIdentifier() == d.DbTree.GetIdentifier()
}

func (d *Database) filterBarHandler(ctx context.Context) {
	accceptFunc := func(text string) {
		d.filter(ctx, text)
	}
	rejectFunc := func() {
		d.Render()
	}
	d.filterBar.DoneFuncHandler(accceptFunc, rejectFunc)
}

func (d *Database) filter(ctx context.Context, text string) {
	dbsWitColls := d.dbsWithColls
	expand := true
	filtered := []mongo.DBsWithCollections{}
	if text == "" {
		filtered = dbsWitColls
		expand = false
	} else {
		re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(text))
		for _, db := range dbsWitColls {
			matchedDB := re.MatchString(db.DB)
			matchedCollections := []string{}

			for _, coll := range db.Collections {
				if re.MatchString(coll) {
					matchedCollections = append(matchedCollections, coll)
				}
			}

			if matchedDB || len(matchedCollections) > 0 {
				filteredDB := mongo.DBsWithCollections{
					DB:          db.DB,
					Collections: matchedCollections,
				}
				if matchedDB {
					filteredDB.Collections = db.Collections
				}
				filtered = append(filtered, filteredDB)
				expand = expand || len(matchedCollections) > 0
			}
		}
	}
	d.DbTree.Render(ctx, filtered, expand)

	d.Flex.RemoveItem(d.filterBar)

	d.App.SetFocus(d.DbTree)
}

func (d *Database) listDbsAndCollections(ctx context.Context) error {
	dbsWitColls, err := d.Dao.ListDbsWithCollections(ctx, "")
	if err != nil {
		return err
	}
	d.dbsWithColls = dbsWitColls

	return nil
}

func (d *Database) SetSelectFunc(f func(ctx context.Context, db string, coll string) error) {
	d.DbTree.SetSelectFunc(f)
}

func (d *Database) NavigateToDbCollection(ctx context.Context, dbName, collectionName string) error {
	if err := d.listDbsAndCollections(ctx); err != nil {
		return fmt.Errorf("failed to load databases and collections: %w", err)
	}

	d.DbTree.Render(ctx, d.dbsWithColls, false)

	return d.DbTree.NavigateToDbCollection(ctx, dbName, collectionName)
}
