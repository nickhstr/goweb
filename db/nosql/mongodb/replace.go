package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo/options"
)

// ReplaceByID updates an collection's document by replacing it.
func (db *DB) ReplaceByID(ctx context.Context, collName, id string, data interface{}) error {
	var err error

	log := log.With().
		Str("collection", collName).
		Str("id", id).
		Logger()

	coll := db.Collection(collName)
	filter := db.IDFilter(id)

	// Set upsert to true to insert the data if it does not already exist.
	opts := options.Replace().SetUpsert(true)

	_, err = coll.ReplaceOne(ctx, filter, data, opts)
	if err != nil {
		log.Err(err).Msg("ReplaceOne failed")
		return err
	}

	// we can the error here since we know the filter is valid bson
	cacheKey, _ := db.CacheKey(collName, id)
	_ = db.cacher.Del(ctx, cacheKey)

	log.Debug().
		Str("collection", collName).
		Str("id", id).
		Msg("Document replaced")

	return nil
}
