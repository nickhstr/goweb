// Package cache provides a simple key-value cache store, supporting
// just a handful of caching operations.
// Note, if not explicitly set, a default cache client will be created;
// this cache client will be used for all usage of this package in a
// given application.
package cache

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/nickhstr/goweb/cache/redis"
	"github.com/nickhstr/goweb/logger"
)

var (
	client     Cacher
	cacherInit sync.Once
	log        = logger.New("cache")
)

// Cacher defines the methods of any cache client.
type Cacher interface {
	Del(context.Context, ...string) error
	Get(context.Context, string) ([]byte, error)
	Set(context.Context, string, interface{}, time.Duration) error
}

// Default returns the default Cacher.
func Default() Cacher {
	Init(nil)
	return client
}

// Init sets the Cacher to be used for all cache operations.
// If an init func is supplied, it will be used for setup;
// otherwise, the default Cacher will be used.
// The supplied init function must return a Cacher,
// so that `client` may be set.
func Init(init func() Cacher) {
	if init == nil {
		// default to redis.Cacher
		cacherInit.Do(func() {
			client = redis.New()
		})

		return
	}

	cacherInit.Do(func() {
		client = init()
	})
}

// Del removes data at the given key(s).
func Del(ctx context.Context, keys ...string) error {
	Init(nil)

	log := log.With().Str("operation", "DEL").Logger()

	if client == nil {
		err := noClientLogErr(log)
		return err
	}

	return client.Del(ctx, keys...)
}

// Get returns the data stored under the given key.
func Get(ctx context.Context, key string) ([]byte, error) {
	Init(nil)

	log := log.With().Str("operation", "GET").Logger()

	if client == nil {
		err := noClientLogErr(log)
		return []byte{}, err
	}

	return client.Get(ctx, key)
}

// Set stores data for a set period of time at the given key.
func Set(ctx context.Context, key string, data interface{}, expiration time.Duration) error {
	Init(nil)

	log := log.With().Str("operation", "SET").Logger()

	if client == nil {
		err := noClientLogErr(log)
		return err
	}

	return client.Set(ctx, key, data, expiration)
}

// Creates no-client error, logs it, and returns it.
func noClientLogErr(log logger.Logger) error {
	err := errors.New("no cache client available")
	log.Error().Msg(err.Error())

	return err
}
