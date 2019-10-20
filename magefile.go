// +build mage

package main

import (
	"fmt"
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/nickhstr/goweb/tools"
)

var Default = Install

const coverageOut = "coverage.out"

// Runs all tests and reports coverage.
func Coverage() error {
	var err error

	mg.Deps(CreateCoverage)
	fmt.Println("=============================== Coverage summary ===============================")

	err = sh.RunV("go", "tool", "cover", "-func", coverageOut)
	if err != nil {
		return err
	}

	fmt.Println("================================================================================")

	return err
}

// Opens coverage report in a browser.
func CoverageHtml() error {
	mg.Deps(CreateCoverage)
	fmt.Println("📊 Opening coverage report in browser...")

	return sh.Run("go", "tool", "cover", "-html", coverageOut)
}

// Runs all tests, and outputs a coverage report.
func CreateCoverage() error {
	fmt.Println("🏃 Running tests and creating coverage report...")
	os.Setenv(mg.VerboseEnv, "true")
	env := map[string]string{
		"GO_ENV": "test",
	}
	err := sh.RunWith(env, "go", "test", "-race", "-coverprofile", coverageOut, "./...")

	fmt.Println("✅ Done.")

	return err
}

// Installs all dependencies.
func Install() error {
	var (
		err       error
		toolsPath = "./internal/tools/tools.go"
	)

	fmt.Println("downloading dependencies")
	err = sh.RunV("go", "mod", "download")
	if err != nil {
		return err
	}

	f, err := tools.DepsFile(toolsPath)
	if err != nil {
		return err
	}

	defer f.Close()

	toolDeps, err := tools.ToInstall(f)
	if err != nil {
		return err
	}

	for _, dep := range toolDeps {
		fmt.Printf("installing %s\n", dep)
		err = sh.RunV("go", "install", dep)
		if err != nil {
			return err
		}
	}

	fmt.Println("👍 Done.")

	return err
}

// Only downloads Go dependencies. Other tools are installed separately.
func InstallCI() error {
	fmt.Println("downloading dependencies")
	err := sh.RunV("go", "mod", "download")

	return err
}

// Lints all files.
func Lint() error {
	var err error

	err = sh.RunV("golangci-lint", "run")
	if err != nil {
		return err
	}

	fmt.Println("✨ Done.")

	return err
}

// Runs all tests.
func Test() error {
	var err error

	fmt.Println("🏃 Running all Go tests...")
	// Set verbose env var to get test output
	// Needed until next release of Mage
	os.Setenv(mg.VerboseEnv, "true")
	env := map[string]string{
		"GO_ENV": "test",
	}
	err = sh.RunWith(env, "go", "test", "-race", "./...")
	if err != nil {
		return err
	}

	fmt.Println("✅ Done.")

	return err
}
