package mongo

import (
	"sort"

	"github.com/kopecmaciej/vi-mongo/internal/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CollectionState struct {
	Db     string
	Coll   string
	Page   int64
	Limit  int64
	Count  int64
	Sort   string
	Filter string
	Docs   map[string]primitive.M
}

func (c *CollectionState) UpdateFilter(filter string) {
	filter = util.CleanJsonWhitespaces(filter)
	if util.IsJsonEmpty(filter) {
		c.Filter = ""
		return
	}
	c.Filter = filter
	c.Page = 0
}

func (c *CollectionState) UpdateSort(sort string) {
	sort = util.CleanJsonWhitespaces(sort)
	if util.IsJsonEmpty(sort) {
		c.Sort = ""
		return
	}
	c.Sort = sort
}

func (c *CollectionState) GetSortedDocs() []primitive.M {
	keys := make([]string, 0, len(c.Docs))
	for k := range c.Docs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	docs := make([]primitive.M, 0, len(keys))
	for _, k := range keys {
		docs = append(docs, copyDoc(c.Docs[k]))
	}
	return docs
}

func (c *CollectionState) GetDocById(id interface{}) primitive.M {
	doc, ok := c.Docs[StringifyId(id)]
	if !ok {
		return nil
	}
	return copyDoc(doc)
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

func (c *CollectionState) PopulateDocs(docs []primitive.M) {
	c.Docs = make(map[string]primitive.M)
	for _, doc := range docs {
		c.Docs[StringifyId(doc["_id"])] = copyDoc(doc)
	}
}

func (c *CollectionState) UpdateRawDoc(doc string) error {
	docMap, err := ParseJsonToBson(doc)
	if err != nil {
		return err
	}
	c.Docs[StringifyId(docMap["_id"])] = docMap
	return nil
}

func (c *CollectionState) AppendDoc(doc primitive.M) {
	if c.Docs == nil {
		c.Docs = make(map[string]primitive.M)
	}
	c.Docs[StringifyId(doc["_id"])] = doc
	c.Count++
}

func (c *CollectionState) DeleteDoc(id interface{}) {
	delete(c.Docs, StringifyId(id))
	c.Count--
}

// Helper function to create a deep copy of a primitive.M
func copyDoc(doc primitive.M) primitive.M {
	docCopy := make(primitive.M)
	for key, value := range doc {
		docCopy[key] = value
	}
	return docCopy
}
