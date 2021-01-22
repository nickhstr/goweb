package mongodb

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/nickhstr/goweb/cache"
	"github.com/nickhstr/goweb/logger"
	"github.com/nickhstr/goweb/mongodb"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var log = logger.New("mongodb")

var (
	ErrBadClient = errors.New("mongodb: bad client")
	ErrBadFilter = errors.New("mongodb: invalid filter")
)

// DB represents a Mongodb client, with an optional caching layer.
type DB struct {
	cacher         cache.Cacher
	cacheKeyPrefix string
	cacheTTL       time.Duration
	client         *mongo.Client
	connected      bool
	name           string
	mu             *sync.Mutex
}

// New creates a new DB instance.
// A default Mongodb client is used.
// Make sure to call Connect before using the DB.
func New(name string) (*DB, error) {
	client, err := mongodb.New(mongodb.ClientOptions{
		PoolSize:    viper.GetUint64("MONGO_POOL_SIZE"),
		URI:         viper.GetString("MONGO_URI"),
		UseNewRelic: viper.GetBool("MONGO_USE_NEW_RELIC"),
	})
	db := &DB{
		cache.Default(),
		name,
		5 * time.Minute,
		client,
		false,
		name,
		&sync.Mutex{},
	}

	if err != nil {
		err = fmt.Errorf("%w: %s", ErrBadClient, err.Error())
	}

	return db, err
}

// NewWithClient creates a new DB instance, configured with the
// supplied Mongodb client.
// Make sure to call Connect before using the DB.
func NewWithClient(name string, client *mongo.Client) *DB {
	return &DB{
		cache.Default(),
		name,
		5 * time.Minute,
		client,
		false,
		name,
		&sync.Mutex{},
	}
}

// Connect connects the Mongodb client with the server.
func (db *DB) Connect(ctx context.Context) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.connected {
		return nil
	}

	err := db.client.Connect(ctx)
	if err == nil {
		db.connected = true
	}

	return err
}

// Disconnect connects the Mongodb client with the server.
// Call this only when done using the DB.
func (db *DB) Disconnect(ctx context.Context) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if !db.connected {
		return nil
	}

	err := db.client.Disconnect(ctx)
	if err == nil {
		db.connected = false
	}

	return err
}

// SetCacher sets the DB's Cacher.
// Note: this is not thread safe, and does not coordinate
// handing-off from one Cacher to another. Use this when setting-up
// a DB.
func (db *DB) SetCacher(c cache.Cacher) *DB {
	db.cacher = c
	return db
}

// SetCacher sets the DB's cache key prefix.
// Note: this is not thread safe. Use this when setting-up a DB.
func (db *DB) SetCacheKeyPrefix(prefix string) *DB {
	db.cacheKeyPrefix = prefix
	return db
}

// SetCacher sets the DB's cache time-to-live (TTL).
// Note: this is not thread safe. Use this when setting-up a DB.
func (db *DB) SetCacheTTL(d time.Duration) *DB {
	db.cacheTTL = d
	return db
}

// Ping verifies if the client can connect to the database.
func (db *DB) Ping(ctx context.Context) error {
	return db.client.Ping(ctx, nil)
}

// Collection returns the named Mongodb collection.
func (db *DB) Collection(name string) *mongo.Collection {
	return db.client.Database(db.name).Collection(name)
}

// CacheKey generates a standardized cache key.
func (db *DB) CacheKey(collName string, filter interface{}, args ...string) (string, error) {
	return cacheKey(db.cacheKeyPrefix, collName, filter, args...)
}

func cacheKey(prefix, collName string, filter interface{}, args ...string) (string, error) {
	var key string

	encodedFilter, err := bson.Marshal(filter)
	if err != nil {
		return key, fmt.Errorf("%w: %s", ErrBadFilter, err.Error())
	}

	// Create hash from filter.
	// sha1 is pretty quick, and we don't need to worry about
	// cryptographical security for its use in a cache key.
	sum := fmt.Sprintf("%x", sha1.Sum(encodedFilter))
	keyArgs := append(
		args,
		"collection:"+collName,
		"filter:"+sum,
	)
	key = prefix + cache.Key(keyArgs...)

	return key, err
}

// IDFilter returns a filter for querying by ID.
func (db *DB) IDFilter(id string) bson.M {
	return bson.M{"_id": id}
}
