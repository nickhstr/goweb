package mongodb

import (
	"context"
	"errors"
)

var ErrBadDelete = errors.New("mongodb: document does not exist")

// DeleteOne attempts to delete a document by ID.
func (db *DB) DeleteOne(ctx context.Context, collName, id string) error {
	log := log.With().
		Str("collection", collName).
		Str("id", id).
		Logger()

	coll := db.Collection(collName)
	filter := db.IDFilter(id)

	result, err := coll.DeleteOne(ctx, filter)
	if err != nil {
		log.Err(err).Msg("DeleteOne failed")
		return err
	}

	// If no entity was deleted, return an error to indicate a bad request
	if result.DeletedCount < 1 {
		return ErrBadDelete
	}

	// we can the error here since we know the filter is valid bson
	cacheKey, _ := db.CacheKey(collName, id)

	_ = db.cacher.Del(ctx, cacheKey)

	log.Debug().
		Str("collection", collName).
		Str("id", id).
		Msg("Document deleted")

	return nil
}
