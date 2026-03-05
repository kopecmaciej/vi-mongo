package mongo

var (
	objectID = MongoKeyword{
		Display:     "ObjectID(\"\"))",
		InsertText:  "ObjectID(\"<$0>\")",
		Description: "ObjectID is a 12-byte BSON type",
	}
	bsonTypes = []MongoKeyword{
		{
			Display:     "ISODate(\"\")",
			InsertText:  "ISODate(\"<$0>\")",
			Description: "ISODate represents a date in ISO 8601 format",
		},
		{
			Display:     "NumberDecimal()",
			InsertText:  "NumberDecimal(<$0>)",
			Description: "NumberDecimal is a 128-bit decimal-based floating-point (Decimal128)",
		},
		{
			Display:     "NumberLong()",
			InsertText:  "NumberLong(<$0>)",
			Description: "NumberLong is a 64-bit integer",
		},
		{
			Display:     "NumberInt()",
			InsertText:  "NumberInt(<$0>)",
			Description: "NumberInt is a 32-bit integer",
		},
		{
			Display:     "Timestamp()",
			InsertText:  "Timestamp(<$0>)",
			Description: "Timestamp is a BSON timestamp type",
		},
		{
			Display:     "BinData()",
			InsertText:  "BinData(<$0>)",
			Description: "BinData is a BSON binary data type",
		},
		{
			Display:     "MinKey",
			InsertText:  "MinKey",
			Description: "MinKey compares lower than all other BSON types",
		},
		{
			Display:     "MaxKey",
			InsertText:  "MaxKey",
			Description: "MaxKey compares higher than all other BSON types",
		},
		{
			Display:     "RegExp(\"\")",
			InsertText:  "RegExp(\"<$0>\")",
			Description: "RegExp represents a regular expression",
		},
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
			InsertText:  "$not: ",
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
			InsertText:  "$regex: \"<$0>\"",
			Description: "Matches values that match a specified regular expression.",
		},
		{
			Display:     "$regexWithOptions",
			InsertText:  "$regex: \"<$0>\", $options: \"\"",
			Description: "Matches values that match a specified regular expression with options.",
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
	geospatialOperators = []MongoKeyword{
		{
			Display:     "$geoIntersects",
			InsertText:  "$geoIntersects: ",
			Description: "Selects geometries that intersect with a GeoJSON geometry.",
		},
		{
			Display:     "$geoWithin",
			InsertText:  "$geoWithin: ",
			Description: "Selects geometries within a bounding GeoJSON geometry.",
		},
		// Add similar entries for $near, $nearSphere, $box, $center, $centerSphere, $geometry, $maxDistance, $minDistance, $polygon
	}
	arrayOperators = []MongoKeyword{
		{
			Display:     "$all",
			InsertText:  "$all: [<$0>]",
			Description: "Matches arrays that contain all elements specified in the query.",
		},
		{
			Display:     "$elemMatch",
			InsertText:  "$elemMatch: {<$0>}",
			Description: "Matches documents that contain an array field with at least one element that matches all the specified query criteria.",
		},
		{
			Display:     "$size",
			InsertText:  "$size: ",
			Description: "Matches any array with the number of elements specified by the argument.",
		},
	}
	bitwiseOperators = []MongoKeyword{
		{
			Display:     "$bitsAllClear",
			InsertText:  "$bitsAllClear: ",
			Description: "Matches documents where all bits in the specified field are clear.",
		},
		{
			Display:     "$bitsAllSet",
			InsertText:  "$bitsAllSet: ",
			Description: "Matches documents where all bits in the specified field are set.",
		},
		{
			Display:     "$bitsAnyClear",
			InsertText:  "$bitsAnyClear: ",
			Description: "Matches documents where at least one bit in the specified field is clear.",
		},
		{
			Display:     "$bitsAnySet",
			InsertText:  "$bitsAnySet: ",
			Description: "Matches documents where at least one bit in the specified field is set.",
		},
	}
	projectionOperators = []MongoKeyword{
		{
			Display:     "$elemMatch",
			InsertText:  "$elemMatch: {<$0>}",
			Description: "Matches documents that contain an array field with at least one element that matches all the specified query criteria.",
		},
		{
			Display:     "$meta",
			InsertText:  "$meta: ",
			Description: "Matches documents that contain a field with the specified metadata.",
		},
		{
			Display:     "$slice",
			InsertText:  "$slice: ",
			Description: "Selects a portion of an array.",
		},
	}
	miscellaneousOperators = []MongoKeyword{
		{
			Display:     "$comment",
			InsertText:  "$comment: ",
			Description: "Adds a comment to the query.",
		},
		{
			Display:     "$rand",
			InsertText:  "$rand: ",
			Description: "Selects a random document from the collection.",
		},
		{
			Display:     "$natural",
			InsertText:  "$natural: ",
			Description: "Selects documents in natural order.",
		},
	}
	queryModifiers = []MongoKeyword{
		{
			Display:     "$comment",
			InsertText:  "$comment: ",
			Description: "Adds a comment to the query.",
		},
		{
			Display:     "$explain",
			InsertText:  "$explain: ",
			Description: "Returns information about the query plan.",
		},
		{
			Display:     "$hint",
			InsertText:  "$hint: ",
			Description: "Specifies the index to use for the query.",
		},
		{
			Display:     "$max",
			InsertText:  "$max: ",
			Description: "Specifies the maximum number of documents to return.",
		},
		{
			Display:     "$maxScan",
			InsertText:  "$maxScan: ",
			Description: "Specifies the maximum number of documents to scan.",
		},
		{
			Display:     "$maxTimeMS",
			InsertText:  "$maxTimeMS: ",
			Description: "Specifies the maximum time in milliseconds to spend on the query.",
		},
		{
			Display:     "$min",
			InsertText:  "$min: ",
			Description: "Specifies the minimum number of documents to return.",
		},
		{
			Display:     "$orderby",
			InsertText:  "$orderby: ",
			Description: "Specifies the order in which to return documents.",
		},
		{
			Display:     "$returnKey",
			InsertText:  "$returnKey: ",
			Description: "Returns only the specified fields.",
		},
		{
			Display:     "$showDiskLoc",
			InsertText:  "$showDiskLoc: ",
			Description: "Returns the location of the documents on disk.",
		},
		{
			Display:     "$snapshot",
			InsertText:  "$snapshot: ",
			Description: "Returns a snapshot of the collection at the time the query is executed.",
		},
		{
			Display:     "$natural",
			InsertText:  "$natural: ",
			Description: "Selects documents in natural order.",
		},
	}
	aggregationPipelineOperators = []MongoKeyword{
		{
			Display:     "$addFields",
			InsertText:  "$addFields: { <$0> }",
			Description: "Adds new fields to the documents in the pipeline.",
		},
		{
			Display:     "$bucket",
			InsertText:  "$bucket: { groupBy: \"$<$0>\", boundaries: [], default: \"Other\", output: {} }",
			Description: "Groups documents into buckets based on a specified expression.",
		},
		{
			Display:     "$bucketAuto",
			InsertText:  "$bucketAuto: { groupBy: \"$<$0>\", buckets: 5 }",
			Description: "Groups documents into a fixed number of buckets automatically.",
		},
		{
			Display:     "$collStats",
			InsertText:  "$collStats: { latencyStats: {}, storageStats: {}, count: {} }",
			Description: "Returns statistics about the collection.",
		},
		{
			Display:     "$count",
			InsertText:  "$count: \"<$0>\"",
			Description: "Counts documents and stores the result in a field.",
		},
		{
			Display:     "$facet",
			InsertText:  "$facet: { <$0>: [] }",
			Description: "Processes multiple aggregation pipelines within a single stage.",
		},
		{
			Display:     "$geoNear",
			InsertText:  "$geoNear: { near: { type: \"Point\", coordinates: [<$0>] }, distanceField: \"\", spherical: true }",
			Description: "Finds the nearest documents to a specified point.",
		},
		{
			Display:     "$graphLookup",
			InsertText:  "$graphLookup: { from: \"<$0>\", startWith: \"$\", connectFromField: \"\", connectToField: \"\", as: \"\" }",
			Description: "Performs a recursive graph lookup on a collection.",
		},
		{
			Display:     "$group",
			InsertText:  "$group: { _id: \"$<$0>\" }",
			Description: "Groups documents by a specified expression.",
		},
		{
			Display:     "$indexStats",
			InsertText:  "$indexStats: {}",
			Description: "Returns statistics about the indexes on the collection.",
		},
		{
			Display:     "$limit",
			InsertText:  "$limit: <$0>",
			Description: "Limits the number of documents returned.",
		},
		{
			Display:     "$listSessions",
			InsertText:  "$listSessions: {}",
			Description: "Returns a list of all sessions in the database.",
		},
		{
			Display:     "$listLocalSessions",
			InsertText:  "$listLocalSessions: {}",
			Description: "Returns a list of all local sessions in the database.",
		},
		{
			Display:     "$lookup",
			InsertText:  "$lookup: { from: \"<$0>\", localField: \"\", foreignField: \"\", as: \"\" }",
			Description: "Joins two collections on a local and foreign field.",
		},
		{
			Display:     "$lookupPipeline",
			InsertText:  "$lookup: { from: \"<$0>\", let: {}, pipeline: [], as: \"\" }",
			Description: "Joins a collection using a sub-pipeline (uncorrelated or let-based).",
		},
		{
			Display:     "$match",
			InsertText:  "$match: { <$0> }",
			Description: "Filters documents based on a specified expression.",
		},
		{
			Display:     "$merge",
			InsertText:  "$merge: { into: \"<$0>\", on: \"_id\", whenMatched: \"merge\", whenNotMatched: \"insert\" }",
			Description: "Merges pipeline output into a collection.",
		},
		{
			Display:     "$out",
			InsertText:  "$out: \"<$0>\"",
			Description: "Writes the output of the pipeline to a specified collection.",
		},
		{
			Display:     "$planCacheStats",
			InsertText:  "$planCacheStats: {}",
			Description: "Returns statistics about the plan cache.",
		},
		{
			Display:     "$project",
			InsertText:  "$project: { <$0>: 1 }",
			Description: "Selects or computes fields to include in the output.",
		},
		{
			Display:     "$redact",
			InsertText:  "$redact: { $cond: { if: { <$0> }, then: \"$$DESCEND\", else: \"$$PRUNE\" } }",
			Description: "Restricts document content based on a condition.",
		},
		{
			Display:     "$replaceRoot",
			InsertText:  "$replaceRoot: { newRoot: \"$<$0>\" }",
			Description: "Replaces the root document with a specified embedded document.",
		},
		{
			Display:     "$replaceWith",
			InsertText:  "$replaceWith: \"$<$0>\"",
			Description: "Replaces the root document with a specified expression.",
		},
		{
			Display:     "$sample",
			InsertText:  "$sample: { size: <$0> }",
			Description: "Randomly selects a specified number of documents from the collection.",
		},
		{
			Display:     "$set",
			InsertText:  "$set: { <$0> }",
			Description: "Sets the value of a specified field in the output.",
		},
		{
			Display:     "$skip",
			InsertText:  "$skip: <$0>",
			Description: "Skips a specified number of documents in the output.",
		},
		{
			Display:     "$sort",
			InsertText:  "$sort: { <$0>: 1 }",
			Description: "Sorts the output. Use 1 for ascending, -1 for descending.",
		},
		{
			Display:     "$sortByCount",
			InsertText:  "$sortByCount: \"$<$0>\"",
			Description: "Groups by a field and sorts by the count of documents in each group.",
		},
		{
			Display:     "$unset",
			InsertText:  "$unset: \"<$0>\"",
			Description: "Removes a specified field from the output.",
		},
		{
			Display:     "$unwind",
			InsertText:  "$unwind: { path: \"$<$0>\", preserveNullAndEmptyArrays: false }",
			Description: "Deconstructs an array field into separate documents.",
		},
	}
	updateOperators = []MongoKeyword{
		{
			Display:     "$addToSet",
			InsertText:  "$addToSet: {<$0>}",
			Description: "Adds all elements of a specified array to a set in the document.",
		},
		{
			Display:     "$currentDate",
			InsertText:  "$currentDate: ",
			Description: "Sets the value of a specified field to the current date.",
		},
		{
			Display:     "$inc",
			InsertText:  "$inc: {<$0>}",
			Description: "Increments the value of a specified field by a specified amount.",
		},
		{
			Display:     "$min",
			InsertText:  "$min: {<$0>}",
			Description: "Sets the value of a specified field to the minimum of its current value and a specified value.",
		},
		{
			Display:     "$max",
			InsertText:  "$max: {<$0>}",
			Description: "Sets the value of a specified field to the maximum of its current value and a specified value.",
		},
		{
			Display:     "$mul",
			InsertText:  "$mul: {<$0>}",
			Description: "Multiplies the value of a specified field by a specified amount.",
		},
		{
			Display:     "$pop",
			InsertText:  "$pop: {<$0>}",
			Description: "Removes the first or last element from an array in the document.",
		},
		{
			Display:     "$pull",
			InsertText:  "$pull: {<$0>}",
			Description: "Removes all occurrences of a specified value from an array in the document.",
		},
		{
			Display:     "$push",
			InsertText:  "$push: {<$0>}",
			Description: "Adds a specified value to an array in the document.",
		},
		{
			Display:     "$rename",
			InsertText:  "$rename: {<$0>}",
			Description: "Renames a specified field in the document.",
		},
		{
			Display:     "$set",
			InsertText:  "$set: {<$0>}",
			Description: "Sets the value of a specified field in the document.",
		},
		{
			Display:     "$setOnInsert",
			InsertText:  "$setOnInsert: {<$0>}",
			Description: "Sets the value of a specified field in the document only when the document is inserted.",
		},
		{
			Display:     "$unset",
			InsertText:  "$unset: {<$0>}",
			Description: "Removes a specified field from the document.",
		},
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

// GetAggregationPipelineOperators returns the list of aggregation pipeline stage operators.
func GetAggregationPipelineOperators() []MongoKeyword {
	return aggregationPipelineOperators
}

// getMongoOperators returns list of all mongo operators
func getMongoOperators() []MongoKeyword {
	operators := []MongoKeyword{}

	operators = append(operators, objectID)
	operators = append(operators, bsonTypes...)
	operators = append(operators, comparisonOperators...)
	operators = append(operators, elementOperators...)
	operators = append(operators, evaluationOperators...)
	operators = append(operators, logicalOperators...)
	operators = append(operators, geospatialOperators...)
	operators = append(operators, arrayOperators...)
	operators = append(operators, bitwiseOperators...)
	operators = append(operators, projectionOperators...)
	operators = append(operators, miscellaneousOperators...)
	operators = append(operators, queryModifiers...)
	operators = append(operators, updateOperators...)

	return operators
}
