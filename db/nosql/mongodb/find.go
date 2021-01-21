package mongodb

import (
	"context"
	"errors"
	"fmt"

	"github.com/nickhstr/goweb/cache"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ErrBadFind serves as a base error to wrap for errors related to a query.
var ErrBadFind = errors.New("mongodb: bad collection query")

// FindByID queries by the given ID for a single document,
// decoding the found BSON document into result.
// Note, result must be a pointer.
func (db *DB) FindByID(ctx context.Context, collName, id string, result interface{}) error {
	filter := db.IDFilter(id)
	return db.FindOne(ctx, collName, filter, result)
}

// FindAll queries for multiple documents, decoding found
// documents into results.
// A start index and limit are required. Start must be at least zero,
// and less than the number of documents in the collection;
// limit can be zero (to get all documents), or some integer
// less than or equal to the document count.
// Note, results must be a pointer to a slice.
func (db *DB) FindAll(ctx context.Context, collName string, start, limit int64, results interface{}) error {
	return db.FindMany(ctx, collName, start, limit, bson.M{}, results)
}

// FindMany queries for multiple documents, using the
// given filter, and decodes found documents into results.
// A start index and limit are required. Start must be at least zero,
// and less than the number of matched documents in the collection;
// limit can be zero (to get all matched documents), or some integer
// less than or equal to the matched document count.
// Note, results must be a pointer to a slice.
func (db *DB) FindMany(ctx context.Context, collName string, start, limit int64, filter interface{}, results interface{}) error {
	var (
		count int64
		err   error
	)

	coll := db.Collection(collName)

	count, err = coll.CountDocuments(ctx, filter)
	if err != nil {
		return err
	}

	if count < 1 {
		return fmt.Errorf("%w: %s documents not found", ErrBadFind, collName)
	}

	// Check to make sure start index is valid
	if start >= count {
		err = fmt.Errorf(
			"%w: start index must be less than the number of documents (%d)",
			ErrBadFind,
			count,
		)

		return err
	}

	opts := options.Find()

	if start > 0 {
		opts.SetSkip(start)
	}

	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return err
	}

	return cursor.All(ctx, results)
}

// FindOne queries by the given filter for a single document,
// decoding the found BSON document into result.
// Note, result must be a pointer.
// By default, a cache is used; this can be bypassed by adding
// to the context a "no cache" flag, using cache.ContextWithNoCache.
func (db *DB) FindOne(ctx context.Context, collName string, filter interface{}, result interface{}) error {
	var err error

	cacheKey, err := db.CacheKey(collName, filter)
	if err != nil {
		return err
	}

	if cache.UseCache(ctx) {
		// try cache first
		data, err := db.cacher.Get(ctx, cacheKey)
		if err == nil {
			log.Debug().
				Str("collection", collName).
				Bool("cache", true).
				Msg("Document found")

			err = bson.Unmarshal(data, result)
			if err != nil {
				return err
			}

			return nil
		}
	}

	coll := db.Collection(collName)

	err = coll.FindOne(ctx, filter, options.FindOne()).Decode(result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("%w: %s document not found", ErrBadFind, collName)
		}

		return err
	}

	data, err := bson.Marshal(result)
	if err != nil {
		return err
	}

	_ = db.cacher.Set(ctx, cacheKey, data, db.cacheTTL)

	log.Debug().
		Str("collection", collName).
		Bool("cache", false).
		Msg("Document found")

	return err
}
