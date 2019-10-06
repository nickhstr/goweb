PROJECTNAME=$(shell basename "$(PWD)")

# First target, is the default command run if 'make' is invoked without any targets
all: help

## create-coverage: Outputs test coverage to 'coverage.out'
.PHONY: create-coverage
create-coverage:
	@echo "🏃 Running tests and creating coverage report..."
	@GO_ENV=test go test -race -coverprofile=coverage.out ./...
	@echo "✅ Done."

## coverage: Runs tests and reports coverage
.PHONY: coverage
coverage: create-coverage
	@echo "=============================== Coverage Summary ==============================="
	@go tool cover -func=coverage.out
	@echo "================================================================================"

## coverage-html: Runs tests and opens a browser window to visualize test coverage
.PHONY: coverage-html
coverage-html: create-coverage
	@echo "Opening coverage report in browser..."
	@go tool cover -html=coverage.out

## lint: Runs golangci-lint against entire project
.PHONY: lint
lint:
	@echo "🔍 Linting files..."
	@golangci-lint run
	@echo "✨ Done."

## install: Downloads all app dependencies
.PHONY: install
install:
	@go mod download
	@echo "👍 Done."

## test: Runs all tests
.PHONY: test
test:
	@echo "🏃 Running all tests..."
	GO_ENV=test go test -race ./...
	@echo "✅ Done."

## help: List available commands
.PHONY: help
help: Makefile
	@echo
	@echo " Choose a command to run in "$(PROJECTNAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo
