// Package mongodb provides general Mongo database utilities, such as connecting to
// a Mongo database.
// Configuration of these utilities is limited, as they aim to be standard, easy-to-use
// APIs. If more control of the database is required, using the Mongo driver directly
// is recommended.
package mongodb

import (
	"context"
	"strconv"
	"sync"

	"github.com/newrelic/go-agent/v3/integrations/nrmongo"
	"github.com/nickhstr/goweb/env"
	"github.com/nickhstr/goweb/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client     *mongo.Client
	clientErr  error
	clientInit sync.Once
	log        = logger.New("mongodb")
)

// Client initializes a connected Mongo client only once,
// always returning the same instance.
func Client(ctx context.Context) (*mongo.Client, error) {
	clientInit.Do(func() {
		client, clientErr = NewClient(ctx)
		if clientErr == nil {
			clientErr = client.Connect(ctx)
		}
	})

	return client, clientErr
}

// NewClient creates a new opinionated client.
func NewClient(ctx context.Context) (*mongo.Client, error) {
	opts := options.Client().
		ApplyURI(env.Get("MONGO_URI")).
		SetMonitor(nrmongo.NewCommandMonitor(nil))

	maxPoolSize := env.Get("MONGO_POOL_SIZE")
	if maxPoolSize != "" {
		poolSize, err := strconv.Atoi(maxPoolSize)
		if err != nil {
			// Don't set default pool size. Mongo will do that for us.
			log.Err(err).Msg("Invalid MONGO_POOL_SIZE env var, using default size")
		}

		opts.SetMaxPoolSize(uint64(poolSize))
	}

	return mongo.NewClient(opts)
}

// DB retrieves the named database.
func DB(ctx context.Context, name string) (*mongo.Database, error) {
	var (
		db  *mongo.Database
		err error
	)

	client, err := Client(ctx)
	if err != nil {
		return db, err
	}

	db = client.Database(name)

	return db, err
}

// Collection grabs the named collection from the given database name.
// If the collection does not exist, it will be created when inserting.
func Collection(ctx context.Context, dbName string, collName string) (*mongo.Collection, error) {
	var (
		coll *mongo.Collection
		err  error
	)

	db, err := DB(ctx, dbName)
	if err != nil {
		return coll, err
	}

	coll = db.Collection(collName)

	return coll, err
}
