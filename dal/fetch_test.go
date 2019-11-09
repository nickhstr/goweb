package dal_test

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/nickhstr/goweb/dal"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestFetchConfigValidate(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		fc        *dal.FetchConfig
		shouldErr bool
	}{
		{
			nil,
			true,
		},
		{
			&dal.FetchConfig{},
			true,
		},
		{
			&dal.FetchConfig{
				Request: &http.Request{
					Method: http.MethodGet,
					URL: &url.URL{
						Scheme: "http",
						Host:   "foo.com",
						Path:   "/bar",
					},
				},
			},
			false,
		},
		{
			&dal.FetchConfig{
				Request: &http.Request{
					Method: http.MethodGet,
					URL: &url.URL{
						Scheme: "http",
						Host:   "foo.com",
						Path:   "/bar",
					},
				},
				Client: dal.DefaultClient,
			},
			false,
		},
	}

	for _, test := range tests {
		if test.shouldErr {
			assert.Error(test.fc.Validate(), "Validate should return an error: %t", test.shouldErr)
		} else {
			assert.Nil(test.fc.Validate(), "Validate should return an error: %t", test.shouldErr)
		}
	}
}

func TestFetch(t *testing.T) {
	assert := assert.New(t)
	gock.InterceptClient(dal.DefaultClient)
	defer gock.Off()
	defer gock.RestoreClient(dal.DefaultClient)

	tests := []struct {
		msg          string
		fc           *dal.FetchConfig
		expectedData []byte
	}{
		{
			"response body should be text",
			&dal.FetchConfig{
				Request: &http.Request{
					URL: &url.URL{
						Scheme: "http",
						Host:   "foo.com",
						Path:   "/bar",
					},
					Method: http.MethodGet,
				},
				NoCache: true,
			},
			[]byte("baz"),
		},
		{
			"response body should be JSON",
			&dal.FetchConfig{
				Request: &http.Request{
					URL: &url.URL{
						Scheme: "http",
						Host:   "foo.com",
						Path:   "/bar/post",
					},
					Method: http.MethodPost,
				},
			},
			[]byte(`{"foo": "bar"}`),
		},
	}

	for _, test := range tests {
		bodyReader := bytes.NewBuffer(test.expectedData)

		gock.Intercept()
		req := gock.NewRequest().SetURL(test.fc.URL)
		req.Method = test.fc.Method
		exp := gock.NewMock(req, gock.NewResponse())
		gock.Register(exp)
		req.Reply(http.StatusOK).Body(bodyReader)

		resp, err := dal.Fetch(test.fc)

		assert.Nil(err, test.msg)
		assert.Equal(test.expectedData, resp, test.msg)
	}
}

func TestDelete(t *testing.T) {
	assert := assert.New(t)
	gock.InterceptClient(dal.DefaultClient)
	defer gock.Off()
	defer gock.RestoreClient(dal.DefaultClient)

	tests := []struct {
		msg          string
		uri          string
		expectedData []byte
	}{
		{
			"response body should be text",
			"http://foo.com/bar",
			[]byte("baz"),
		},
		{
			"reponse body should be JSON",
			"http://foo.com/api/json",
			[]byte(`{"foo": "bar"}`),
		},
	}

	for _, test := range tests {
		bodyReader := bytes.NewBuffer(test.expectedData)

		gock.New(test.uri).
			Delete("/").
			Reply(http.StatusOK).
			Body(bodyReader)

		resp, err := dal.Delete(test.uri)

		assert.Nil(err, test.msg)
		assert.Equal(test.expectedData, resp, test.msg)
	}
}

func TestGet(t *testing.T) {
	assert := assert.New(t)
	gock.InterceptClient(dal.DefaultClient)
	defer gock.Off()
	defer gock.RestoreClient(dal.DefaultClient)

	tests := []struct {
		msg          string
		uri          string
		expectedData []byte
	}{
		{
			"response body should be text",
			"http://foo.com/bar",
			[]byte("baz"),
		},
		{
			"reponse body should be JSON",
			"http://foo.com/api/json",
			[]byte(`{"foo": "bar"}`),
		},
	}

	for _, test := range tests {
		bodyReader := bytes.NewBuffer(test.expectedData)

		gock.New(test.uri).
			Get("/").
			Reply(http.StatusOK).
			Body(bodyReader)

		resp, err := dal.Get(test.uri)

		assert.Nil(err, test.msg)
		assert.Equal(test.expectedData, resp, test.msg)
	}
}

func TestPost(t *testing.T) {
	assert := assert.New(t)
	gock.InterceptClient(dal.DefaultClient)
	defer gock.Off()
	defer gock.RestoreClient(dal.DefaultClient)

	tests := []struct {
		msg          string
		uri          string
		expectedData []byte
	}{
		{
			"response body should be text",
			"http://foo.com/bar",
			[]byte("baz"),
		},
		{
			"reponse body should be JSON",
			"http://foo.com/api/json",
			[]byte(`{"foo": "bar"}`),
		},
	}

	for _, test := range tests {
		resBodyReader := bytes.NewBuffer(test.expectedData)
		reqBodyReader := bytes.NewBuffer([]byte(`{"post": "message"}`))
		contentType := "application/json"

		gock.New(test.uri).
			Post("/").
			Reply(http.StatusOK).
			Body(resBodyReader)

		resp, err := dal.Post(test.uri, contentType, reqBodyReader)

		assert.Nil(err, test.msg)
		assert.Equal(test.expectedData, resp, test.msg)
	}
}

func TestTTLFromResponse(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		msg string
		*http.Response
		expected time.Duration
	}{
		{
			"should get TTL from response",
			&http.Response{
				Header: http.Header{
					"Cache-Control": []string{
						"max-age=900",
					},
				},
			},
			900 * time.Second,
		},
		{
			"should get TTL from response with multiple cache-control header values",
			&http.Response{
				Header: http.Header{
					"Cache-Control": []string{
						"must-revalidate",
						"public",
						"max-age=300",
					},
				},
			},
			300 * time.Second,
		},
		{
			"should use default TTL",
			&http.Response{},
			60 * time.Second,
		},
	}

	for _, test := range tests {
		assert.Equal(test.expected, dal.TTLFromResponse(test.Response), test.msg)
	}
}
