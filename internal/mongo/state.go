package mongo

import (
	"reflect"
	"sync"

	"github.com/kopecmaciej/vi-mongo/internal/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CollectionState is used to store the state of a collection and use it
// while rendering doesn't require fetching from the database
type CollectionState struct {
	Db         string
	Coll       string
	Page       int64
	Limit      int64
	Count      int64
	Sort       string
	Filter     string
	Projection string
	// docs are only one private as they cannot be changed in uncontrolled way
	docs []primitive.M
}

func (c *CollectionState) GetAllDocs() []primitive.M {
	docsCopy := make([]primitive.M, len(c.docs))
	for i, doc := range c.docs {
		docsCopy[i] = deepCopy(doc)
	}
	return docsCopy
}

func (c *CollectionState) GetDocById(id interface{}) primitive.M {
	for _, doc := range c.docs {
		if reflect.TypeOf(doc["_id"]) == reflect.TypeOf(id) {
			if doc["_id"] == id {
				return deepCopy(doc)
			}
		}
	}
	return nil
}

func (c *CollectionState) GetJsonDocById(id interface{}) (string, error) {
	doc := c.GetDocById(id)
	jsoned, err := ParseBsonDocument(doc)
	if err != nil {
		return "", err
	}
	indentedJson, err := IndentJson(jsoned)
	if err != nil {
		return "", err
	}
	return indentedJson.String(), nil
}

func NewCollectionState(db, coll string) *CollectionState {
	return &CollectionState{
		Db:   db,
		Coll: coll,
		Page: 0,
	}
}

func (c *CollectionState) SetFilter(filter string) {
	filter = util.CleanJsonWhitespaces(filter)
	if util.IsJsonEmpty(filter) {
		c.Filter = ""
		return
	}
	c.Filter = filter
	c.Page = 0
}

func (c *CollectionState) SetSort(sort string) {
	sort = util.CleanJsonWhitespaces(sort)
	if util.IsJsonEmpty(sort) {
		c.Sort = ""
		return
	}
	c.Sort = sort
}

func (c *CollectionState) SetProjection(projection string) {
	projection = util.CleanJsonWhitespaces(projection)
	if util.IsJsonEmpty(projection) {
		c.Projection = ""
		return
	}
	c.Projection = projection
}

func (c *CollectionState) PopulateDocs(docs []primitive.M) {
	c.docs = make([]primitive.M, len(docs))
	for i, doc := range docs {
		c.docs[i] = deepCopy(doc)
	}
}

func (c *CollectionState) UpdateRawDoc(doc string) error {
	docMap, err := ParseJsonToBson(doc)
	if err != nil {
		return err
	}
	for i, existingDoc := range c.docs {
		if existingDoc["_id"] == docMap["_id"] {
			c.docs[i] = docMap
			return nil
		}
	}
	c.docs = append(c.docs, docMap)
	return nil
}

func (c *CollectionState) AppendDoc(doc primitive.M) {
	c.docs = append(c.docs, doc)
	c.Count++
}

func (c *CollectionState) DeleteDoc(id interface{}) {
	for i, doc := range c.docs {
		if doc["_id"] == id {
			c.docs = append(c.docs[:i], c.docs[i+1:]...)
			c.Count--
			return
		}
	}
}
func deepCopy(doc primitive.M) primitive.M {
	docCopy := make(primitive.M)
	for key, value := range doc {
		docCopy[key] = value
	}
	return docCopy
}

// StateMap persevere states when hopping between diffrent mongodb servers
type StateMap struct {
	mu            sync.RWMutex
	states        map[string]*CollectionState
	hiddenColumns map[string][]string
}

func (sm *StateMap) AddHiddenColumn(db, coll, column string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	key := sm.Key(db, coll)
	sm.hiddenColumns[key] = append(sm.hiddenColumns[key], column)
}

func (sm *StateMap) GetHiddenColumns(db, coll string) []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	key := sm.Key(db, coll)
	return sm.hiddenColumns[key]
}

func (sm *StateMap) ResetHiddenColumns(db, coll string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	key := sm.Key(db, coll)
	sm.hiddenColumns[key] = nil
}

func NewStateMap() *StateMap {
	return &StateMap{
		states:        make(map[string]*CollectionState),
		hiddenColumns: make(map[string][]string),
	}
}

func (sm *StateMap) Get(key string) (*CollectionState, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	state, ok := sm.states[key]
	return state, ok
}

func (sm *StateMap) Set(key string, state *CollectionState) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.states[key] = state
}

func (sm *StateMap) Key(db, coll string) string {
	return db + "." + coll
}
