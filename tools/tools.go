package tools

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ToInstall gets the Go tools to install from an `internal/` tools.go file.
func ToInstall(rc io.Reader) ([]string, error) {
	var (
		tools          []string
		err            error
		cutsetLeft     = "\t_ "
		cutsetQuotes   = "\""
		depLinePattern = "\t_ \""
	)

	scanner := bufio.NewScanner(rc)

	for scanner.Scan() {
		line := scanner.Text()
		if err := scanner.Err(); err != nil {
			fmt.Println(err)

			return []string{}, err
		}

		// Only target lines with the name of the dependency, excluding Mage; Mage is
		// installed manually by the user
		if strings.Contains(line, depLinePattern) && !strings.Contains(line, "magefile/mage") {
			dep := strings.TrimLeft(line, cutsetLeft)
			dep = strings.Trim(dep, cutsetQuotes)
			tools = append(tools, dep)
		}
	}

	return tools, err
}

// DepsFile gets the os.File representation of a given relative path.
func DepsFile(relPath string) (*os.File, error) {
	var err error

	path, err := filepath.Abs(relPath)
	if err != nil {
		err = fmt.Errorf("failed to create absolute file path for %s:\n%w", relPath, err)
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		err = fmt.Errorf("cannot open dependencies file %s:\n%w", relPath, err)
		return nil, err
	}

	return f, nil
}
