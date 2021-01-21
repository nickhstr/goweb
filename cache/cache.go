// Package cache provides a simple interface for cache implementations.
package cache

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/nickhstr/goweb/cache/redis"
)

// Cacher defines the methods of any cache client.
type Cacher interface {
	Del(context.Context, ...string) error
	Get(context.Context, string) ([]byte, error)
	Set(context.Context, string, interface{}, time.Duration) error
}

// Default returns the default Cacher.
func Default() Cacher {
	return redis.New()
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
