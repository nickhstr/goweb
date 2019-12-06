// Package redis provides a wrapper around github.com/go-redis/redis, specifically
// to satisfy the cache.Cacher interface.
package redis

import (
	"errors"
	"net"
	"strings"
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

type Cacher interface {
	Del(...string) error
	Get(string) ([]byte, error)
	Set(string, interface{}, time.Duration) error
}

// Client holds a redisClient instance, and satisfies the cache.Cacher interface.
type Client struct {
	client redisClient
}

// Del deletes keys.
func (cc *Client) Del(keys ...string) error {
	_, err := cc.client.Del(keys...).Result()
	if err != nil {
		log.Err(err).
			Str("keys", strings.Join(keys, ",")).
			Str("command", "DEL").
			Msg("Redis command failed")
	}
	return err
}

// Get returns the data stored under a key.
func (cc *Client) Get(key string) ([]byte, error) {
	data, err := cc.client.Get(key).Bytes()
	if err != nil {
		if err.Error() == "redis: nil" {
			log.Debug().
				Str("key", key).
				Msg("Key not found")
		} else {
			log.Err(err).
				Str("key", key).
				Str("command", "GET").
				Msg("Redis command failed")
		}
	}
	return data, err
}

// Set stores data under a key for a set amount of time.
func (cc *Client) Set(key string, val interface{}, t time.Duration) error {
	_, err := cc.client.Set(key, val, t).Result()
	if err != nil {
		log.Err(err).
			Str("key", key).
			Str("command", "SET").
			Msg("Redis command failed")
	}
	return err
}

// New returns an instance of Cacher.
func New() Cacher {
	var c Cacher

	addr := net.JoinHostPort(env.Get("REDIS_HOST"), env.Get("REDIS_PORT"))
	mode := env.Get("REDIS_MODE")
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
		c = &Client{rc}
	case "server":
		options := &redis.Options{
			Addr:            addr,
			MaxRetries:      maxRetries,
			MinRetryBackoff: minRetryBackoff,
			MaxRetryBackoff: maxRetryBackoff,
			OnConnect:       onConnect,
		}
		rc = redis.NewClient(options)
		c = &Client{rc}
	default:
		// supplied mode is undefined or not a supported mode
		// default to the noop Cacher
		c = &noopClient{}
	}

	return c
}

// Noop helper for when the necessary redis environment variables are not set, and we don't want to
// create redis connection errors.
type noopClient struct{}

var noopMsg = "redis noop"

func (n noopClient) Del(keys ...string) error {
	return errors.New(noopMsg)
}
func (n noopClient) Get(key string) ([]byte, error) {
	return []byte{}, errors.New(noopMsg)
}
func (n noopClient) Set(key string, val interface{}, t time.Duration) error {
	return errors.New(noopMsg)
}
