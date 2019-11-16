package dal

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nickhstr/goweb/cache"
	"github.com/nickhstr/goweb/logger"
)

var log = logger.New("dal")

// DefaultClient is the default http client for Fetch. Similar to http.DefaultClient, but sets
// a timeout.
var DefaultClient = &http.Client{
	Timeout: 15 * time.Second,
}

// FetchConfig holds all the information needed by dal.Fetch() to make a request.
type FetchConfig struct {
	*http.Request
	*http.Client
	// TTL is the time to live for a request's response in cache
	TTL time.Duration
	// CacheKey offers an optional key to use in place of the default key.
	// When defined, the response body of the request will be cached, no
	// matter the Method used.
	CacheKey string
	// Opt out of using the cache
	NoCache bool
}

// Validate verifies that the given FetchConfig has the basic pieces of information supplied.
func (fc *FetchConfig) Validate() error {
	if fc == nil {
		return errors.New("dal.FetchConfig: no FetchConfig provided")
	}
	if fc.Request == nil {
		return errors.New("dal.FetchConfig: no Request provided")
	}
	if fc.Client == nil {
		// Default to creating a new client versus using http.DefaultClient,
		// so that a timeout may be used without modifying http.DefaultClient
		fc.Client = DefaultClient
	}
	if fc.CacheKey == "" {
		fc.CacheKey = "dal:" + fc.URL.String()
	}
	if !fc.NoCache {
		switch fc.Method {
		case http.MethodGet:
			fc.NoCache = false
		default:
			fc.NoCache = true
		}
	}

	return nil
}

// Fetch makes a request, caches its response, and returns the response body.
// By default, the cache is only used for GET requests. For all
// other methods, fc.CacheKey must be defined.
func Fetch(fc *FetchConfig) ([]byte, error) {
	var (
		response []byte
		err      error
		ttl      = fc.TTL
	)

	err = fc.Validate()
	if err != nil {
		return response, err
	}

	fetchURL := fc.URL.String()
	start := time.Now()

	if !fc.NoCache {
		// Try to get response from cache
		cachedResp, err := cache.Get(fc.CacheKey)
		if err == nil {
			log.Info().
				Str("url", fetchURL).
				Str("response-time", time.Since(start).String()).
				Bool("redis", true).
				Msg("DAL request")

			return cachedResp, nil
		}
	}

	// Make request
	resp, err := fc.Client.Do(fc.Request)
	if err != nil {
		log.Error().
			Str("url", fetchURL).
			Err(err).
			Msg(err.Error())
		return response, err
	}

	if fc.TTL == 0 {
		ttl = TTLFromResponse(resp)
	}

	log.Info().
		Str("url", fetchURL).
		Str("response-time", time.Since(start).String()).
		Dur("ttl", ttl).
		Bool("redis", false).
		Msg("DAL request")

	body, err := ResponseBody(resp)
	if err != nil {
		log.Error().
			Str("url", fetchURL).
			Err(err).
			Msg(err.Error())
		return []byte{}, err
	}

	if !fc.NoCache {
		// Try to store response in cache
		cache.Set(fc.CacheKey, body, ttl)
	}

	return body, nil
}

/***** Fetch Wrappers *****/

// Delete is a convenience wrapper for Fetch.
func Delete(url string) ([]byte, error) {
	var err error

	r, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		err = fmt.Errorf("dal.Delete: http.NewRequest error: %w", err)
		return []byte{}, err
	}

	fc := &FetchConfig{
		Request: r,
		Client:  DefaultClient,
	}
	return Fetch(fc)
}

// Get is a convenience wrapper for Fetch.
func Get(url string) ([]byte, error) {
	var err error

	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		err = fmt.Errorf("dal.Get: http.NewRequest error: %w", err)
		return []byte{}, err
	}

	fc := &FetchConfig{
		Request: r,
		Client:  DefaultClient,
	}
	return Fetch(fc)
}

// Post is a convenience wrapper for Fetch.
func Post(url, contentType string, body io.Reader) ([]byte, error) {
	var err error

	r, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		err = fmt.Errorf("dal.Post: http.NewRequest error: %w", err)
		return []byte{}, err
	}

	r.Header.Set("Content-Type", contentType)

	fc := &FetchConfig{
		Request: r,
		Client:  DefaultClient,
	}
	return Fetch(fc)
}

/***** Utils *****/

// Used for ttlFromResponse
var maxAgeRegex = regexp.MustCompile(`max-age=\d+`)

// TTLFromResponse attempts to get a TTL value from a response's "cache-control" header,
// otherwise returning a default.
func TTLFromResponse(r *http.Response) time.Duration {
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

// ResponseBody returns the response's body data, with support for gzipped responses.
func ResponseBody(resp *http.Response) ([]byte, error) {
	var reader io.ReadCloser

	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, _ = gzip.NewReader(resp.Body)
	default:
		reader = resp.Body
	}

	// Make sure to always close response reader
	defer reader.Close()

	return ioutil.ReadAll(reader)
}
