package tools_test

import (
	"strings"
	"testing"

	"github.com/nickhstr/goweb/tools"
	"github.com/stretchr/testify/assert"
)

func TestToInstall(t *testing.T) {
	assert := assert.New(t)
	mockTools := `// +build tools

package tools

import (
	_ "some/dummy/project"
	_ "a/versioned/package/v2"
	_ "lib/_with_underscore_"
	_ "test.com/very/useful-lib"
	_ "github.com/magefile/mage"
)

`
	mockToolsReader := strings.NewReader(mockTools)
	expected := []string{
		"some/dummy/project",
		"a/versioned/package/v2",
		"lib/_with_underscore_",
		"test.com/very/useful-lib",
	}
	actual, err := tools.ToInstall(mockToolsReader)

	assert.Equal(expected, actual, "tool dependencies are stored in slice")
	assert.Nil(err, "ToInstall() has no errors")
}
