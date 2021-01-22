// Package mongodb provides an easy-to-use way to create a new
// Mongodb client. Configuration is limited, in the name of
// simplicity. More advanced clients should be created with the
// Mongodb driver instead.
package mongodb

import (
	"errors"
	"fmt"

	"github.com/newrelic/go-agent/v3/integrations/nrmongo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrClientConfig = errors.New("mongodb: bad config for client")

// ClientOptions provides some limited configuration for
// clients created with NewClient.
type ClientOptions struct {
	// PoolSize is the maximum number of connections allowed
	// to the server.
	PoolSize uint64
	// URI is the Mongodb connection URI.
	URI string
	// UseNewRelic allows clients to add New Relic database
	// segment monitoring.
	UseNewRelic bool
}

// New creates a new opinionated client.
func New(opts ClientOptions) (*mongo.Client, error) {
	clientOpts := options.Client().ApplyURI(opts.URI)

	if opts.UseNewRelic {
		clientOpts.SetMonitor(nrmongo.NewCommandMonitor(nil))
	}

	if opts.PoolSize != 0 {
		clientOpts.SetMaxPoolSize(opts.PoolSize)
	}

	client, err := mongo.NewClient(clientOpts)
	if err != nil {
		return client, fmt.Errorf("%w: %s", ErrClientConfig, err.Error())
	}

	return client, nil
}
