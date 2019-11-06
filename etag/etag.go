package etag

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

func hash(p []byte) string {
	h := sha1.Sum(p)
	return hex.EncodeToString(h[:])
}

// Generate an etag for given byte slice. Can be set as weak with second boolean parameter.
func Generate(p []byte, weak bool) string {
	tag := fmt.Sprintf("\"%d-%s\"", len(p), hash(p))
	if weak {
		tag = "W/" + tag
	}

	return tag
}
