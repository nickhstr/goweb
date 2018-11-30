package etag

import (
	"crypto/sha1"
	"fmt"
)

func getHash(p []byte) string {
	return fmt.Sprintf("%x", sha1.Sum(p))
}

// Generate an etag for given byte slice. Can be set as weak with second boolean parameter.
func Generate(p []byte, weak bool) string {
	tag := fmt.Sprintf("\"%d-%s\"", len(p), getHash(p))
	if weak {
		tag = "W/" + tag
	}

	return tag
}
