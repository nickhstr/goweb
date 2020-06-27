# goweb
A collection of reusable Go packages for web development. The goal is to remain simple
and idiomatic, while adhering to the standard library's APIs when possible.

## Installation
`go get github.com/nickhstr/goweb`

Note: Go modules are the only supported dependency tool.


## Features
* General middleware
* Configurable logger - built on github.com/rs/zerolog
* Router - routes registration with github.com/go-chi/chi
* Server - dns lookup caching and automatic port resolution
* Data access layer - request client
* Cache - a key-value cache, using Redis
* Newrelic - handler wrapper and custom logging, using github.com/newrelic/go-agent
* Environment variable helpers
* Mongodb helpers

## Contributors

### Setup
Install dependencies:

`make` or `make install`

### Workflow
Run `make help` to view the available common tasks, such as linting, testing, etc.
