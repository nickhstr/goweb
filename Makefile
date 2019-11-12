PROJECTNAME ?= ${shell basename "${PWD}"}

# First target, is the default command run if 'make' is invoked without any targets
all: install

## coverage: Runs tests and reports coverage
.PHONY: coverage
coverage: create-coverage
	@echo "=============================== Coverage Summary ==============================="
	@go tool cover -func=coverage.out
	@echo "================================================================================"

## coverage-html: Runs tests and opens a browser window to visualize test coverage
.PHONY: coverage-html
coverage-html: create-coverage
	@echo "📊 Opening coverage report in browser..."
	@go tool cover -html=coverage.out

## create-coverage: Outputs test coverage to 'coverage.out'
.PHONY: create-coverage
create-coverage:
	@echo "🏃 Running tests and creating coverage report..."
	@GO_ENV=test go test -race -coverprofile=coverage.out ./...
	@echo "✅ Done."

## install: Downloads/installs all app dependencies
.PHONY: install
install:
	@echo "🚚 Downloading dependencies..."
	@go mod download
	@go mod vendor
	@echo "✨ Done."

## lint: Runs golangci-lint against entire project
.PHONY: lint
lint:
	@echo "🔍 Linting files..."
	go run vendor/github.com/golangci/golangci-lint/cmd/golangci-lint/main.go run ${flags}
	@echo "✨ Done."

## test: Runs all tests
.PHONY: test
test:
	${eval flags ?= -race}
	${eval packages ?= ./...}
	@echo "🏃 Running all Go tests..."
	GO_ENV=test go test ${flags} ${packages}
	@echo "✅ Done."

## test-watch: Runs tests and watches for changes
.PHONY: test-watch
test-watch:
	@echo "🏃 Running test watcher..."
	@go run vendor/github.com/cortesi/modd/cmd/modd/main.go --file=./internal/tools/modd.test.conf

## help: List available commands
.PHONY: help
help: Makefile
	@echo
	@echo " Choose a command to run in "${PROJECTNAME}":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/  /'
	@echo
