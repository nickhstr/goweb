// Package redis provides a wrapper around github.com/go-redis/redis, specifically
// to satisfy the cache.Cacher interface.
package redis

import (
	"io"
	"net"
	"time"

	"github.com/go-redis/redis"
	"github.com/nickhstr/goweb/env"
	"github.com/nickhstr/goweb/logger"
)

var log = logger.New("redis")

type redisClient interface {
	Del(...string) *redis.IntCmd
	Get(string) *redis.StringCmd
	Set(string, interface{}, time.Duration) *redis.StatusCmd
}

// Cacher holds a redisClient instance, and satisfies the cache.Cacher interface.
type Cacher struct {
	client redisClient
}

// Del deletes keys.
func (c *Cacher) Del(keys ...string) error {
	_, err := c.client.Del(keys...).Result()
	return err
}

// Get returns the data stored under a key.
func (c *Cacher) Get(key string) ([]byte, error) {
	data, err := c.client.Get(key).Bytes()
	if err != nil {
		if err.Error() == "redis: nil" {
			log.Debug().
				Str("key", key).
				Msg("Key not found")
		} else if err == io.EOF {
			log.Error().
				Str("key", key).
				Err(err).
				Msg("Redis unavailable")
		}
	}
	return data, err
}

// Set stores data under a key for a set amount of time.
func (c *Cacher) Set(key string, val interface{}, t time.Duration) error {
	_, err := c.client.Set(key, val, t).Result()
	return err
}

// New returns an instance of Cacher.
func New() *Cacher {
	addr := net.JoinHostPort(
		env.Get("REDIS_HOST", "localhost"),
		env.Get("REDIS_PORT", "6379"),
	)
	mode := env.Get("REDIS_MODE", "server")
	maxRetries := 1
	minRetryBackoff := 8 * time.Millisecond
	maxRetryBackoff := 512 * time.Millisecond
	onConnect := func(c *redis.Conn) error {
		log.Info().
			Str("address", addr).
			Str("mode", mode).
			Msg("Connected to Redis")
		return nil
	}

	var rc redisClient
	switch mode {
	case "cluster":
		clusterOptions := &redis.ClusterOptions{
			Addrs:           []string{addr},
			MaxRetries:      maxRetries,
			MinRetryBackoff: minRetryBackoff,
			MaxRetryBackoff: maxRetryBackoff,
			OnConnect:       onConnect,
		}
		rc = redis.NewClusterClient(clusterOptions)
	case "server":
		fallthrough
	default:
		options := &redis.Options{
			Addr:            addr,
			MaxRetries:      maxRetries,
			MinRetryBackoff: minRetryBackoff,
			MaxRetryBackoff: maxRetryBackoff,
			OnConnect:       onConnect,
		}
		rc = redis.NewClient(options)
	}

	return &Cacher{rc}
}
