package middleware_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nickhstr/goweb/middleware"
	"github.com/stretchr/testify/assert"
)

var goodResponseHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "all good in the hood")
})

var notFoundResponseHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "nowhere to be found")
})

var badResponseHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprint(w, "something bad happened")
})

type testCache struct {
	hits  int
	cache map[string][]byte
}

func newTestCache() *testCache {
	return &testCache{
		0,
		map[string][]byte{},
	}
}

func newTestCacheWithData(data map[string][]byte) *testCache {
	return &testCache{
		0,
		data,
	}
}

func (t *testCache) Del(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		delete(t.cache, key)
	}

	return nil
}

func (t *testCache) Get(ctx context.Context, key string) ([]byte, error) {
	if data, ok := t.cache[key]; ok {
		t.hits++
		return data, nil
	}

	return []byte{}, fmt.Errorf("no data under key: %s", key)
}

func (t *testCache) Set(ctx context.Context, key string, data interface{}, d time.Duration) error {
	dataBytes, ok := data.([]byte)
	if !ok {
		return errors.New("for testing, use []byte for data")
	}

	t.cache[key] = dataBytes

	return nil
}

func (t *testCache) Hits() int {
	return t.hits
}

func marshallJSON(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}

func TestCache(t *testing.T) {
	tests := []struct {
		name                 string
		cacher               *testCache
		cacheOpts            middleware.CacheOptions
		requestMethod        string
		requestPath          string
		handler              http.Handler
		expectCachedResponse bool
		expectedBody         []byte
		expectedStatusCode   int
	}{
		{
			"good responses should be cached by default",
			newTestCache(),
			middleware.CacheOptions{},
			http.MethodGet,
			"/some/path/to/good/endpoint",
			goodResponseHandler,
			true,
			[]byte("all good in the hood"),
			http.StatusOK,
		},
		{
			"not found responses should not be cached by default",
			newTestCache(),
			middleware.CacheOptions{},
			http.MethodGet,
			"/some/path/to/not-found/endpoint",
			notFoundResponseHandler,
			false,
			[]byte("nowhere to be found"),
			http.StatusNotFound,
		},
		{
			"bad responses should not be cached by default",
			newTestCache(),
			middleware.CacheOptions{},
			http.MethodGet,
			"/some/path/to/bad/endpoint",
			badResponseHandler,
			false,
			[]byte("something bad happened"),
			http.StatusInternalServerError,
		},
		{
			"requests with cache=false query param should not use cache",
			newTestCache(),
			middleware.CacheOptions{},
			http.MethodGet,
			"/some/path/to/good/endpoint?cache=false",
			goodResponseHandler,
			false,
			[]byte("all good in the hood"),
			http.StatusOK,
		},
		{
			"when configured to use stale, cache should be used for responses with allowed stale response codes",
			newTestCacheWithData(map[string][]byte{
				"/some/path/to/bad/endpoint": marshallJSON(&middleware.CachedResponse{
					Header:     http.Header{},
					Body:       []byte("some preexisting response body text"),
					StatusCode: 200,
					Expiration: 1257894000,
				}),
			}),
			middleware.CacheOptions{
				UseStale: true,
			},
			http.MethodGet,
			"/some/path/to/bad/endpoint",
			badResponseHandler,
			true,
			[]byte("some preexisting response body text"),
			http.StatusOK,
		},
		{
			"when configured to use stale, cache should not be used for responses with disallowed stale response codes",
			newTestCacheWithData(map[string][]byte{
				"/some/path/to/bad/endpoint": marshallJSON(&middleware.CachedResponse{
					Header:     http.Header{},
					Body:       []byte("some preexisting response body text"),
					StatusCode: 200,
					Expiration: 1257894000,
				}),
			}),
			middleware.CacheOptions{
				UseStale: true,
			},
			http.MethodGet,
			"/some/path/to/bad/endpoint",
			notFoundResponseHandler,
			false,
			[]byte("nowhere to be found"),
			http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			firstRec := httptest.NewRecorder()
			secondRec := httptest.NewRecorder()
			handler := middleware.Cache(test.cacher, test.cacheOpts)(test.handler)
			req, _ := http.NewRequest(test.requestMethod, test.requestPath, nil)

			handler.ServeHTTP(firstRec, req)
			handler.ServeHTTP(secondRec, req)

			maybeCachedBody, _ := ioutil.ReadAll(secondRec.Result().Body)
			cachedResponseHeader := secondRec.Result().Header.Get("x-cached-response")

			if test.expectCachedResponse {
				assert.Equal("true", cachedResponseHeader)
			} else {
				assert.Equal("", cachedResponseHeader)
			}
			assert.Equal(test.expectedBody, maybeCachedBody)
			assert.Equal(test.expectedStatusCode, secondRec.Result().StatusCode)
		})
	}
}
