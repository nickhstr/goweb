package etag_test

import (
	"testing"

	"github.com/nickhstr/goweb/etag"
	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name        string
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
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			if test.shouldEqual {
				assert.Equal(test.hash, etag.Generate(test.data, test.weak))
			} else {
				assert.NotEqual(test.hash, etag.Generate(test.data, test.weak))
			}
		})
	}
}

// Create variable outside of BenchmarkGenerate's scope to avoid
// compiler optimizations artificially lowering the run time
// of the benchmark.
var benchedTag string

func BenchmarkGenerate(b *testing.B) {
	b.ReportAllocs()

	var hash string

	content := []byte(`let's see how fast Generate can create a hash...
		Here's a line.
		And here's another line.
		Just testing some content here, don't mind me.
	`)

	for n := 0; n < b.N; n++ {
		hash = etag.Generate(content, false)
	}

	benchedTag = hash
}
