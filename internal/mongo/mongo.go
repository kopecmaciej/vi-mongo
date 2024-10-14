package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ServerStatus are the values chosen from the serverStatus command
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

// DBsWithCollections is a object used to store the database name and its collections
type DBsWithCollections struct {
	DB          string
	Collections []string
}

// IndexInfo represents the combined information about an index from multiple commands
type IndexInfo struct {
	Name       string
	Definition bson.M
	Type       string
	Size       string
	Usage      string
	Properties []string
}

// indexStats represents the data that is returned by the indexStats command
type indexStats struct {
	Size     string
	Accesses primitive.M
}
