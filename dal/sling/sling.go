package sling

import (
	"github.com/dghubble/sling"
	"github.com/nickhstr/goweb/dal/client"
)

var dalClient = client.New()

// New returns a Sling instance, using the DAL's Client as its Doer.
func New() *sling.Sling {
	return sling.New().Doer(dalClient)
}
