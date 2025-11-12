package database

import (
	"context"
	"fmt"
	"time"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
)

// TODO: what if there are muliple database that needs to be connected how to handle that
// at the same time mongo also needs to be connected with the same system is it currently possible no ?
// a new context needs to be created with the current implementation

// Database defines a common interface for different database backends.
type Database interface {
	Connect(ctx context.Context) error
	Close() error
	Query(ctx context.Context, query string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) Row
	Exec(ctx context.Context, query string, args ...any) (ExecResult, error)
	BeginTransaction(ctx context.Context) (Transaction, error)
	IsDebugQuery(query string, args ...any)
	RowsToSlice(rows Rows) ([][]string, error)
	Ping() error
	GetProviderDB() any // GetDB should return the underlying query or database object (e.g., *sqlc.Queries)
	IsQueryProviderAvailable() bool
	FetchStopChannel() <-chan struct{}
	FetchCheckAliveInterval() time.Duration
	StartMonitor()
	StopMonitor()
	GetLogger() *log.Log
}

// DBOptions defines a generic interface for database-specific options.
type DBOptions any

// GenericQueriesFactory defines a generic interface for database queries.
type GenericQueriesFactory[Q any] func(db any) Q

// NewDatabase creates and initializes a new database instance using generics.
func NewDatabase[T Database, O DBOptions, Q any](
	dbFactory func(O, GenericQueriesFactory[Q]) T,
	queriesFactory GenericQueriesFactory[Q],
	options O,
) (Database, error) {
	// Initialize the database instance with the provided factory.
	db := dbFactory(options, queriesFactory)

	// Establish the connection and initialize the query struct.
	if err := db.Connect(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}
	// go MonitorDB[Database](context.Background(), db)
	db.StartMonitor()

	return db, nil
}

// Transaction defines a common interface for database transactions.
type Transaction interface {
	Query(ctx context.Context, query string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) Row
	Exec(ctx context.Context, query string, args ...any) (ExecResult, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	IsDebugQuery(query string, args ...any)
	RowsToSlice(rows Rows) ([][]string, error)
	GetTx() any
}

// Rows defines a common interface for query results, providing methods for interacting with a set of rows returned from a database query.
type Rows interface {
	// Close closes the resource associated with the set of rows.
	Close() error
	// Next advances to the next row in the result set. Returns true if there is a next row, false otherwise.
	Next() bool
	// Scan reads the columns in the current row into the provided values.
	Scan(dest ...any) error
	// Err returns any error encountered during the iteration over the rows.
	Err() error
}

// ExtendedRows extends the Rows interface to include a method for retrieving column names.
type ExtendedRows interface {
	Rows
	Columns() ([]string, error)
}

// Row defines a common interface for a single row result from a database query.
type Row interface {
	// Scan reads the columns in the row into the provided values.
	Scan(dest ...any) error
}

// ExecResult interface defines methods for retrieving information about the result of an executed SQL statement.
type ExecResult interface {
	// RowsAffected returns the number of rows affected by the executed statement.
	RowsAffected() int64
	// LastInsertId returns the ID of the last inserted row, if applicable.
	LastInsertId() int64
}

// DBConfig interface defines the common configuration for any database.
type DBConfig interface {
	GetDSN() string
	GetMaxConns() int
	IsDebugMode() bool
	GetQueryProvider() string
	GetLogger() *log.Log
	GetCheckAliveInterval() time.Duration
	setDSN(string)
	setMaxConns(int)
	setDebugMode(bool)
	setQueryProvider(string)
	setLogger(*log.Log)
	setCheckAliveInterval(time.Duration)
}

// MonitorDB continuously monitors the database connection.
func MonitorDB[T Database](ctx context.Context, dbInstance T) {
	ticker := time.NewTicker(dbInstance.FetchCheckAliveInterval())
	defer ticker.Stop()

	logger := dbInstance.GetLogger()
	if logger == nil {
		logger = log.NewBasicLogger(helpers.IsProdEnvironment(), true)
	}
	defer func() {
		_ = logger.Sync()
	}()

	for {
		select {
		case <-ctx.Done():
			logger.Info(constant.SystemStopped, log.Any("message", "Stopping database monitor..."))
			return
		case <-dbInstance.FetchStopChannel():
			logger.Info(constant.SystemStopped, log.Any("message", "Received stop signal, stopping database monitor..."))
			return
		case <-ticker.C:
			if err := dbInstance.Ping(); err != nil {
				logger.Error(constant.SystemWarning, log.Any("message", log.Sprintf("Database connection failed: %s", err.Error())))
				//pgxpool will connect automaticaaly if the database disconnects
				// if err := dbInstance.Connect(ctx); err != nil {
				// 	logger.Error(constant.SystemError, log.Any("message", log.Sprintf("failed to connect to the database: %s", err.Error())))
				// }
			}
		}
	}
}
