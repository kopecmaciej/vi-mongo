package mongo

import "go.mongodb.org/mongo-driver/bson/primitive"

type ServerStatus struct {
	Ok             int32  `bson:"ok"`
	Version        string `bson:"version"`
	Uptime         int32  `bson:"uptime"`
	CurrentConns   int32  `bson:"connections.current"`
	AvailableConns int32  `bson:"connections.available"`
	OpCounters     struct {
		Insert int32 `bson:"insert"`
		Query  int32 `bson:"query"`
		Update int32 `bson:"update"`
		Delete int32 `bson:"delete"`
	} `bson:"opcounters"`
	Mem struct {
		Resident int32 `bson:"resident"`
		Virtual  int32 `bson:"virtual"`
	} `bson:"mem"`
	Repl struct {
		ReadOnly bool `bson:"readOnly"`
		IsMaster bool `bson:"ismaster"`
	} `bson:"repl"`
}

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

func (c *CollectionState) GetDocById(id interface{}) primitive.M {
	return c.Docs[StringifyId(id)]
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
		c.Docs[StringifyId(doc["_id"])] = doc
	}
}
