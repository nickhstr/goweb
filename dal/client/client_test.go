package client

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestClientDo(t *testing.T) {
	gock.Intercept()
	defer gock.Off()

	tests := []struct {
		name         string
		client       *Client
		req          *http.Request
		expectedData []byte
	}{
		{
			"response body should be text",
			New(),
			&http.Request{
				URL: &url.URL{
					Scheme: "http",
					Host:   "foo.com",
					Path:   "/bar",
				},
				Method: http.MethodGet,
			},
			[]byte("baz"),
		},
		{
			"response body should be JSON",
			New(),
			&http.Request{
				URL: &url.URL{
					Scheme: "http",
					Host:   "foo.com",
					Path:   "/bar/post",
				},
				Method: http.MethodPost,
			},
			[]byte(`{"foo": "bar"}`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			bodyReader := bytes.NewBuffer(test.expectedData)

			gock.Intercept()
			req := gock.NewRequest().SetURL(test.req.URL)
			req.Method = test.req.Method
			exp := gock.NewMock(req, gock.NewResponse())
			gock.Register(exp)
			req.Reply(http.StatusOK).Body(bodyReader)

			resp, _ := test.client.Do(test.req)
			body, err := ioutil.ReadAll(resp.Body)

			assert.Nil(err)
			assert.Equal(test.expectedData, body)
		})
	}
}

func Test_ttlFromResponse(t *testing.T) {
	tests := []struct {
		name string
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
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(test.expected, ttlFromResponse(test.Response))
		})
	}
}
