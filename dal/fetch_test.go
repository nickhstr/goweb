package dal_test

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"

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
				Request: nil,
				Client:  dal.DefaultClient,
			},
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

func TestGet(t *testing.T) {
	assert := assert.New(t)
	gock.InterceptClient(dal.DefaultClient)
	defer gock.Off()
	defer gock.RestoreClient(dal.DefaultClient)

	tests := []struct {
		msg    string
		uri    string
		path   string
		status int
		body   []byte
	}{
		{
			"response body should be text",
			"http://foo.com",
			"/bar",
			http.StatusOK,
			[]byte("baz"),
		},
		{
			"reponse body should be JSON",
			"http://foo.com",
			"/api/json",
			http.StatusOK,
			[]byte(`{"foo": "bar"}`),
		},
	}

	for _, test := range tests {
		bodyReader := bytes.NewBuffer(test.body)

		gock.New(test.uri).
			Get(test.path).
			Reply(test.status).
			Body(bodyReader)

		resp, err := dal.Get(test.uri + test.path)

		assert.Nil(err, test.msg)
		assert.Equal(test.body, resp, test.msg)
	}
}
