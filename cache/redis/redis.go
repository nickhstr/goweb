// Package redis provides a wrapper around github.com/go-redis/redis, specifically
// to satisfy the cache.Cacher interface.
package redis

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/newrelic/go-agent/v3/integrations/nrredis-v7"
	"github.com/nickhstr/goweb/config"
	"github.com/nickhstr/goweb/logger"
	"github.com/spf13/viper"
)

var log = logger.New("redis")

type redisCacher interface {
	Del(...string) *redis.IntCmd
	Get(string) *redis.StringCmd
	Set(string, interface{}, time.Duration) *redis.StatusCmd
}

type Cacher interface {
	Del(context.Context, ...string) error
	Get(context.Context, string) ([]byte, error)
	Set(context.Context, string, interface{}, time.Duration) error
}

// Client holds a redisCacher instance, and satisfies the cache.Cache interface.
type Client struct {
	client redisCacher
}

// Del deletes keys.
func (c *Client) Del(ctx context.Context, keys ...string) error {
	var err error

	switch cc := c.client.(type) {
	case *redis.ClusterClient:
		_, err = cc.WithContext(ctx).Del(keys...).Result()
	case *redis.Client:
		_, err = cc.WithContext(ctx).Del(keys...).Result()
	}

	if err != nil {
		log.Err(err).
			Str("keys", strings.Join(keys, ",")).
			Str("command", "DEL").
			Msg("Redis command failed")
	}

	return err
}

// Get returns the data stored under a key.
func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	var (
		data []byte
		err  error
	)

	switch cc := c.client.(type) {
	case *redis.ClusterClient:
		data, err = cc.WithContext(ctx).Get(key).Bytes()
	case *redis.Client:
		data, err = cc.WithContext(ctx).Get(key).Bytes()
	}

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
func (c *Client) Set(ctx context.Context, key string, val interface{}, t time.Duration) error {
	var err error

	switch cc := c.client.(type) {
	case *redis.ClusterClient:
		_, err = cc.WithContext(ctx).Set(key, val, t).Result()
	case *redis.Client:
		_, err = cc.WithContext(ctx).Set(key, val, t).Result()
	}

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

	mode := viper.GetString("REDIS_MODE")
	if mode != "" {
		// a mode is supplied, validate the base
		// variables required of every mode
		if err := config.Validate([]string{
			"REDIS_HOST",
			"REDIS_PORT",
		}); err != nil {
			log.Err(err).Msg("redis new Cacher error")
			return &noopClient{}
		}
	}

	addr := net.JoinHostPort(viper.GetString("REDIS_HOST"), viper.GetString("REDIS_PORT"))
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

	switch mode {
	case "cluster":
		clusterOptions := &redis.ClusterOptions{
			Addrs:           []string{addr},
			MaxRetries:      maxRetries,
			MinRetryBackoff: minRetryBackoff,
			MaxRetryBackoff: maxRetryBackoff,
			OnConnect:       onConnect,
		}
		rc := redis.NewClusterClient(clusterOptions)
		rc.AddHook(nrredis.NewHook(nil))
		c = &Client{rc}
	case "server":
		options := &redis.Options{
			Addr:            addr,
			MaxRetries:      maxRetries,
			MinRetryBackoff: minRetryBackoff,
			MaxRetryBackoff: maxRetryBackoff,
			OnConnect:       onConnect,
		}
		rc := redis.NewClient(options)
		rc.AddHook(nrredis.NewHook(options))
		c = &Client{rc}
	case "sentinel":
		if err := config.Validate([]string{
			"REDIS_PASSWORD",
			"REDIS_MASTER_NAME",
		}); err != nil {
			// required variables are not defined,
			// return a no-op client
			log.Err(err).Msg("redis new Cacher error")
			return &noopClient{}
		}

		password := viper.GetString("REDIS_PASSWORD")
		masterName := viper.GetString("REDIS_MASTER_NAME")
		sentinelOptions := &redis.FailoverOptions{
			SentinelAddrs:    []string{addr},
			SentinelPassword: password,
			Password:         password,
			MasterName:       masterName,
			MaxRetries:       maxRetries,
			MinRetryBackoff:  minRetryBackoff,
			MaxRetryBackoff:  maxRetryBackoff,
			OnConnect:        onConnect,
		}
		rc := redis.NewFailoverClient(sentinelOptions)
		rc.AddHook(nrredis.NewHook(nil))
		c = &Client{rc}
	default:
		// supplied mode is undefined or not a supported mode
		// default to the no-op Cacher
		c = &noopClient{}
	}

	return c
}

// Noop helper for when the necessary redis environment variables are not set, and we don't want to
// create redis connection errors.
type noopClient struct{}

var noopMsg = "redis noop"

func (n noopClient) Del(ctx context.Context, keys ...string) error {
	return errors.New(noopMsg)
}
func (n noopClient) Get(ctx context.Context, key string) ([]byte, error) {
	return []byte{}, errors.New(noopMsg)
}
func (n noopClient) Set(ctx context.Context, key string, val interface{}, t time.Duration) error {
	return errors.New(noopMsg)
}
