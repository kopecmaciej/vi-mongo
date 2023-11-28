package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

type Dao struct {
	client *mongo.Client
	Config *Config
}

func NewDao(client *mongo.Client, config *Config) *Dao {
	return &Dao{
		client: client,
		Config: config,
	}
}

func (d *Dao) GetServerStatus(ctx context.Context) (*ServerStatus, error) {
	var status ServerStatus
	err := d.client.Database("admin").RunCommand(ctx, primitive.D{{Key: "serverStatus", Value: 1}}).Decode(&status)
	if err != nil {
		return nil, err
	}

	isMaster, err := d.runAdminCommand(ctx, "isMaster", 1)
	if err != nil {
		return nil, err
	}
	status.Repl.ReadOnly = isMaster["readOnly"].(bool)
	status.Repl.IsMaster = isMaster["ismaster"].(bool)

	return &status, nil
}

func (d *Dao) GetLiveSessions(ctx context.Context) (int64, error) {
	results, err := d.runAdminCommand(ctx, "currentOp", 1)
	if err != nil {
		return 0, err
	}

	sessions := results["inprog"].(primitive.A)

	return int64(len(sessions)), nil
}

type DBsWithCollections struct {
	DB          string
	Collections []string
}

func (d *Dao) ListDbsWithCollections(ctx context.Context, nameRegex string) ([]DBsWithCollections, error) {
	dbCollMap := []DBsWithCollections{}

	filter := primitive.M{}
	if nameRegex != "" {
		filter = primitive.M{"name": primitive.Regex{Pattern: nameRegex, Options: "i"}}
	}

	dbs, err := d.client.ListDatabaseNames(ctx, filter)
	if err != nil {
		return nil, err
	}

	for _, db := range dbs {
		colls, err := d.client.Database(db).ListCollectionNames(ctx, primitive.M{})
		if err != nil {
			return nil, err
		}
		dbCollMap = append(dbCollMap, DBsWithCollections{DB: db, Collections: colls})
	}

	return dbCollMap, nil
}

type Filter struct {
	Key   string
	Value string
}

func (d *Dao) ListDocuments(ctx context.Context, db string, collection string, filter primitive.M, page, limit int64) ([]primitive.M, int64, error) {
	count, err := d.client.Database(db).Collection(collection).CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	coll := d.client.Database(db).Collection(collection)

	options := options.FindOptions{
		Limit: &limit,
		Skip:  &page,
		Sort:  primitive.D{{Key: "_id", Value: 1}},
	}

	cursor, err := coll.Find(ctx, filter, &options)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(nil)

	var documents []primitive.M
	for cursor.Next(nil) {
		var document primitive.M
		err := cursor.Decode(&document)
		if err != nil {
			return nil, 0, err
		}
		documents = append(documents, document)
	}
	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}
	return documents, count, nil
}

// save doc
func (d *Dao) UpdateDocument(ctx context.Context, db string, collection string, id primitive.ObjectID, document primitive.M) error {
	_, err := d.client.Database(db).Collection(collection).InsertOne(ctx, document)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) runAdminCommand(ctx context.Context, key string, value interface{}) (primitive.M, error) {
	results := primitive.M{}
	command := primitive.D{{Key: key, Value: value}}

	err := d.client.Database("admin").RunCommand(ctx, command).Decode(&results)
	if err != nil {
		return nil, err
	}

	return results, nil
}
