package cache

import (
	"errors"
	"io"
	"time"

	"github.com/go-redis/redis" // nolint: gotype
	"github.com/nickhstr/goweb/env"
	"github.com/nickhstr/goweb/logger"
)

type redisClient interface {
	Get(string) *redis.StringCmd
	Set(string, interface{}, time.Duration) *redis.StatusCmd
}

// var logger = log.With().Str("namespace", "redis").Logger()
var log = logger.New(nil).With().Str("namespace", "redis").Logger()
var redisInstance redisClient

// Get returns the data stored under the given key.
func Get(key string) ([]byte, error) {
	var (
		data []byte
		err  error
	)

	redis := getRedis()
	if redis == nil {
		err = errors.New("No redis client available")
		log.Warn().Str("operation", "GET").Msg(err.Error())

		return []byte{}, err
	}

	data, err = redis.Get(key).Bytes()
	if err != nil {
		if err.Error() == "redis: nil" {
			log.Warn().
				Str("operation", "GET").
				Str("key", key).
				Msg("Key not found")
		} else if err == io.EOF {
			log.Error().
				Str("operation", "GET").
				Str("key", key).
				Err(err).
				Msg("Redis unavailable")
		} else {
			log.Error().
				Str("operation", "GET").
				Str("key", key).
				Err(err).
				Msg(err.Error())
		}
	}

	return data, err
}

// Set stores data for a set period of time at the given key.
func Set(key string, data []byte, expiration time.Duration) {
	redis := getRedis()
	if redis == nil {
		log.Warn().
			Str("operation", "SET").
			Msg("No redis client available")

		return
	}

	_, err := redis.Set(key, data, expiration).Result()
	if err != nil {
		log.Error().
			Str("operation", "SET").
			Err(err).
			Msg(err.Error())
	}
}

func getRedis() redisClient {
	if env.Get("REDIS_HOST") == "" ||
		env.Get("REDIS_PORT") == "" ||
		env.Get("REDIS_MODE") == "" {
		log.Error().
			Str("redis-host", env.Get("REDIS_HOST")).
			Str("redis-port", env.Get("REDIS_PORT")).
			Str("redis-mode", env.Get("REDIS_MODE")).
			Msg("Environment variable(s) not set")

		return nil
	}

	if redisInstance != nil {
		return redisInstance
	}

	addr := env.Get("REDIS_HOST", "localhost") + ":" + env.Get("REDIS_PORT", "6379")
	mode := env.Get("REDIS_MODE", "server")
	maxRetries := 20
	minRetryBackoff := 50 * time.Millisecond
	maxRetryBackoff := 2000 * time.Millisecond
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
		redisInstance = redis.NewClusterClient(clusterOptions)
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
		redisInstance = redis.NewClient(options)
	}

	return redisInstance
}
