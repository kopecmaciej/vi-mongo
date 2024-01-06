package mongo

var (
	comparisonOperators = []string{
		"$eq", "$gt", "$gte", "$in", "$lt", "$lte", "$ne", "$nin",
	}
	logicalOperators = []string{
		"$and", "$not", "$nor", "$or",
	}
	elementOperators = []string{
		"$exists", "$type",
	}
	evaluationOperators = []string{
		"$expr", "$jsonSchema", "$mod", "$regex", "$text", "$where",
	}
	arrayOperators = []string{
		"$all", "$elemMatch", "$size",
	}
	projectionOperators = []string{
		"$elemMatch", "$meta", "$slice",
	}
	queryModifiers = []string{
		"$comment", "$explain", "$hint", "$max", "$maxScan", "$maxTimeMS",
		"$min", "$orderby", "$returnKey", "$showDiskLoc", "$snapshot", "$natural",
	}
	geospatialOperators = []string{
		"$geoIntersects", "$geoWithin", "$near", "$nearSphere",
	}
	aggregationPipelineOperators = []string{
		"$addFields", "$bucket", "$bucketAuto", "$collStats", "$count",
		"$facet", "$geoNear", "$graphLookup", "$group", "$indexStats",
		"$limit", "$listSessions", "$listLocalSessions", "$lookup", "$match",
		"$merge", "$out", "$planCacheStats", "$project", "$redact",
		"$replaceRoot", "$replaceWith", "$sample", "$set", "$skip",
		"$sort", "$sortByCount", "$unset", "$unwind",
	}
	updateOperators = []string{
		"$addToSet", "$currentDate", "$inc", "$min", "$max", "$mul",
		"$pop", "$pull", "$push", "$rename", "$set", "$setOnInsert", "$unset",
	}
	bitwiseOperators = []string{
		"$bit",
	}

	objectID = MongoKeyword{
		Display:     "ObjectId",
		InsertText:  "ObjectId(\"<>\")",
		Description: "ObjectId is a 12-byte BSON type",
	}
)

// MongoKeyword represents single mongo keyword
// Display is displayed in autocomplete
// InsertText is inserted into input, if empty then Display is used
// Description is displayed in autocomplete description
type MongoKeyword struct {
	Display     string
	InsertText  string
	Description string
}

type MongoAutocomplete struct {
	Operators []string
	ObjectID  MongoKeyword
}

func NewMongoAutocomplete() *MongoAutocomplete {
	return &MongoAutocomplete{
		Operators: getMongoOperators(),
		ObjectID:  objectID,
	}
}

// getMongoOperators returns list of all mongo operators
func getMongoOperators() []string {
	operators := []string{}

	operators = append(operators, comparisonOperators...)
	operators = append(operators, logicalOperators...)
	operators = append(operators, elementOperators...)
	operators = append(operators, evaluationOperators...)
	operators = append(operators, arrayOperators...)
	operators = append(operators, projectionOperators...)
	operators = append(operators, queryModifiers...)
	operators = append(operators, geospatialOperators...)
	operators = append(operators, aggregationPipelineOperators...)
	operators = append(operators, updateOperators...)
	operators = append(operators, bitwiseOperators...)

	return operators
}
