package dal

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nickhstr/goweb/cache"
	"github.com/nickhstr/goweb/logger"
)

var log = logger.New("dal")

// FetchConfig holds all the information needed by dal.Fetch() to make a request.
type FetchConfig struct {
	url.URL
	// Method describes which request method to use.
	Method string
	// Body store the request's body.
	Body []byte
	http.Header
	// Optionally provide an http.Client
	// When not set, a new client is created in fc.validate()
	*http.Client
	// Query is the url.Values form of the request's query params.
	// These are to be encoded for use by URL's RawQuery.
	Query url.Values
	// TTL is the time to live for a request's response in cache
	TTL time.Duration
	// CacheKey offers an optional key to use in place of the default key.
	// When defined, the response body of the request will be cached, no
	// matter the Method used.
	CacheKey string
}

func (fc *FetchConfig) composeRawQuery() {
	// URL.RawQuery has higher priority over Query. Only set it if does not
	// already have a query.
	if fc.URL.RawQuery == "" {
		fc.URL.RawQuery = fc.Query.Encode()
	}
}

// String reassembles the URL and Query into a valid URL string.
func (fc *FetchConfig) String() string {
	fc.composeRawQuery()

	return fc.URL.String()
}

// Verifies that the given FetchConfig has the basic pieces of information supplied.
func (fc *FetchConfig) validate() error {
	if fc == nil {
		return errors.New("no FetchConfig provided")
	}
	if fc.URL == (url.URL{}) {
		return errors.New("empty URL config provided")
	}
	if fc.Scheme == "" {
		return errors.New("no Scheme provided")
	}
	if fc.Host == "" {
		return errors.New("no Host provided")
	}
	if fc.Method == "" {
		fc.Method = http.MethodGet
	}
	if fc.CacheKey == "" && fc.Method == http.MethodGet {
		// Only set CacheKey when not already provided and the method is GET
		// Otherwise, leave empty to avoid using the cache
		fc.CacheKey = "dal:" + fc.String()
	}
	if fc.Client == nil {
		// Default to creating a new client versus using http.DefaultClient,
		// so that a timeout may be used without modifying http.DefaultClient
		fc.Client = &http.Client{
			Timeout: 15 * time.Second,
		}
	}

	return nil
}

// NewRequest creates an http.Request from a FetchConfig.
func (fc *FetchConfig) NewRequest() (*http.Request, error) {
	var (
		req *http.Request
		err error
	)

	// Create body ready for requests with a non nil body
	reqBody := bytes.NewBuffer(fc.Body)
	req, err = http.NewRequest(fc.Method, fc.String(), reqBody)
	if err != nil {
		return req, err
	}

	// Canonicalize the headers
	if fc.Header != nil {
		for key, val := range fc.Header {
			key = http.CanonicalHeaderKey(key)
			req.Header[key] = val
		}
	}

	return req, nil
}

// Fetch makes a request, caches its response, and returns the response body.
// By default, the cache is only used for GET requests. For all
// other methods, fc.CacheKey must be defined
func Fetch(fc FetchConfig) ([]byte, error) {
	var (
		response []byte
		err      error
		ttl      = fc.TTL
	)

	err = fc.validate()
	if err != nil {
		return response, err
	}

	fetchURL := fc.String()
	start := time.Now()

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

	// Create request
	req, err := fc.NewRequest()
	if err != nil {
		log.Error().
			Str("url", fetchURL).
			Err(err).
			Msg(err.Error())
		return response, err
	}
	// Make request
	resp, err := fc.Client.Do(req)
	if err != nil {
		log.Error().
			Str("url", fetchURL).
			Err(err).
			Msg(err.Error())
		return response, err
	}

	if fc.TTL == 0 {
		ttl = ttlFromResponse(resp)
	}

	log.Info().
		Str("url", fetchURL).
		Str("response-time", time.Since(start).String()).
		Dur("ttl", ttl).
		Bool("redis", false).
		Msg("DAL request")

	body, err := responseBody(resp)
	if err != nil {
		log.Error().
			Str("url", fetchURL).
			Err(err).
			Msg(err.Error())
		return []byte{}, err
	}

	// Try to store response in cache
	cache.Set(fc.CacheKey, body, ttl)

	return body, nil
}

/***** Fetch Wrappers *****/

// Delete is a convenience wrapper for Fetch.
func Delete(rawurl string) ([]byte, error) {
	var err error

	u, err := url.Parse(rawurl)
	if err != nil {
		err := fmt.Errorf("dal.Delete URL parse error: %w", err)
		return []byte{}, err
	}

	fc := FetchConfig{
		URL:    *u,
		Method: http.MethodDelete,
	}
	return Fetch(fc)
}

// Get is a convenience wrapper for Fetch.
func Get(rawurl string) ([]byte, error) {
	var err error

	u, err := url.Parse(rawurl)
	if err != nil {
		err := fmt.Errorf("dal.Get URL parse error: %w", err)
		return []byte{}, err
	}

	fc := FetchConfig{
		URL:    *u,
		Method: http.MethodGet,
	}
	return Fetch(fc)
}

// Head is a convenience wrapper for Fetch.
func Head(rawurl string) ([]byte, error) {
	var err error

	u, err := url.Parse(rawurl)
	if err != nil {
		err := fmt.Errorf("dal.Head URL parse error: %w", err)
		return []byte{}, err
	}

	fc := FetchConfig{
		URL:    *u,
		Method: http.MethodHead,
	}
	return Fetch(fc)
}

// Patch is a convenience wrapper for Fetch.
func Patch(rawurl string, body []byte) ([]byte, error) {
	var err error

	u, err := url.Parse(rawurl)
	if err != nil {
		err := fmt.Errorf("dal.Patch URL parse error: %w", err)
		return []byte{}, err
	}

	fc := FetchConfig{
		URL:    *u,
		Method: http.MethodPatch,
		Body:   body,
	}
	return Fetch(fc)
}

// Post is a convenience wrapper for Fetch.
func Post(rawurl string, body []byte) ([]byte, error) {
	var err error

	u, err := url.Parse(rawurl)
	if err != nil {
		err := fmt.Errorf("dal.Post URL parse error: %w", err)
		return []byte{}, err
	}

	fc := FetchConfig{
		URL:    *u,
		Method: http.MethodPost,
		Body:   body,
	}
	return Fetch(fc)
}

// Put is a convenience wrapper for Fetch.
func Put(rawurl string, body []byte) ([]byte, error) {
	var err error

	u, err := url.Parse(rawurl)
	if err != nil {
		err := fmt.Errorf("dal.Put URL parse error: %w", err)
		return []byte{}, err
	}

	fc := FetchConfig{
		URL:    *u,
		Method: http.MethodPut,
		Body:   body,
	}
	return Fetch(fc)
}

/***** Utils *****/

// Used for ttlFromResponse
var maxAgeRegex = regexp.MustCompile(`max-age=\d+`)

// Attempts to get a TTL value from a response's "cache-control" header.
// Otherwise, the given default TTL is used.
func ttlFromResponse(r *http.Response) time.Duration {
	var ttl time.Duration

	headerKey := http.CanonicalHeaderKey("cache-control")
	cacheControlValues := r.Header[headerKey]

	for _, val := range cacheControlValues {
		match := maxAgeRegex.FindString(val)

		if match != "" {
			maxAgeKeyVal := strings.Split(match, "=")
			maxAge, err := strconv.Atoi(maxAgeKeyVal[1])
			if err != nil {
				log.Error().
					Err(err).
					Msgf("Failed to convert %s to an integer", maxAgeKeyVal[1])
			}
			ttl = time.Duration(maxAge) * time.Second
			break
		}
	}

	if ttl == 0 {
		defaultTTL := 60 * time.Second
		ttl = defaultTTL
	}

	return ttl
}

// Returns the response's body, with support for gzipped responses.
func responseBody(resp *http.Response) ([]byte, error) {
	var reader io.ReadCloser

	switch resp.Header.Get(http.CanonicalHeaderKey("content-encoding")) {
	case "gzip":
		reader, _ = gzip.NewReader(resp.Body)
	default:
		reader = resp.Body
	}

	// Make sure to always close response reader
	defer reader.Close()

	return ioutil.ReadAll(reader)
}
