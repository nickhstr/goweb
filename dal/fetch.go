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

	"github.com/nickhstr/goweb/cache" // nolint: gotype
	"github.com/nickhstr/goweb/logger"
)

var log = logger.New(nil).With().Str("namespace", "dal").Logger()

// FetchConfig holds all the information needed by dal.Fetch() to make a request.
type FetchConfig struct {
	*url.URL
	// Method describes which request method to use.
	Method string
	// Body store the request's body.
	Body []byte
	http.Header
	// Query is the url.Values form of the request's query params.
	// These are to be encoded for use by URL's RawQuery.
	Query url.Values
}

// Fetch makes a request, and is responsible for caching the response data.
func Fetch(fc *FetchConfig) ([]byte, error) {
	var (
		response []byte
		err      error
	)

	err = validateFetchConfig(fc)
	if err != nil {
		return response, err
	}

	fc.composeRawQuery()
	fetchURL := fc.URL.String()
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

	req, err := createRequest(fc, fetchURL)
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

	ttl := time.Duration(getTTLFromResponse(resp, 60)) * time.Second
	log.Info().
		Str("url", fetchURL).
		Str("response-time", fmt.Sprintf("%v", time.Since(start))).
		Dur("ttl", ttl).
		Bool("redis", false).
		Msg("DAL request")

	var reader io.ReadCloser

	switch resp.Header.Get(http.CanonicalHeaderKey("content-encoding")) {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		defer reader.Close()
	default:
		reader = resp.Body
	}

	response, err = ioutil.ReadAll(reader)

	if err != nil {
		log.Error().
			Str("url", fetchURL).
			Err(err).
			Msg(err.Error())
		return []byte{}, err
	}

	// Try to store response in cache
	cache.Set(fetchURL, response, ttl)

	return response, nil
}

func validateFetchConfig(fc *FetchConfig) error {
	if fc == nil {
		return errors.New("No FetchConfig provided")
	}
	if fc.URL == nil {
		return errors.New("No URL config provided")
	}
	if fc.Method == "" {
		fc.Method = http.MethodGet
	}

	return nil
}

func getTTLFromResponse(r *http.Response, defaultTTL int) int {
	var ttl int

	cacheControl := r.Header[http.CanonicalHeaderKey("cache-control")]
	maxAgeRegex, err := regexp.Compile(`max-age=\d+`)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to compile getTTLFromResponse()'s regex")
		return ttl
	}

	for _, val := range cacheControl {
		match := maxAgeRegex.FindString(val)

		if match != "" {
			maxAgeKeyVal := strings.Split(match, "=")
			ttl, err = strconv.Atoi(maxAgeKeyVal[1])
			if err != nil {
				log.Error().
					Err(err).
					Msgf("Failed to convert %s to an integer", maxAgeKeyVal[1])
			}
			break
		}
	}

	if ttl == 0 {
		ttl = defaultTTL
	}

	return ttl
}

func (fc *FetchConfig) composeRawQuery() {
	// URL.RawQuery has higher priority over Query. Only set it if does not
	// already have a query.
	if fc.URL.RawQuery == "" {
		fc.URL.RawQuery = fc.Query.Encode()
	}
}

func createRequest(fc *FetchConfig, fetchURL string) (*http.Request, error) {
	var (
		req *http.Request
		err error
	)

	// Create body ready for requests with a non nil body
	reqBody := bytes.NewBuffer(fc.Body)
	req, err = http.NewRequest(fc.Method, fetchURL, reqBody)
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