package client

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nickhstr/goweb/cache"
	"github.com/nickhstr/goweb/logger"
)

const defaultCacheKeyPrefix = "dal:"

var (
	log               = logger.New("dal")
	maxAgeRegex       = regexp.MustCompile(`max-age=\d+`)
	defaultHTTPClient = &http.Client{
		Timeout: 15 * time.Second,
	}
)

// Client is an enhanced http.Client.
// By defualt, a caching layer is used
// for GET requests.
type Client struct {
	httpClient     *http.Client
	cacher         cache.Cacher
	cacheKeyPrefix string
	skipCache      bool
	ttl            time.Duration
}

// New returns a new Client instance.
func New() *Client {
	return &Client{
		httpClient:     defaultHTTPClient,
		cacher:         cache.Default(),
		cacheKeyPrefix: defaultCacheKeyPrefix,
		ttl:            60 * time.Second,
	}
}

// SetHTTPClient sets the Client's http.Client.
func (c *Client) SetHTTPClient(httpClient *http.Client) *Client {
	c.httpClient = httpClient
	return c
}

// SetCacher sets the client's Cacher.
func (c *Client) SetCacher(cacher cache.Cacher) *Client {
	c.cacher = cacher
	return c
}

// SetCacheKeyPrefix sets the prefix for all of its cache keys.
func (c *Client) SetCacheKeyPrefix(prefix string) *Client {
	c.cacheKeyPrefix = prefix
	return c
}

// SetSkipCache sets the skipCache option;
// true to bypass the cache, otherwise cache responses.
func (c *Client) SetSkipCache(skip bool) *Client {
	c.skipCache = skip
	return c
}

// SetTTL sets the Time to Live (TTL) for
// a request's cached response.
func (c *Client) SetTTL(ttl time.Duration) *Client {
	c.ttl = ttl
	return c
}

// Do sends the request, maybe caches the response,
// and returns the response.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	var (
		resp = new(http.Response)
		err  error
	)

	if req == nil {
		return nil, errors.New("client: nil request supplied")
	}

	start := time.Now()
	url := req.URL.String()
	cacheKey := c.cacheKeyPrefix + url
	ctx := req.Context()

	// only try cache for GET requests
	skipCache := c.skipCache || c.cacher == nil || req.Method != http.MethodGet

	if !skipCache {
		cachedResp, err := c.cacher.Get(ctx, cacheKey)
		if err == nil {
			log.Info().
				Str("url", url).
				Str("method", req.Method).
				Str("responseTime", time.Since(start).String()).
				Bool("cache", true).
				Msg("DAL request")

			resp.Body = ioutil.NopCloser(bytes.NewBuffer(cachedResp))
			resp.StatusCode = http.StatusOK

			return resp, err
		}
	}

	resp, err = c.httpClient.Do(req)
	if err != nil {
		log.Error().
			Str("url", url).
			Str("method", req.Method).
			Err(err).
			Msg("Failed to do request")

		return resp, err
	}

	if c.ttl == 0 {
		c.ttl = ttlFromResponse(resp)
	}

	log.Info().
		Str("url", url).
		Str("method", req.Method).
		Str("responseTime", time.Since(start).String()).
		Dur("ttl", c.ttl).
		Bool("cache", false).
		Msg("DAL request")

	if !skipCache {
		// read body to store in cache
		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			return resp, err
		}

		// Try to store response in cache,
		// swallowing any errors
		_ = c.cacher.Set(ctx, cacheKey, body, c.ttl)

		// restore response body with body just read
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}

	return resp, err
}

// ttlFromResponse attempts to get a TTL value from
// a response's "cache-control" header, otherwise
// returning a default.
func ttlFromResponse(r *http.Response) time.Duration {
	// Number of seconds for ttl
	var ttl int

	headerKey := http.CanonicalHeaderKey("Cache-Control")
	cacheControlValues := r.Header[headerKey]

	for _, val := range cacheControlValues {
		match := maxAgeRegex.FindString(val)

		if match != "" {
			maxAgeKeyVal := strings.Split(match, "=")

			maxAge, err := strconv.Atoi(maxAgeKeyVal[1])
			if err == nil {
				ttl = maxAge
				break
			}
		}
	}

	if ttl == 0 {
		ttl = 60
	}

	return time.Duration(ttl) * time.Second
}
