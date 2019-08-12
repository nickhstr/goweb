//+build mage

package main

import (
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const coverageOut = "coverage.out"

// Runs all tests and reports coverage.
func Coverage() error {
	var err error

	mg.Deps(CreateCoverage)
	sh.RunV("echo", "========== Coverage Summary ==========")
	err = sh.RunV("go", "tool", "cover", "-func", coverageOut)
	if err != nil {
		return err
	}

	return sh.RunV("echo", "======================================")
}

// Opens coverage report in a browser.
func CoverageHtml() error {
	mg.Deps(CreateCoverage)
	sh.RunV("echo", "🛠  Opening coverage report in browser...")

	return sh.Run("go", "tool", "cover", "-html", coverageOut)
}

// Runs all tests, and outputs a coverage report.
func CreateCoverage() error {
	sh.RunV("echo", "🏃 Running tests and creating coverage report...")
	os.Setenv(mg.VerboseEnv, "true")
	env := map[string]string{
		"GO_ENV": "test",
	}
	sh.RunWith(env, "go", "test", "-race", "-coverprofile", coverageOut, "./...")

	return sh.RunV("echo", "👍 Done.")
}

// Installs all dependencies.
func Install() error {
	var err error

	sh.RunV("echo", "🛠  Installing package dependencies...")
	err = sh.RunV("go", "mod", "download")
	if err != nil {
		return err
	}

	devDeps := []string{}

	for _, dep := range devDeps {
		err = sh.RunV("go", "install", dep)
		if err != nil {
			return err
		}
	}

	return sh.RunV("echo", "👍 Done.")
}

// Lints all files.
func Lint() error {
	var err error

	sh.RunV("echo", "🔍  Linting files...")
	err = sh.RunV("golangci-lint", "run")
	if err != nil {
		return err
	}

	return sh.RunV("echo", "👍 Done.")
}

// Runs all tests.
func Test() error {
	var err error

	sh.RunV("echo", "🏃 Running all Go tests...")
	os.Setenv(mg.VerboseEnv, "true")
	env := map[string]string{
		"GO_ENV": "test",
	}
	err = sh.RunWith(env, "go", "test", "-race", "./...")
	if err != nil {
		return err
	}

	return sh.RunV("echo", "👍 Done.")
}
