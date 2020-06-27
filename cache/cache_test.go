package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/nickhstr/goweb/cache"
	"github.com/stretchr/testify/assert"
)

type mockCache struct{}

func (mc mockCache) Del(ctx context.Context, key ...string) error {
	return nil
}
func (mc mockCache) Get(ctx context.Context, key string) ([]byte, error) {
	return []byte{}, nil
}
func (mc mockCache) Set(ctx context.Context, key string, val interface{}, t time.Duration) error {
	return nil
}

func TestDel(t *testing.T) {
	assert := assert.New(t)

	cache.Init(func() cache.Cacher {
		return &mockCache{}
	})

	err := cache.Del(context.Background(), "key")

	assert.Nil(err)
}

func TestGet(t *testing.T) {
	assert := assert.New(t)

	cache.Init(func() cache.Cacher {
		return &mockCache{}
	})

	val, err := cache.Get(context.Background(), "key")

	assert.Nil(err)
	assert.Equal([]byte{}, val)
}

func TestSet(t *testing.T) {
	assert := assert.New(t)

	cache.Init(func() cache.Cacher {
		return &mockCache{}
	})

	assert.NotPanics(func() { _ = cache.Set(context.Background(), "key", []byte{}, 60*time.Second) })
}
