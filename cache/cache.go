// Package cache provides a simple interface for cache implementations.
package cache

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/nickhstr/goweb/cache/redis"
	"github.com/nickhstr/goweb/logger"
)

var log = logger.New("cache")

// Cacher defines the methods of any cache client.
type Cacher interface {
	// Del deletes the given key(s).
	Del(context.Context, ...string) error
	// Get returns the bytes stored under the given key.
	Get(context.Context, string) ([]byte, error)
	// Set stores a value under a given key, for as long
	// as the given duration.
	Set(context.Context, string, interface{}, time.Duration) error
}

var ErrNoop = errors.New("noop cacher")

// Noop provides a no-op Cacher, useful for testing or other
// environments where a real Cacher cannot be used.
type Noop struct {
	// should the no-op operations error
	shouldErr bool
}

// NewNoop returns a new Noop Cacher.
// It can be configured to error, or
// not error, for all operations.
func NewNoop(shouldErr bool) Cacher {
	return &Noop{shouldErr}
}

func (n Noop) Del(ctx context.Context, keys ...string) error {
	log.Debug().Str("operation", "DEL").Msg("noop operation")

	if n.shouldErr {
		return ErrNoop
	}

	return nil
}
func (n Noop) Get(ctx context.Context, key string) ([]byte, error) {
	log.Debug().Str("operation", "GET").Msg("noop operation")

	if n.shouldErr {
		return []byte{}, ErrNoop
	}

	return []byte{}, nil
}
func (n Noop) Set(ctx context.Context, key string, data interface{}, expiration time.Duration) error {
	log.Debug().Str("operation", "SET").Msg("noop operation")

	if n.shouldErr {
		return ErrNoop
	}

	return nil
}

// Default returns the default Cacher.
func Default() Cacher {
	return redis.New()
}

// PrefixedCacher wraps a Cacher, and prefixes all cache keys.
type PrefixedCacher struct {
	client    Cacher
	keyPrefix string
}

// NewPrefixedCacher returns a new PrefixedCacher instance.
func NewPrefixedCacher(c Cacher, keyPrefix string) *PrefixedCacher {
	return &PrefixedCacher{c, keyPrefix}
}

func (p *PrefixedCacher) Del(ctx context.Context, keys ...string) error {
	for i, key := range keys {
		keys[i] = p.keyPrefix + key
	}

	return p.client.Del(ctx, keys...)
}
func (p *PrefixedCacher) Get(ctx context.Context, key string) ([]byte, error) {
	return p.client.Get(ctx, p.keyPrefix+key)
}

func (p *PrefixedCacher) Set(ctx context.Context, key string, v interface{}, d time.Duration) error {
	return p.client.Set(ctx, p.keyPrefix+key, v, d)
}

// noCacheContextKey is used in a context to indicate that the cache
// should not be used.
// An empty struct is used in favor of any other type (such as a
// string), as it uses less memory.
type noCacheContextKey struct{}

var nc = noCacheContextKey{}

// ContextWithNoCache creates a new context with
// the no cache key/value added to the parent context.
func ContextWithNoCache(parent context.Context) context.Context {
	// Use empty struct for value, as it uses
	// the least amount of memory.
	return context.WithValue(parent, nc, struct{}{})
}

// UseCache reports whether or not we can attempt
// to get data from the cache.
func UseCache(ctx context.Context) bool {
	// the noCacheContextKey is not in context,
	// we can try the cache
	v := ctx.Value(nc)
	return v == nil
}

// Key formats a slice of string arguments in a uniform, standard
// manner.
// Useful for when multiple parameters compose the cache key, and
// their order needs to remain the same.
func Key(args ...string) string {
	sort.Strings(args)

	return strings.Join(args, ";")
}
