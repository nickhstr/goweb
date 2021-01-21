package sling

import (
	"github.com/dghubble/sling"
	"github.com/nickhstr/goweb/dal/client"
)

var dalClient = client.New()

// New returns a Sling instance, using the default DAL Client as
// its Doer.
func New() *sling.Sling {
	return sling.New().Doer(dalClient)
}

// NewWithClient returns a Sling instance, using the supplied DAL
// Client as its Doer.
func NewWithClient(c *client.Client) *sling.Sling {
	return sling.New().Doer(c)
}
