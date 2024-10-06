package mongo

import (
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
	Docs   []primitive.M
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
	return c.Docs
}

func (c *CollectionState) GetDocById(id interface{}) primitive.M {
	for _, doc := range c.Docs {
		if doc["_id"] == id {
			return copyDoc(doc)
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

func (c *CollectionState) PopulateDocs(docs []primitive.M) {
	c.Docs = make([]primitive.M, len(docs))
	for i, doc := range docs {
		c.Docs[i] = copyDoc(doc)
	}
}

func (c *CollectionState) UpdateRawDoc(doc string) error {
	docMap, err := ParseJsonToBson(doc)
	if err != nil {
		return err
	}
	for i, existingDoc := range c.Docs {
		if existingDoc["_id"] == docMap["_id"] {
			c.Docs[i] = docMap
			return nil
		}
	}
	c.Docs = append(c.Docs, docMap)
	return nil
}

func (c *CollectionState) AppendDoc(doc primitive.M) {
	c.Docs = append(c.Docs, doc)
	c.Count++
}

func (c *CollectionState) DeleteDoc(id interface{}) {
	for i, doc := range c.Docs {
		if doc["_id"] == id {
			c.Docs = append(c.Docs[:i], c.Docs[i+1:]...)
			c.Count--
			return
		}
	}
}

// Helper function to create a deep copy of a primitive.M
func copyDoc(doc primitive.M) primitive.M {
	docCopy := make(primitive.M)
	for key, value := range doc {
		docCopy[key] = value
	}
	return docCopy
}
