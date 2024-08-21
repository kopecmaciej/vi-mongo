package mongo

import (
	"context"

	"github.com/kopecmaciej/mongui/internal/config"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Dao struct {
	client *mongo.Client
	Config *config.MongoConfig
}

func NewDao(client *mongo.Client, config *config.MongoConfig) *Dao {
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
	var ok bool
	status.Repl.IsMaster, ok = isMaster["ismaster"].(bool)
	if !ok {
		status.Repl.IsMaster = false
	}

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

func (d *Dao) ListDocuments(ctx context.Context, state *CollectionState) ([]primitive.M, int64, error) {
	filter, err := ParseStringQuery(state.Filter)
	if err != nil {
		return nil, 0, err
	}
	sort, err := ParseStringQuery(state.Sort)
	if err != nil {
		return nil, 0, err
	}

	count, err := d.client.Database(state.Db).Collection(state.Coll).CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	coll := d.client.Database(state.Db).Collection(state.Coll)

	options := options.FindOptions{
		Limit: &state.Limit,
		Skip:  &state.Page,
		Sort:  sort,
	}

	cursor, err := coll.Find(ctx, filter, &options)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var documents []primitive.M
	for cursor.Next(ctx) {
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

func (d *Dao) GetDocument(ctx context.Context, db string, collection string, id primitive.ObjectID) (primitive.M, error) {
	var document primitive.M
	err := d.client.Database(db).Collection(collection).FindOne(ctx, primitive.M{"_id": id}).Decode(&document)
	if err != nil {
		return nil, err
	}
	return document, nil
}

func (d *Dao) InsetDocument(ctx context.Context, db string, collection string, document primitive.M) (interface{}, error) {
	res, err := d.client.Database(db).Collection(collection).InsertOne(ctx, document)
	if err != nil {
		return nil, err
	}

	log.Debug().Msgf("Document inserted, document: %v, db: %v, collection: %v", document, db, collection)

	return res.InsertedID, nil
}

func (d *Dao) UpdateDocument(ctx context.Context, db string, collection string, id interface{}, document primitive.M) error {
	updated, err := d.client.Database(db).Collection(collection).UpdateOne(ctx, primitive.M{"_id": id}, primitive.M{"$set": document})
	if err != nil {
		log.Error().Msgf("Error updating document: %v", err)
		return err
	}

	if updated.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	log.Debug().Msgf("Document updated, id: %v, document: %v, db: %v, collection: %v", id, document, db, collection)

	return nil
}

func (d *Dao) DeleteDocument(ctx context.Context, db string, collection string, id interface{}) error {
	deleted, err := d.client.Database(db).Collection(collection).DeleteOne(ctx, primitive.M{"_id": id})
	if err != nil {
		return err
	}

	if deleted.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	log.Debug().Msgf("Document deleted, id: %v, db: %v, collection: %v", id, db, collection)

	return nil
}

func (d *Dao) AddCollection(ctx context.Context, db string, collection string) error {
	err := d.client.Database(db).CreateCollection(ctx, collection)
	if err != nil {
		return err
	}

	log.Debug().Msgf("Collection added, db: %v, collection: %v", db, collection)

	return nil
}

func (d *Dao) DeleteCollection(ctx context.Context, db string, collection string) error {
	err := d.client.Database(db).Collection(collection).Drop(ctx)
	if err != nil {
		return err
	}

	log.Debug().Msgf("Collection deleted, db: %v, collection: %v", db, collection)

	return nil
}

func (d *Dao) ForceClose(ctx context.Context) error {
	if err := d.client.Disconnect(ctx); err != nil {
		log.Error().Err(err).Msg("Error disconnecting from the database")
		return err
	}

	log.Debug().Msg("Connection closed")
	return nil
}

func (d *Dao) RunCommand(ctx context.Context, db string, command primitive.D) (primitive.M, error) {
	results := primitive.M{}

	err := d.client.Database(db).RunCommand(ctx, command).Decode(&results)
	if err != nil {
		return nil, err
	}

	log.Debug().Msgf("Command run, db: %v, command: %v, results: %v", db, command, results)

	return results, nil
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
