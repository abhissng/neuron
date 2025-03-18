package database

import (
	"context"
	"time"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// PostgresDBOptions struct for PostgreSQL configuration.
type PostgresDBOptions struct {
	dsn                string
	queryProvider      string
	maxConns           int
	debugMode          bool
	log                *log.Log
	checkAliveInterval time.Duration
}

// PostgresDBTX is the common interface for executing database queries using SQLC.
type PostgresDBTX interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

// PostgresQueriesFactory is a generic factory function for creating `Queries` instances.
type PostgresQueriesFactory[T any] func(db PostgresDBTX) T

// GetDSN returns the DSN for PostgreSQL.
func (p *PostgresDBOptions) GetDSN() string {
	return p.dsn
}

// GetMaxConns returns the maximum number of connections for PostgreSQL.
func (p *PostgresDBOptions) GetMaxConns() int {
	return p.maxConns
}

// IsDebugMode returns whether debug mode is enabled for PostgreSQL.
func (p *PostgresDBOptions) IsDebugMode() bool {
	return p.debugMode
}

// GetLogger returns the Logger for PostgreSQL.
func (p *PostgresDBOptions) GetLogger() *log.Log {
	return p.log
}

// GetQueryProvider returns the QueryProvider for PostgreSQL.
func (p *PostgresDBOptions) GetQueryProvider() string {
	return p.queryProvider
}

// GetCheckAliveInterval returns the interval for checking the health of the database connection.
func (p *PostgresDBOptions) GetCheckAliveInterval() time.Duration {
	return p.checkAliveInterval
}

// setDSN sets the Data Source Name (DSN) for the PostgreSQL connection.
// The DSN string should be in the format:
// "postgresql://user:password@host:port/dbname?options
func (p *PostgresDBOptions) setDSN(dsn string) {
	p.dsn = dsn
}

// setMaxConns sets the maximum number of open connections to the PostgreSQL database.
func (p *PostgresDBOptions) setMaxConns(maxConns int) {
	p.maxConns = maxConns
}

// setDebugMode enables or disables debug mode for the PostgreSQL client.
// When enabled, debug messages will be logged.
func (p *PostgresDBOptions) setDebugMode(debug bool) {
	p.debugMode = debug
}

// setQueryProvider sets the name of the provider for retrieving SQL queries.
// This can be used to switch between different query sources (e.g., raw, sqlc).
func (p *PostgresDBOptions) setQueryProvider(queryProvider string) {
	p.queryProvider = queryProvider
}

// setLogger sets the logger for the PostgreSQL client to use for logging messages.
func (p *PostgresDBOptions) setLogger(logger *log.Log) {
	p.log = logger
}

// setCheckAliveInterval sets the interval for checking the health of the database connection.
func (p *PostgresDBOptions) setCheckAliveInterval(interval time.Duration) {
	p.checkAliveInterval = interval
}
