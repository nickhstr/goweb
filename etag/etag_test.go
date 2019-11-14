package etag_test

import (
	"testing"

	"github.com/nickhstr/goweb/etag"
	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		msg         string
		data        []byte
		weak        bool
		hash        string
		shouldEqual bool
	}{
		{
			"given identical input, Generate should return same output",
			[]byte("some text"),
			true,
			etag.Generate([]byte("some text"), true),
			true,
		},
		{
			"given different input, Generate should return different output",
			[]byte("some text"),
			false,
			etag.Generate([]byte("different text"), false),
			false,
		},
		{
			"given input of same length but different content, Generate should return different output",
			[]byte("some text"),
			false,
			etag.Generate([]byte("more text"), false),
			false,
		},
	}

	for _, test := range tests {
		if test.shouldEqual {
			assert.Equal(test.hash, etag.Generate(test.data, test.weak), test.msg)
		} else {
			assert.NotEqual(test.hash, etag.Generate(test.data, test.weak), test.msg)
		}
	}
}
