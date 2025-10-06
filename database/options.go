package database

import (
	"time"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/types"
)

// DBOption defines a function that modifies a DBConfig.
type DBOption func(c DBConfig)

// NewDBOptions creates a new DBConfig with the given options.
func NewDBOptions[T DBConfig](opts ...DBOption) T {
	// Create a default configuration based on the type T
	var defaultConfig T
	switch any(defaultConfig).(type) {
	case *PostgresDBOptions:
		defaultConfig = any(&PostgresDBOptions{
			dsn:                "postgresql://user:password@localhost:5432/mydatabase",
			maxConns:           10,
			debugMode:          false,
			checkAliveInterval: constant.DatabaseCheckAliveInterval,
		}).(T)
	case *MySQLDBOptions:
		defaultConfig = any(&MySQLDBOptions{
			dsn:                "user:password@tcp(localhost:3306)/mydatabase",
			maxConns:           10,
			debugMode:          false,
			checkAliveInterval: constant.DatabaseCheckAliveInterval,
		}).(T)
	}

	// Apply options
	for _, opt := range opts {
		opt(defaultConfig)
	}
	return defaultConfig
}

// WithDSN sets the DSN for any database.
func WithDSN(dsn string) DBOption {
	return func(c DBConfig) {
		c.setDSN(dsn)
	}
}

// WithMaxConns sets the maximum number of connections for any database.
func WithMaxConns(maxConns int) DBOption {
	maxConns = helpers.GetMaxConns(maxConns)
	return func(c DBConfig) {
		c.setMaxConns(maxConns)
	}
}

// WithDebugMode enables or disables debug mode for any database.
func WithDebugMode(debug bool) DBOption {
	var logs = &log.Log{}
	if debug {
		logs = log.NewBasicLogger(false, true)
	}
	return func(c DBConfig) {
		c.setDebugMode(debug)
		c.setLogger(logs)
	}
}

// WithQueryProvider sets the name of the provider for retrieving SQL queries.
func WithQueryProvider(queryProvider types.DBType) DBOption {
	return func(c DBConfig) {
		c.setQueryProvider(queryProvider.String())
	}
}

// WithLogger sets the logger for any database.
func WithLogger(logger *log.Log) DBOption {
	return func(c DBConfig) {
		c.setLogger(logger)
	}
}

// WithCheckAliveInterval sets the interval for checking the health of the database connection.
func WithCheckAliveInterval(interval time.Duration) DBOption {
	return func(c DBConfig) {
		c.setCheckAliveInterval(interval)
	}
}
