package mongodb

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
)

// ErrInsertConflict indicates an existing document of the same ID.
var ErrInsertConflict = errors.New("mongodb: insert failed, document exists")

// InsertOne tries to insert the given data into the collection.
// If data is already stored under the same ID, an error is returned.
func (db *DB) InsertOne(ctx context.Context, collName, id string, data interface{}) error {
	var err error

	log := log.With().
		Str("collection", collName).
		Str("id", id).
		Logger()

	coll := db.Collection(collName)
	filter := db.IDFilter(id)

	// Check to see if the document is already stored
	err = coll.FindOne(ctx, filter).Err()
	// Document already exists, or something went wrong
	if err != mongo.ErrNoDocuments {
		// Something went wrong
		if err != nil {
			log.Err(err).Msg("FindOne failed")
			return err
		}

		return ErrInsertConflict
	}

	_, err = coll.InsertOne(ctx, data)
	if err != nil {
		log.Err(err).Msg("InsertOne failed")
		return err
	}

	log.Debug().
		Str("collection", collName).
		Str("id", id).
		Msg("Document inserted")

	return nil
}
