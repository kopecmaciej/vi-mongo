package mongo

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strconv"
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
	err := d.client.Ping(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to ping MongoDB")
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}
	return nil
}

func (d *Dao) GetServerStatus(ctx context.Context) (*ServerStatus, error) {
	var status ServerStatus
	err := d.client.Database("admin").RunCommand(ctx, primitive.D{{Key: "serverStatus", Value: 1}}).Decode(&status)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get server status")
		return nil, fmt.Errorf("failed to get server status: %w", err)
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

func (d *Dao) ListDbsWithCollections(ctx context.Context, nameRegex string) ([]DBsWithCollections, error) {
	dbCollMap := []DBsWithCollections{}

	filter := primitive.M{}
	if nameRegex != "" {
		filter = primitive.M{"name": primitive.Regex{Pattern: nameRegex, Options: "i"}}
	}

	listDbOptions := options.ListDatabases().SetAuthorizedDatabases(*d.Config.GetOptions().AuthorizedDatabases)
	dbNames, err := d.client.ListDatabaseNames(ctx, filter, listDbOptions)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list databases")
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}

	for _, dbName := range dbNames {
		listCollOptions := options.ListCollections().SetAuthorizedCollections(*d.Config.GetOptions().AuthorizedCollections)

		collNames, err := d.client.Database(dbName).ListCollectionNames(ctx, primitive.M{}, listCollOptions)
		if err != nil {
			log.Error().Err(err).Str("database", dbName).Msg("Failed to list collections")
			continue
		}

		slices.Sort(collNames)

		dbCollMap = append(dbCollMap, DBsWithCollections{DB: dbName, Collections: collNames})
	}

	return dbCollMap, nil
}

func (d *Dao) ListDocuments(ctx context.Context, state *CollectionState, filter primitive.M, sort primitive.M,
	projection primitive.M, countCallback func(int64)) ([]primitive.M, error) {

	coll := d.client.Database(state.Db).Collection(state.Coll)

	options := options.FindOptions{
		Limit:      &state.Limit,
		Skip:       &state.Skip,
		Sort:       sort,
		Projection: projection,
	}

	cursor, err := coll.Find(ctx, filter, &options)
	if err != nil {
		log.Error().Err(err).Str("db", state.Db).Str("collection", state.Coll).Msg("Failed to find documents")
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to close cursor")
		}
	}()

	var documents []primitive.M
	err = cursor.All(ctx, &documents)
	if err != nil {
		log.Error().Err(err).Str("db", state.Db).Str("collection", state.Coll).Msg("Failed to decode documents")
		return nil, fmt.Errorf("failed to decode documents: %w", err)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	go func() {
		count, err := coll.CountDocuments(ctx, filter)
		if err != nil {
			log.Error().Err(err).Str("db", state.Db).Str("collection", state.Coll).Msg("Failed to count documents")
			return
		}

		if countCallback != nil {
			countCallback(count)
		}
	}()

	return documents, nil
}

func (d *Dao) GetDocument(ctx context.Context, db string, coll string, id primitive.ObjectID) (primitive.M, error) {
	var document primitive.M
	err := d.client.Database(db).Collection(coll).FindOne(ctx, primitive.M{"_id": id}).Decode(&document)
	if err != nil {
		log.Error().Err(err).Str("db", db).Str("collection", coll).Str("id", id.Hex()).Msg("Failed to get document")
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	return document, nil
}

func (d *Dao) InsetDocument(ctx context.Context, db string, coll string, document primitive.M) (any, error) {
	res, err := d.client.Database(db).Collection(coll).InsertOne(ctx, document)
	if err != nil {
		log.Error().Err(err).Str("db", db).Str("collection", coll).Msg("Failed to insert document")
		return nil, fmt.Errorf("failed to insert document: %w", err)
	}

	log.Debug().Msgf("Document inserted, document: %v, db: %v, collection: %v", document, db, coll)

	return res.InsertedID, nil
}

func (d *Dao) UpdateDocument(ctx context.Context, db string, coll string, id any, originalDoc, document primitive.M) error {
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

	updated, err := d.client.Database(db).Collection(coll).UpdateOne(ctx, primitive.M{"_id": id}, update)
	if err != nil {
		log.Error().Err(err).Str("db", db).Str("collection", coll).Interface("id", id).Msg("Failed to update document")
		return fmt.Errorf("failed to update document: %w", err)
	}

	if updated.MatchedCount == 0 {
		log.Error().Str("db", db).Str("collection", coll).Interface("id", id).Msg("No document found to update")
		return mongo.ErrNoDocuments
	}

	log.Debug().Msgf("Document updated, id: %v, document: %v, db: %v, collection: %v", id, document, db, coll)

	return nil
}

func (d *Dao) DeleteDocument(ctx context.Context, db string, coll string, id any) error {
	deleted, err := d.client.Database(db).Collection(coll).DeleteOne(ctx, primitive.M{"_id": id})
	if err != nil {
		log.Error().Err(err).Str("db", db).Str("collection", coll).Interface("id", id).Msg("Failed to delete document")
		return fmt.Errorf("failed to delete document: %w", err)
	}

	if deleted.DeletedCount == 0 {
		log.Error().Str("db", db).Str("collection", coll).Interface("id", id).Msg("No document found to delete")
		return mongo.ErrNoDocuments
	}

	log.Debug().Msgf("Document deleted, id: %v, db: %v, collection: %v", id, db, coll)

	return nil
}

func (d *Dao) AddCollection(ctx context.Context, db string, coll string) error {
	err := d.client.Database(db).CreateCollection(ctx, coll)
	if err != nil {
		log.Error().Err(err).Str("db", db).Str("collection", coll).Msg("Failed to add collection")
		return err
	}

	log.Debug().Msgf("Collection added, db: %v, collection: %v", db, coll)

	return nil
}

func (d *Dao) DeleteCollection(ctx context.Context, db string, coll string) error {
	err := d.client.Database(db).Collection(coll).Drop(ctx)
	if err != nil {
		log.Error().Err(err).Str("db", db).Str("collection", coll).Msg("Failed to delete collection")
		return err
	}

	log.Debug().Msgf("Collection deleted, db: %v, collection: %v", db, coll)

	return nil
}

func (d *Dao) RenameCollection(ctx context.Context, db string, oldColl string, newColl string) error {
	renameCmd := bson.D{
		{Key: "renameCollection", Value: fmt.Sprintf("%s.%s", db, oldColl)},
		{Key: "to", Value: fmt.Sprintf("%s.%s", db, newColl)},
	}

	err := d.client.Database("admin").RunCommand(ctx, renameCmd).Err()
	if err != nil {
		log.Error().Err(err).Msg("Failed to rename collection")
		return err
	}

	log.Debug().Msgf("Collection renamed, db: %v, old collection: %v, new collection: %v", db, oldColl, newColl)

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

func (d *Dao) runAdminCommand(ctx context.Context, key string, value any) (primitive.M, error) {
	results := primitive.M{}
	command := primitive.D{{Key: key, Value: value}}

	err := d.client.Database("admin").RunCommand(ctx, command).Decode(&results)
	if err != nil {
		log.Error().Err(err).Str("key", key).Interface("value", value).Msg("Failed to run addmin command")
		return nil, err
	}

	return results, nil
}

// GetIndexes fetches the indexes for a given database and collection
func (d *Dao) GetIndexes(ctx context.Context, db string, coll string) ([]IndexInfo, error) {
	collHandle := d.client.Database(db).Collection(coll)
	cursor, err := collHandle.Indexes().List(ctx)
	if err != nil {
		log.Error().Err(err).Str("db", db).Str("collection", coll).Msg("Error fetching indexes")
		return nil, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to close cursor")
		}
	}()

	var indexes []IndexInfo
	for cursor.Next(ctx) {
		var idx bson.M
		if err := cursor.Decode(&idx); err != nil {
			log.Error().Err(err).Str("db", db).Str("collection", coll).Msg("Error unmarshalling indexes")
			return nil, err
		}

		indexInfo := IndexInfo{
			Name:       idx["name"].(string),
			Definition: idx["key"].(bson.M),
			Type:       "REGULAR",
			Size:       "N/A",
			Usage:      "N/A",
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
		log.Error().Err(err).Str("db", db).Str("collection", coll).Msg("Error unmarshalling indexes")
		return nil, err
	}

	// Fetch index sizes and usage statistics
	stats, err := d.getIndexStats(ctx, db, coll)
	if err != nil {
		log.Error().Err(err).Str("db", db).Str("collection", coll).Msg("Error fetching index statistics")
	} else {
		for i, idx := range indexes {
			if stat, ok := stats[idx.Name]; ok {
				indexes[i].Size = stat.Size
				indexes[i].Usage = formatIndexUsage(stat.Accesses["ops"].(int64), stat.Accesses["since"].(primitive.DateTime).Time())
			}
		}
	}

	return indexes, nil
}

func (d *Dao) getIndexStats(ctx context.Context, db string, collection string) (map[string]indexStats, error) {
	cursor, err := d.client.Database(db).Collection(collection).Aggregate(ctx, mongo.Pipeline{
		bson.D{{Key: "$indexStats", Value: bson.D{}}},
	})
	if err != nil {
		log.Error().Err(err).Str("db", db).Str("collection", collection).Msg("Error getting indexes")
		return nil, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to close cursor")
		}
	}()

	var collStats bson.M
	err = d.client.Database(db).RunCommand(ctx, bson.D{{Key: "collStats", Value: collection}}).Decode(&collStats)
	if err != nil {
		log.Error().Err(err).Str("db", db).Str("collection", collection).Msg("Error while running command collStats")
		return nil, err
	}

	sizesMap := collStats["indexSizes"].(bson.M)

	var stats []bson.M
	if err := cursor.All(ctx, &stats); err != nil {
		log.Error().Err(err).Str("db", db).Str("collection", collection).Msg("Error decoding stats")
		return nil, err
	}

	statsMap := make(map[string]indexStats)
	for _, stat := range stats {
		size, ok := sizesMap[stat["name"].(string)]
		if !ok {
			size = "N/A"
		}
		sizeNum, err := strconv.ParseInt(fmt.Sprintf("%d", size), 10, 64)
		if err != nil {
			sizeNum = 0
		}
		statsMap[stat["name"].(string)] = indexStats{Size: fmt.Sprintf("%.1f KB", float64(sizeNum)/1024), Accesses: stat["accesses"].(primitive.M)}
	}

	return statsMap, nil
}
func formatIndexUsage(ops int64, since time.Time) string {
	return fmt.Sprintf("%d (since %s)", ops, since.Format("2006-01-02 15:04:05"))
}

func (d *Dao) CreateIndex(ctx context.Context, db, coll string, indexDef mongo.IndexModel) error {
	_, err := d.client.Database(db).Collection(coll).Indexes().CreateOne(ctx, indexDef)
	if err != nil {
		log.Error().Err(err).Str("db", db).Str("collection", coll).Msg("Error creating index")
		return err
	}
	return nil
}

func (d *Dao) DropIndex(ctx context.Context, db, coll, indexName string) error {
	_, err := d.client.Database(db).Collection(coll).Indexes().DropOne(ctx, indexName)
	if err != nil {
		log.Error().Err(err).Str("db", db).Str("collection", coll).Msg("Error droping index")
		return err
	}
	return nil
}
