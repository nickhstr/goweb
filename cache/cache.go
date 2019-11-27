package cache

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/nickhstr/goweb/env"
	"github.com/nickhstr/goweb/logger"
)

type redisClient interface {
	Del(...string) *redis.IntCmd
	Get(string) *redis.StringCmd
	Set(string, interface{}, time.Duration) *redis.StatusCmd
}

var log = logger.New("redis")
var client redisClient
var clientInit sync.Once
var noClientMsg = "no redis client available"

// Del removes data at the given key(s)
func Del(keys ...string) error {
	clientInit.Do(clientSetup)
	log := log.With().Str("operation", "DEL").Logger()

	var err error

	if client == nil {
		err = errors.New(noClientMsg)
		log.Err(err).Msg(err.Error())
		return err
	}

	_, err = client.Del(keys...).Result()
	if err != nil {
		log.Warn().
			Err(err).
			Msg(err.Error())
	}

	return err
}

// Get returns the data stored under the given key.
func Get(key string) ([]byte, error) {
	clientInit.Do(clientSetup)
	log := log.With().Str("operation", "GET").Logger()

	var (
		data []byte
		err  error
	)

	if client == nil {
		err = errors.New(noClientMsg)
		log.Err(err).Msg(err.Error())
		return []byte{}, err
	}

	data, err = client.Get(key).Bytes()
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
		} else {
			log.Warn().
				Str("key", key).
				Err(err).
				Msg(err.Error())
		}
	}

	return data, err
}

// Set stores data for a set period of time at the given key.
func Set(key string, data []byte, expiration time.Duration) error {
	clientInit.Do(clientSetup)
	log := log.With().Str("operation", "SET").Logger()

	var err error

	if client == nil {
		err = errors.New(noClientMsg)
		log.Err(err).Msg(err.Error())
		return err
	}

	_, err = client.Set(key, data, expiration).Result()
	if err != nil {
		log.Warn().
			Err(err).
			Msg(err.Error())
		return err
	}

	return nil
}

func clientSetup() {
	if env.Get("REDIS_HOST") == "" ||
		env.Get("REDIS_PORT") == "" ||
		env.Get("REDIS_MODE") == "" {
		log.Error().
			Str("redis-host", env.Get("REDIS_HOST")).
			Str("redis-port", env.Get("REDIS_PORT")).
			Str("redis-mode", env.Get("REDIS_MODE")).
			Msg("Environment variable(s) not set")

		return
	}

	if client != nil {
		return
	}

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

	switch mode {
	case "cluster":
		clusterOptions := &redis.ClusterOptions{
			Addrs:           []string{addr},
			MaxRetries:      maxRetries,
			MinRetryBackoff: minRetryBackoff,
			MaxRetryBackoff: maxRetryBackoff,
			OnConnect:       onConnect,
		}
		client = redis.NewClusterClient(clusterOptions)
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
		client = redis.NewClient(options)
	}
}
