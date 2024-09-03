package mongo

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
