package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/abhissng/neuron/adapters/log"
)

// MySQLDBOptions struct for MySQL configuration.
type MySQLDBOptions struct {
	dsn                string
	queryProvider      string
	maxConns           int
	debugMode          bool
	log                *log.Log
	checkAliveInterval time.Duration
	startMonitor       bool
}

// MySqlDBTX is the common interface for executing database queries using SQLC.
type MySqlDBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

// MySqlQueriesFactory is a generic factory function for creating `Queries` instances.
type MySqlQueriesFactory[T any] func(db MySqlDBTX) T

// GetDSN returns the DSN for MySQL.
func (m *MySQLDBOptions) GetDSN() string {
	return m.dsn
}

// GetMaxConns returns the maximum number of connections for MySQL.
func (m *MySQLDBOptions) GetMaxConns() int {
	return m.maxConns
}

// IsDebugMode returns whether debug mode is enabled for MySQL.
func (m *MySQLDBOptions) IsDebugMode() bool {
	return m.debugMode
}

// GetLogger returns the Logger for MySQL.
func (m *MySQLDBOptions) GetLogger() *log.Log {
	return m.log
}

// GetQueryProvider returns the QueryProvider for MySQL.
func (m *MySQLDBOptions) GetQueryProvider() string {
	return m.queryProvider
}

// GetCheckAliveInterval returns the interval for checking the health of the database connection.
func (m *MySQLDBOptions) GetCheckAliveInterval() time.Duration {
	return m.checkAliveInterval
}

// GetStartMonitor returns whether health monitoring should start automatically.
func (m *MySQLDBOptions) GetStartMonitor() bool {
	return m.startMonitor
}

// setDSN sets the Data Source Name (DSN) for the MySQL connection.
// The DSN string should be in the format:
// user:password@tcp(host:port)/dbname?options
func (m *MySQLDBOptions) setDSN(dsn string) {
	m.dsn = dsn
}

// setMaxConns sets the maximum number of open connections to the MySQL database.
func (m *MySQLDBOptions) setMaxConns(maxConns int) {
	m.maxConns = maxConns
}

// setDebugMode enables or disables debug mode for the MySQL client.
// When enabled, debug messages will be logged.
func (m *MySQLDBOptions) setDebugMode(debug bool) {
	m.debugMode = debug
}

// setQueryProvider sets the name of the provider for retrieving SQL queries.
// This can be used to switch between different query sources (e.g., raw, sqlc).
func (m *MySQLDBOptions) setQueryProvider(queryProvider string) {
	m.queryProvider = queryProvider
}

// setCheckAliveInterval sets the interval for checking the health of the database connection.
func (m *MySQLDBOptions) setCheckAliveInterval(interval time.Duration) {
	m.checkAliveInterval = interval
}

// setStartMonitor toggles whether health monitoring should start automatically.
func (m *MySQLDBOptions) setStartMonitor(start bool) {
	m.startMonitor = start
}

// setLogger sets the logger for the MySQL client to use for logging messages.
func (m *MySQLDBOptions) setLogger(logger *log.Log) {
	m.log = logger
}
