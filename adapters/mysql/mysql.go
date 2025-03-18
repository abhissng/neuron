package mysql

import (
	"context"
	"database/sql"
	"fmt"

	// "log"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/database"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
)

// MySQL-specific implementation
type MySQLDB[T any] struct {
	conn     *sql.DB
	options  *database.MySQLDBOptions
	db       T
	factory  database.MySqlQueriesFactory[T]
	stopChan chan struct{}
}

// NewMySQLFactory creates a new MySQLDB factory with the given options.
func NewMySQLFactory[T any](options *database.MySQLDBOptions, factory database.GenericQueriesFactory[T]) *MySQLDB[T] {
	return &MySQLDB[T]{
		options: options,
		factory: func(db database.MySqlDBTX) T {
			return factory(db)
		},
		stopChan: make(chan struct{}),
	}
}

// Connect establishes a connection to the MySQL database.
func (m *MySQLDB[T]) Connect(ctx context.Context) error {
	conn, err := sql.Open("mysql", m.options.GetDSN())
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}
	// Set maximum open connections
	conn.SetMaxOpenConns(m.options.GetMaxConns())
	m.conn = conn

	if err := m.Ping(); err != nil {
		return err
	}

	if m.options.IsDebugMode() {
		logs := log.NewBasicLogger(helpers.IsProdEnvironment())
		logs.Info("Connected to MySQL")
		_ = logs.Sync()
	}
	return nil
}

// GetProviderDB returns the initialized query struct.
func (m *MySQLDB[T]) GetProviderDB() any {
	switch m.options.GetQueryProvider() {
	case constant.SQLCProvider.String():
		{
			return m.db
		}
	default:
		var defaultValue T // This will give you the zero value of T.
		return defaultValue
	}
}

// IsQueryProviderAvailable returns the if queryProvider is provided or not
func (m *MySQLDB[T]) IsQueryProviderAvailable() bool {
	return !helpers.IsEmpty(m.options.GetQueryProvider())
}

func (m *MySQLDB[T]) Ping() error {
	if err := m.conn.Ping(); err != nil {
		return err
	}
	if err := m.checkDatabaseExists(context.Background()); err != nil {
		return err
	}
	return nil
}

// checkDatabaseExists verifies that the specified database exists.
func (m *MySQLDB[T]) checkDatabaseExists(ctx context.Context) error {

	// Check if the database exists
	dbName, err := database.ExtractDBNameFromDSN(constant.MySQL, m.options.GetDSN()) // Extract the database name from the DSN
	if err != nil {
		return fmt.Errorf("database name cannot be blank: %w", err)
	}

	// Query the PostgreSQL system catalog to check for the database
	var exists bool
	err = m.QueryRow(ctx, constant.DatabaseExistQuery, dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("database '%s' does not exist", dbName)
	}
	return nil
}
