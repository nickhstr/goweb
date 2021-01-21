package mongodb

import (
	"context"
)

// Aggregate uses an aggregation pipeline to process
// data records and decode into a results data structure.
// Results must be a pointer to a slice.
func (db *DB) Aggregate(ctx context.Context, collName string, pipeline interface{}, results interface{}) error {
	coll := db.Collection(collName)

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}

	return cursor.All(ctx, results)
}
