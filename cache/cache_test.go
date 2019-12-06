package cache_test

import (
	"testing"
	"time"

	"github.com/nickhstr/goweb/cache"
	"github.com/stretchr/testify/assert"
)

type mockCacher struct{}

func (mc mockCacher) Del(key ...string) error {
	return nil
}
func (mc mockCacher) Get(key string) ([]byte, error) {
	return []byte{}, nil
}
func (mc mockCacher) Set(key string, val interface{}, t time.Duration) error {
	return nil
}

func TestDel(t *testing.T) {
	assert := assert.New(t)
	cache.CacherInit(func() cache.Cacher {
		return &mockCacher{}
	})
	err := cache.Del("key")

	assert.Nil(err)
}

func TestGet(t *testing.T) {
	assert := assert.New(t)
	cache.CacherInit(func() cache.Cacher {
		return &mockCacher{}
	})
	val, err := cache.Get("key")

	assert.Nil(err)
	assert.Equal([]byte{}, val)
}

func TestSet(t *testing.T) {
	assert := assert.New(t)
	cache.CacherInit(func() cache.Cacher {
		return &mockCacher{}
	})

	assert.NotPanics(func() { _ = cache.Set("key", []byte{}, 60*time.Second) })
}
