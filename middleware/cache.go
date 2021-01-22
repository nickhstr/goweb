package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/nickhstr/goweb/cache"
	"github.com/nickhstr/goweb/write"
	"github.com/rs/zerolog/hlog"
)

// CacheOptions are the configurable options for the cache middleware.
// These determine how exactly responses are cached.
type CacheOptions struct {
	// AllowedMethods is the slice of allowed HTTP request methods
	// which can be cached.
	// Defaults are: GET, HEAD.
	AllowedMethods []string

	// AllowedStatuses is the slice of allowed HTTP response status
	// codes which can be cached.
	// Default is: 200.
	AllowedStatuses []int

	// KeyPrefix is the optional cache key prefix.
	KeyPrefix string

	// TTL is the time-to-live for the cached data.
	// Default is: 15 minutes.
	TTL time.Duration

	// UseStale, when true, allows stale cache data to be used in the
	// case of internal server errors.
	UseStale bool

	// StaleStatuses is the slice of status codes allowed to use stale
	// cache data as a fall-back.
	// Default is: 500.
	StaleStatuses []int

	// StaleTTL is the time-to-live for stale cached data, before
	// being automaticlly purged.
	StaleTTL time.Duration
}

// CacheWriter is an enhanced http.ResponseWriter.
// Write operations are duplicated to a data buffer, which is used
// to cache the response.
type CacheWriter struct {
	http.ResponseWriter

	// cacheBuff records data written as the response body.
	cacheBuff *bytes.Buffer

	// statusCode is the response's HTTP status code.
	statusCode int

	useStale      bool
	staleStatuses []int
}

// NewCacheWriter creates a new CacheWriter.
func NewCacheWriter(w http.ResponseWriter, useStale bool, staleStatuses []int) *CacheWriter {
	return &CacheWriter{
		middleware.NewWrapResponseWriter(w, 0),
		&bytes.Buffer{},
		0,
		useStale,
		staleStatuses,
	}
}

// Write writes data to both a ResponseWriter and a
// cache data buffer.
// If stale cache data may be used, only the buffer writes data;
// stale data is written later, using the underlying ResponseWriter
// directly.
func (cw *CacheWriter) Write(p []byte) (int, error) {
	var (
		n   int
		err error
	)

	writers := []io.Writer{
		cw.cacheBuff,
	}

	if !cw.useStale || !includesStaleStatus(cw.statusCode, cw.staleStatuses) {
		writers = append(writers, cw.ResponseWriter)
	}

	for _, w := range writers {
		n, err = w.Write(p)
		if err != nil {
			return n, err
		}

		if n != len(p) {
			err = io.ErrShortWrite
			return n, err
		}
	}

	return len(p), nil
}

func (cw *CacheWriter) WriteHeader(statusCode int) {
	cw.statusCode = statusCode
	if cw.useStale && includesStaleStatus(statusCode, cw.staleStatuses) {
		return
	}

	cw.ResponseWriter.WriteHeader(statusCode)
}

// ReadAll reads all of the data in the cache buffer.
func (cw *CacheWriter) ReadAll() ([]byte, error) {
	return ioutil.ReadAll(cw.cacheBuff)
}

func includesStaleStatus(statusCode int, staleStatuses []int) bool {
	hasStatus := false

	for _, status := range staleStatuses {
		if statusCode == status {
			hasStatus = true
			break
		}
	}

	return hasStatus
}

// ignoredQueryParams are common query params which
// can be ignored when creating a cache key.
var ignoredQueryParams = map[string]struct{}{
	"apiKey": {},
	"cache":  {},
}

func CacheKey(prefix string, r *http.Request) string {
	// in this case, we can safely ignore the error, as
	// we're just copying the query params
	query, _ := url.ParseQuery(r.URL.RawQuery)

	for key := range ignoredQueryParams {
		query.Del(key)
	}

	u := &url.URL{
		Path:     r.URL.Path,
		RawQuery: query.Encode(),
	}

	return prefix + u.String()
}

// AddCacheHeader adds a response header to indicate the resopnse
// is cached.
func AddCacheHeader(h http.Header) http.Header {
	h.Set("x-cached-response", "true")
	return h
}

// ContextFromRequest returns a new context, which may or may not
// add a "no cache" flag.
// Contexts created from this can be used with cache.UseCache to
// determine if the cache can be used.
func ContextFromRequest(r *http.Request) context.Context {
	ctx := r.Context()
	nocache := r.URL.Query().Get("cache") == "false"

	if nocache {
		return cache.ContextWithNoCache(ctx)
	}

	return ctx
}

type CachedResponse struct {
	http.Header

	// Body is the HTTP response.
	Body []byte

	// StatusCode is the HTTP response status code.
	StatusCode int

	// Expiration is the time when the cached data should expire.
	Expiration int64
}

// Cache middleware creator offers caching of responses.
// Unlike caching driven by cache-control headers, responses are
// cached by an external Cacher, so even first-time requesters can
// benefit from cached responses.
// Optionally, stale cache data can be returned in cases of internal
// server errors, to protect against downtime.
func Cache(c cache.Cacher, opts CacheOptions) Middleware {
	if len(opts.AllowedMethods) == 0 {
		opts.AllowedMethods = []string{
			http.MethodGet,
			http.MethodHead,
		}
	}

	if len(opts.AllowedStatuses) == 0 {
		opts.AllowedStatuses = []int{
			0,
			http.StatusOK,
		}
	}

	if opts.TTL == 0 {
		opts.TTL = 15 * time.Minute
	}

	if len(opts.StaleStatuses) == 0 {
		opts.StaleStatuses = []int{
			http.StatusInternalServerError,
		}
	}

	if opts.StaleTTL == 0 {
		opts.StaleTTL = 24 * time.Hour
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := ContextFromRequest(r)
			log := hlog.FromRequest(r)

			methodNotAllowed := true
			for _, method := range opts.AllowedMethods {
				if r.Method == method {
					methodNotAllowed = false
					break
				}
			}

			if !cache.UseCache(ctx) || methodNotAllowed {
				next.ServeHTTP(w, r)
				return
			}

			var resp CachedResponse
			cacheKey := CacheKey(opts.KeyPrefix, r)

			// try to write cached data
			data, cacheErr := c.Get(ctx, cacheKey)
			if cacheErr == nil {
				cacheErr = json.Unmarshal(data, &resp)
				if cacheErr != nil {
					log.Err(cacheErr).Msg("Failed to unmarshal cached response")
					write.Error(w, cacheErr.Error(), http.StatusInternalServerError)
					return
				}

				if time.Now().Unix() < resp.Expiration {
					// copy cached response headers from downstream handlers
					for key := range resp.Header {
						if w.Header().Get(key) == "" {
							w.Header().Set(key, resp.Header.Get(key))
						}
					}

					AddCacheHeader(w.Header())

					if resp.StatusCode != 0 {
						w.WriteHeader(resp.StatusCode)
					}

					w.Write(resp.Body)
					return
				}
			}

			// get fresh response from handler
			cw := NewCacheWriter(w, opts.UseStale, opts.StaleStatuses)
			next.ServeHTTP(cw, r)

			statusCodeNotAllowed := true
			for _, status := range opts.AllowedStatuses {
				if cw.statusCode == status {
					statusCodeNotAllowed = false
					break
				}
			}

			if statusCodeNotAllowed {
				fmt.Println("status code:", cw.statusCode)

				if opts.UseStale {
					// If stale data can be used, the response needs
					// to be written to the ResponseWriter; the
					// CacheWriter only wrote the response body to
					// its internal buffer.
					if includesStaleStatus(cw.statusCode, opts.StaleStatuses) {
						if cacheErr == nil {
							for key := range resp.Header {
								w.Header().Set(key, resp.Header.Get(key))
							}

							AddCacheHeader(w.Header())
							w.WriteHeader(resp.StatusCode)
							w.Write(resp.Body)

							return
						}

						data, _ := cw.ReadAll()
						w.WriteHeader(cw.statusCode)
						w.Write(data)
					}
				}

				// covers cases where previous responses were cached
				_ = c.Del(ctx, cacheKey)

				// response has been written, end early
				return
			}

			body, err := cw.ReadAll()
			if err != nil {
				log.Err(err).Msg("Failed to read cache buffer")
				return
			}

			cachedResp := CachedResponse{
				cw.Header().Clone(),
				body,
				cw.statusCode,
				time.Now().Unix() + int64(opts.TTL),
			}

			data, err = json.Marshal(&cachedResp)
			if err != nil {
				log.Err(err).Msg("Failed to marshal cached response")
				write.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// store response in cache
			err = c.Set(ctx, cacheKey, data, opts.StaleTTL)
			if err != nil {
				log.Err(err).Msg("Failed to set data in cache")
			}
		})
	}
}
