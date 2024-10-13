package mongo

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/kopecmaciej/vi-mongo/internal/config"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
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

func (d *Dao) Ping(ctx context.Context) error {
	return d.client.Ping(ctx, nil)
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

func (d *Dao) ListDocuments(ctx context.Context, state *CollectionState, filter primitive.M, sort primitive.M) ([]primitive.M, int64, error) {
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

func (d *Dao) UpdateDocument(ctx context.Context, db string, collection string, id interface{}, originalDoc, document primitive.M) error {
	setOps := bson.M{}
	unsetOps := bson.M{}

	for key, value := range document {
		if origValue, exists := originalDoc[key]; !exists || !reflect.DeepEqual(origValue, value) {
			setOps[key] = value
		}
	}

	for key := range originalDoc {
		if _, exists := document[key]; !exists {
			unsetOps[key] = 1
		}
	}

	update := bson.M{}
	if len(setOps) > 0 {
		update["$set"] = setOps
	}
	if len(unsetOps) > 0 {
		update["$unset"] = unsetOps
	}

	if len(update) == 0 {
		return nil
	}

	updated, err := d.client.Database(db).Collection(collection).UpdateOne(ctx, primitive.M{"_id": id}, update)
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

func (d *Dao) runAdminCommand(ctx context.Context, key string, value interface{}) (primitive.M, error) {
	results := primitive.M{}
	command := primitive.D{{Key: key, Value: value}}

	err := d.client.Database("admin").RunCommand(ctx, command).Decode(&results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// IndexInfo represents the information about an index
type IndexInfo struct {
	Name       string
	Definition bson.M
	Type       string
	Size       int64
	Usage      string
	Properties []string
}

// GetIndexes fetches the indexes for a given database and collection
func (d *Dao) GetIndexes(ctx context.Context, db string, collection string) ([]IndexInfo, error) {
	coll := d.client.Database(db).Collection(collection)
	cursor, err := coll.Indexes().List(ctx)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var indexes []IndexInfo
	for cursor.Next(ctx) {
		var idx bson.M
		if err := cursor.Decode(&idx); err != nil {
			return nil, err
		}

		indexInfo := IndexInfo{
			Name:       idx["name"].(string),
			Definition: idx["key"].(bson.M),
			Type:       "REGULAR",                                                // Default type
			Size:       0,                                                        // We'll need to calculate this separately
			Usage:      "0 (since " + time.Now().Format("Mon Jan 02 2006") + ")", // Default usage
			Properties: []string{},
		}

		// Determine index type and properties
		if unique, ok := idx["unique"]; ok && unique.(bool) {
			indexInfo.Properties = append(indexInfo.Properties, "UNIQUE")
		}
		if sparse, ok := idx["sparse"]; ok && sparse.(bool) {
			indexInfo.Properties = append(indexInfo.Properties, "SPARSE")
		}
		if ttl, ok := idx["expireAfterSeconds"]; ok && ttl.(int32) > 0 {
			indexInfo.Properties = append(indexInfo.Properties, "TTL")
			indexInfo.Type = "TTL"
		}
		if len(indexInfo.Definition) > 1 {
			indexInfo.Properties = append(indexInfo.Properties, "COMPOUND")
		}
		if indexInfo.Name == "_id_" {
			indexInfo.Properties = append(indexInfo.Properties, "UNIQUE")
		}

		indexes = append(indexes, indexInfo)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	// Fetch index sizes and usage statistics
	stats, err := d.getIndexStats(ctx, db, collection)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to fetch index statistics")
	} else {
		for i, idx := range indexes {
			if stat, ok := stats[idx.Name]; ok {
				indexes[i].Size = stat.Size
				indexes[i].Usage = formatIndexUsage(stat.Accesses.Ops, stat.Accesses.Since)
			}
		}
	}

	return indexes, nil
}

type indexStats struct {
	Size     int64
	Accesses struct {
		Ops   int64
		Since time.Time
	}
}

func (d *Dao) getIndexStats(ctx context.Context, db string, collection string) (map[string]indexStats, error) {
	var result bson.M
	err := d.client.Database(db).RunCommand(ctx, bson.D{{Key: "collStats", Value: collection}}).Decode(&result)
	if err != nil {
		return nil, err
	}

	indexSizes, ok := result["indexSizes"].(bson.M)
	if !ok {
		return nil, fmt.Errorf("indexSizes not found in collStats result")
	}

	indexAccesses, ok := result["indexAccesses"].(bson.M)
	if !ok {
		log.Warn().Msg("indexAccesses not found in collStats result")
	}

	stats := make(map[string]indexStats)
	for name, size := range indexSizes {
		stat := indexStats{Size: int64(size.(int32))}
		if accesses, ok := indexAccesses[name].(bson.M); ok {
			stat.Accesses.Ops = accesses["ops"].(int64)
			stat.Accesses.Since, _ = accesses["since"].(time.Time)
		}
		stats[name] = stat
	}

	return stats, nil
}

func formatIndexUsage(ops int64, since time.Time) string {
	return fmt.Sprintf("%d (since %s)", ops, since.Format("Mon Jan 02 2006"))
}
