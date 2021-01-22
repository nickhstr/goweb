package nosql

import (
	"context"

	"github.com/nickhstr/goweb/logger"
)

// NoopDB serves as a no-op database, used
// primarily for composing mock databases in
// unit tests.
type NoopDB struct {
	name string
	log  logger.Logger
}

func NewNoopDB(name string) NoopDB {
	return NoopDB{name, logger.New("nosql")}
}

func (n NoopDB) Aggregate(ctx context.Context, collName string, pipeline interface{}, results interface{}) error {
	n.log.Debug().Str("operation", "Aggregate").Msg("noop operation")
	return nil
}

func (n NoopDB) DeleteOne(ctx context.Context, collName, id string) error {
	n.log.Debug().Str("operation", "DeleteOne").Msg("noop operation")
	return nil
}

func (n NoopDB) FindByID(ctx context.Context, collName, id string, result interface{}) error {
	n.log.Debug().Str("operation", "FindByID").Msg("noop operation")
	return nil
}

func (n NoopDB) FindOne(ctx context.Context, collName string, filter interface{}, result interface{}) error {
	n.log.Debug().Str("operation", "FindOne").Msg("noop operation")
	return nil
}

func (n NoopDB) FindAll(ctx context.Context, collName string, start, limit int64, results interface{}) error {
	n.log.Debug().Str("operation", "FindAll").Msg("noop operation")
	return nil
}

func (n NoopDB) FindMany(ctx context.Context, collName string, start, limit int64, filter interface{}, results interface{}) error {
	n.log.Debug().Str("operation", "FindMany").Msg("noop operation")
	return nil
}

func (n NoopDB) InsertByID(ctx context.Context, collName, id string, data interface{}) error {
	n.log.Debug().Str("operation", "InsertOne").Msg("noop operation")
	return nil
}

func (n NoopDB) ReplaceByID(ctx context.Context, collName, id string, data interface{}) error {
	n.log.Debug().Str("operation", "ReplaceOne").Msg("noop operation")
	return nil
}

// sanity check for satisfaction of Storer interface
var _ Storer = NewNoopDB("interface-test")
