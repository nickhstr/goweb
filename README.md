# goweb
A collection of reusable Go packages for web development. The goal is to remain simple
and idiomatic, while adhering to the standard library's APIs when possible.

## Installation
`go get -u github.com/nickhstr/goweb`

## Features
* General middleware
* Configurable logger - built on github.com/rs/zerolog
* Router - routes registration with github.com/julienschmidt/httprouter
* Server - dns lookup caching and automatic port resolution
* Data access layer - request client with redis
* Newrelic - handler wrapper and custom logging, using github.com/newrelic/go-agent
* Environment variable helpers
