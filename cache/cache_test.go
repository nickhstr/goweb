package cache

import (
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
)

type mockClient struct{}

func (mc mockClient) Del(key ...string) *redis.IntCmd {
	return &redis.IntCmd{}
}
func (mc mockClient) Get(key string) *redis.StringCmd {
	return &redis.StringCmd{}
}
func (mc mockClient) Set(key string, val interface{}, ttl time.Duration) *redis.StatusCmd {
	return &redis.StatusCmd{}
}

func setupEnv() func() {
	ogHost := os.Getenv("REDIS_HOST")
	ogPort := os.Getenv("REDIS_PORT")
	ogMode := os.Getenv("REDIS_MODE")

	_ = os.Setenv("REDIS_HOST", "localhost")
	_ = os.Setenv("REDIS_PORT", "6379")
	_ = os.Setenv("REDIS_MODE", "server")

	return func() {
		_ = os.Setenv("REDIS_HOST", ogHost)
		_ = os.Setenv("REDIS_PORT", ogPort)
		_ = os.Setenv("REDIS_MODE", ogMode)
	}
}

func TestDel(t *testing.T) {
	assert := assert.New(t)
	restoreEnv := setupEnv()
	defer restoreEnv()

	ogClient := client
	defer func() { client = ogClient }()
	client = mockClient{}

	err := Del("key")

	assert.Nil(err)
}

func TestGet(t *testing.T) {
	assert := assert.New(t)
	restoreEnv := setupEnv()
	defer restoreEnv()

	ogClient := client
	defer func() { client = ogClient }()
	client = mockClient{}

	val, err := Get("key")

	assert.Nil(err)
	assert.Equal([]byte{}, val)
}

func TestSet(t *testing.T) {
	assert := assert.New(t)
	restoreEnv := setupEnv()
	defer restoreEnv()

	ogClient := client
	defer func() { client = ogClient }()
	client = mockClient{}

	assert.NotPanics(func() { _ = Set("key", []byte{}, 60*time.Second) })
}
