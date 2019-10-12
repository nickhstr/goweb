// +build tools

// "tools" is home for all installable tooling dependencies to be versioned.
package tools

import (
	_ "github.com/cortesi/modd/cmd/modd"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/magefile/mage"
)
