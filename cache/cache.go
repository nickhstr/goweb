// Package cache provides a simple key-value cache store, supporting
// just a handful of caching operations.
package cache

import (
	"errors"
	"sync"
	"time"

	"github.com/nickhstr/goweb/cache/redis"
	"github.com/nickhstr/goweb/logger"
)

var log = logger.New("cache")
var client Cacher

// Cacher defines the methods of any cache client.
type Cacher interface {
	Del(...string) error
	Get(string) ([]byte, error)
	Set(string, interface{}, time.Duration) error
}

// Del removes data at the given key(s)
func Del(keys ...string) error {
	// init default Cacher if not set already
	CacherInit(nil)
	log := log.With().Str("operation", "DEL").Logger()

	if client == nil {
		err := noClientLogErr(log)
		return err
	}

	return client.Del(keys...)
}

// Get returns the data stored under the given key.
func Get(key string) ([]byte, error) {
	// init default Cacher if not set already
	CacherInit(nil)
	log := log.With().Str("operation", "GET").Logger()

	if client == nil {
		err := noClientLogErr(log)
		return []byte{}, err
	}

	return client.Get(key)
}

// Set stores data for a set period of time at the given key.
func Set(key string, data []byte, expiration time.Duration) error {
	// init default Cacher if not set already
	CacherInit(nil)
	log := log.With().Str("operation", "SET").Logger()

	if client == nil {
		err := noClientLogErr(log)
		return err
	}

	return client.Set(key, data, expiration)
}

var cacherInit sync.Once

// CacherInit sets the Cacher to be used for all cache operations.
// If an init func is supplied, it will be used for setup; otherwise,
// the default Cacher will be used.
// The supplied init function must accept a Cacher as its argument, so
// that `client` may be set.
func CacherInit(init func() Cacher) {
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

// Creates no-client error, logs it, and returns it
func noClientLogErr(log logger.Logger) error {
	err := errors.New("no cache client available")
	log.Error().Msg(err.Error())
	return err
}
