package db

import (
	"context"
	"fmt"

	sink "github.com/streamingfast/substreams-sink"
)

type UnknownDriverError struct {
	Driver string
}

// Error returns a formatted string description.
func (e UnknownDriverError) Error() string {
	return fmt.Sprintf("unknown database driver: %s", e.Driver)
}

type dialect interface {
	GetCreateCursorQuery(schema string, withPostgraphile bool) string
	GetCreateHistoryQuery(schema string, withPostgraphile bool) string
	ExecuteSetupScript(ctx context.Context, l *Loader, schemaSql string) error
	DriverSupportRowsAffected() bool
	GetUpdateCursorQuery(table, moduleHash string, cursor *sink.Cursor, block_num uint64, block_id string) string
	ParseDatetimeNormalization(value string) string
	Flush(tx Tx, ctx context.Context, l *Loader, outputModuleHash string, lastFinalBlock uint64) (int, error)
	Revert(tx Tx, ctx context.Context, l *Loader, lastValidFinalBlock uint64) error
	OnlyInserts() bool
	AllowPkDuplicates() bool
	CreateUser(tx Tx, ctx context.Context, l *Loader, username string, password string, database string, readOnly bool) error
}

var driverDialect = map[string]dialect{
	"*pq.Driver":            postgresDialect{},   // github.com/lib/pq
	"*clickhouse.stdDriver": clickhouseDialect{}, // github.com/clickhouse-go/v2
}
