package mongo

var (
	objectID = MongoKeyword{
		Display:     "ObjectID(\"\"))",
		InsertText:  "ObjectID(\"<$0>\")",
		Description: "ObjectID is a 12-byte BSON type",
	}
	comparisonOperators = []MongoKeyword{
		{
			Display:     "$eq",
			InsertText:  "$eq: ",
			Description: "Matches values that are equal to a specified value.",
		},
		{
			Display:     "$gt",
			InsertText:  "$gt: ",
			Description: "Matches values that are greater than a specified value.",
		},
		{
			Display:     "$gte",
			InsertText:  "$gte: ",
			Description: "Matches values that are greater than or equal to a specified value.",
		},
		{
			Display:     "$in",
			InsertText:  "$in: [<$0>]",
			Description: "Matches any of the values specified in an array.",
		},
		{
			Display:     "$lt",
			InsertText:  "$lt: ",
			Description: "Matches values that are less than a specified value.",
		},
		{
			Display:     "$lte",
			InsertText:  "$lte: ",
			Description: "Matches values that are less than or equal to a specified value.",
		},
		{
			Display:     "$ne",
			InsertText:  "$ne: ",
			Description: "Matches all values that are not equal to a specified value.",
		},
		{
			Display:     "$nin",
			InsertText:  "$nin: [<$0>]",
			Description: "Matches none of the values specified in an array.",
		},
	}
	logicalOperators = []MongoKeyword{
		{
			Display:     "$and",
			InsertText:  "$and: [<$0>]",
			Description: "Performs a logical AND operation on an array of two or more expressions.",
		},
		{
			Display:     "$not",
			InsertText:  "$not: {<$0>}",
			Description: "Inverts the effect of a query expression.",
		},
		{
			Display:     "$nor",
			InsertText:  "$nor: [<$0>]",
			Description: "Performs a logical NOR operation on an array of two or more expressions.",
		},
		{
			Display:     "$or",
			InsertText:  "$or: [<$0>]",
			Description: "Performs a logical OR operation on an array of two or more expressions.",
		},
	}
	elementOperators = []MongoKeyword{
		{
			Display:     "$exists",
			InsertText:  "$exists: ",
			Description: "Matches documents that have the specified field.",
		},
		{
			Display:     "$type",
			InsertText:  "$type: ",
			Description: "Selects documents if a field is of the specified type.",
		},
	}
	evaluationOperators = []MongoKeyword{
		{
			Display:     "$expr",
			InsertText:  "$expr: {<$0>}",
			Description: "Allows the use of aggregation expressions within the query language.",
		},
		{
			Display:     "$jsonSchema",
			InsertText:  "$jsonSchema: ",
			Description: "Validate documents against the given JSON Schema.",
		},
		{
			Display:     "$mod",
			InsertText:  "$mod: [<$0>]",
			Description: "Performs a modulo operation on the value of a field and selects documents with a specified result.",
		},
		{
			Display:     "$regex",
			InsertText:  "$regex: ",
			Description: "Selects documents where values match a specified regular expression.",
		},
		{
			Display:     "$text",
			InsertText:  "$text: {<$0>}",
			Description: "Performs text search.",
		},
		{
			Display:     "$where",
			InsertText:  "$where: ",
			Description: "Matches documents that satisfy a JavaScript expression.",
		},
	}
	geospatialOperators = []string{
		"$geoIntersects", "$geoWithin", "$near", "$nearSphere", "$box", "$center",
		"$centerSphere", "$geometry", "$maxDistance", "$minDistance", "$polygon",
	}
	arrayOperators = []string{
		"$all", "$elemMatch", "$size",
	}
	bitwiseOperators = []string{
		"$bitsAllClear", "$bitsAllSet", "$bitsAnyClear", "$bitsAnySet",
	}
	projectionOperators = []string{
		"$elemMatch", "$meta", "$slice",
	}
	miscellaneousOperators = []string{
		"$comment", "$rand", "$natural",
	}
	queryModifiers = []string{
		"$comment", "$explain", "$hint", "$max", "$maxScan", "$maxTimeMS",
		"$min", "$orderby", "$returnKey", "$showDiskLoc", "$snapshot", "$natural",
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
)

// MongoKeyword represents single mongo keyword
// Display is displayed in autocomplete
// InsertText is inserted into input, use <$i> marker to mark position
// for cursor, if empty then cursor moves to the end of the text
// Description is displayed in autocomplete description
type MongoKeyword struct {
	Display     string
	InsertText  string
	Description string
}

type MongoAutocomplete struct {
	Operators []MongoKeyword
}

func NewMongoAutocomplete() *MongoAutocomplete {
	return &MongoAutocomplete{
		Operators: getMongoOperators(),
	}
}

func (m *MongoAutocomplete) GetOperatorByDisplay(display string) *MongoKeyword {
	for _, operator := range m.Operators {
		if operator.Display == display {
			return &operator
		}
	}

	return nil
}

// getMongoOperators returns list of all mongo operators
func getMongoOperators() []MongoKeyword {
	operators := []MongoKeyword{}

	operators = append(operators, objectID)
	operators = append(operators, comparisonOperators...)
	operators = append(operators, elementOperators...)
	operators = append(operators, evaluationOperators...)
	operators = append(operators, logicalOperators...)

	return operators
}
