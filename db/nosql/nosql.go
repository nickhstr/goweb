package nosql

import (
	"context"
)

type Aggregator interface {
	// Aggregate uses an aggregation pipeline to process
	// data records and amd decode into a results data structure.
	// Results must be a pointer to a slice.
	Aggregate(ctx context.Context, collName string, pipeline interface{}, results interface{}) error
}

type Deleter interface {
	// DeleteOne attempts to delete a document by ID.
	DeleteOne(ctx context.Context, collName, id string) error
}

type Finder interface {
	// FindByID queries by the given ID for a single document,
	// decoding the found BSON document into result.
	// Note, result must be a pointer.
	FindByID(ctx context.Context, collName, id string, result interface{}) error

	// FindOne queries by the given filter for a single document,
	// decoding the found BSON document into result.
	// Note, result must be a pointer.
	FindOne(ctx context.Context, collName string, filter interface{}, result interface{}) error

	// FindAll queries for multiple documents, decoding found
	// documents into results.
	// A start index and limit are required. Start must be at least zero,
	// and less than the number of documents in the collection;
	// limit can be zero (to get all documents), or some integer
	// less than or equal to the document count.
	FindAll(ctx context.Context, collName string, start, limit int64, results interface{}) error

	// FindMany queries for multiple documents, using the
	// given filter, and decodes found documents into results.
	// A start index and limit are required. Start must be at least zero,
	// and less than the number of matched documents in the collection;
	// limit can be zero (to get all matched documents), or some integer
	// less than or equal to the matched document count.
	// Note, results must be a pointer to a slice.
	FindMany(ctx context.Context, collName string, start, limit int64, filter interface{}, results interface{}) error
}

type Inserter interface {
	// InsertByID tries to insert the given data into the collection.
	// If data is already stored under the same ID, an error is returned.
	InsertByID(ctx context.Context, collName, id string, data interface{}) error
}

type Replacer interface {
	// ReplaceByID updates a collection's document by replacing it.
	ReplaceByID(ctx context.Context, collName, id string, data interface{}) error
}

// Storer abstracts common operations used by nosql databases.
type Storer interface {
	Aggregator
	Deleter
	Finder
	Inserter
	Replacer
}
