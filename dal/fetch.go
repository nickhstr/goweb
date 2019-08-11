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

// Used for getTTLFromResponse
var maxAgeRegex = regexp.MustCompile(`max-age=\d+`)

// FetchConfig holds all the information needed by dal.Fetch() to make a request.
type FetchConfig struct {
	url.URL
	// Method describes which request method to use.
	Method string
	// Body store the request's body.
	Body []byte
	http.Header
	// Query is the url.Values form of the request's query params.
	// These are to be encoded for use by URL's RawQuery.
	Query url.Values
	// TTL is the time to live for a request's response in cache
	TTL time.Duration
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
		return errors.New("No FetchConfig provided")
	}
	if fc.URL == (url.URL{}) ||
		fc.Scheme == "" ||
		fc.Host == "" ||
		fc.Path == "" {
		return errors.New("Invalid URL config provided")
	}
	if fc.Method == "" {
		fc.Method = http.MethodGet
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

// Fetch makes a request, and caches its response.
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
	cachedResp, err := cache.Get(fetchURL)
	if err == nil {
		log.Info().
			Str("url", fetchURL).
			Str("response-time", fmt.Sprintf("%v", time.Since(start))).
			Bool("redis", true).
			Msg("DAL request")

		return cachedResp, nil
	}

	req, err := fc.NewRequest()
	if err != nil {
		log.Error().
			Str("url", fetchURL).
			Err(err).
			Msg(err.Error())
		return response, err
	}

	// Make request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error().
			Str("url", fetchURL).
			Err(err).
			Msg(err.Error())
		return response, err
	}
	// Make sure to always close body
	defer resp.Body.Close()

	if fc.TTL == 0 {
		ttl = getTTLFromResponse(resp)
	}

	log.Info().
		Str("url", fetchURL).
		Str("response-time", fmt.Sprintf("%v", time.Since(start))).
		Dur("ttl", ttl).
		Bool("redis", false).
		Msg("DAL request")

	respBody, err := getResponseBody(resp)
	if err != nil {
		log.Error().
			Str("url", fetchURL).
			Err(err).
			Msg(err.Error())
		return []byte{}, err
	}

	// Try to store response in cache
	cache.Set(fetchURL, respBody, ttl)

	return respBody, nil
}

// Attempts to get a TTL value from a response's "cache-control" header.
// Otherwise, the given default TTL is used.
func getTTLFromResponse(r *http.Response) time.Duration {
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
		defaultTTL := time.Duration(60 * time.Second)
		ttl = defaultTTL
	}

	return ttl
}

func getResponseBody(resp *http.Response) ([]byte, error) {
	var (
		reader       io.ReadCloser
		responseBody []byte
		err          error
	)

	switch resp.Header.Get(http.CanonicalHeaderKey("content-encoding")) {
	case "gzip":
		reader, _ = gzip.NewReader(resp.Body)
		defer reader.Close()
	default:
		reader = resp.Body
	}

	responseBody, err = ioutil.ReadAll(reader)

	return responseBody, err
}
