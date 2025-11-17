package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/database"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
)

// MySQLDB provides a generic MySQL database adapter with connection management.
// It supports query factories, health monitoring, and connection pooling.
type MySQLDB[T any] struct {
	conn               *sql.DB
	options            *database.MySQLDBOptions
	db                 T
	factory            database.MySqlQueriesFactory[T]
	stopChan           chan struct{}
	checkAliveInterval time.Duration
	monitorCancel      context.CancelFunc
	monitorRunning     bool
	monitorMu          sync.Mutex
}

// NewMySQLFactory creates a new MySQLDB instance with the specified options and query factory.
// The factory function is used to create query objects for database operations.
func NewMySQLFactory[T any](options *database.MySQLDBOptions, factory database.GenericQueriesFactory[T]) *MySQLDB[T] {
	return &MySQLDB[T]{
		options: options,
		factory: func(db database.MySqlDBTX) T {
			return factory(db)
		},
		stopChan: make(chan struct{}),
	}
}

// Connect establishes a connection to the MySQL database using the configured options.
// It sets up connection pooling, performs health checks, and starts monitoring if enabled.
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
	m.checkAliveInterval = m.options.GetCheckAliveInterval()

	if m.options.IsDebugMode() {
		logs := m.options.GetLogger()
		if logs == nil {
			logs = log.NewBasicLogger(helpers.IsProdEnvironment(), true)
			defer func() {
				_ = logs.Sync()
			}()
		}
		logs.Info("Connected to MySQL")
	}
	return nil
}

// GetProviderDB returns the database query provider instance.
// It supports different query providers like SQLC and returns the appropriate type.
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

// IsQueryProviderAvailable checks if a query provider is configured.
// It returns true if a query provider is set in the options.
func (m *MySQLDB[T]) IsQueryProviderAvailable() bool {
	return !helpers.IsEmpty(m.options.GetQueryProvider())
}

// Ping tests the database connection and verifies database existence.
// It performs both connection health check and database availability check.
func (m *MySQLDB[T]) Ping() error {
	if err := m.conn.Ping(); err != nil {
		return err
	}
	if err := m.checkDatabaseExists(context.Background()); err != nil {
		return err
	}
	return nil
}

// checkDatabaseExists verifies that the target database exists in the MySQL instance.
// It extracts the database name from DSN and queries the system catalog.
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

// StartMonitor begins database health monitoring in a separate goroutine.
// It stops any existing monitor before starting a new one to prevent duplicates.
func (m *MySQLDB[T]) StartMonitor() {
	m.monitorMu.Lock()
	defer m.monitorMu.Unlock()

	// If already running, stop the previous one
	if m.monitorRunning {
		m.StopMonitor()
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.monitorCancel = cancel
	m.stopChan = make(chan struct{})

	m.monitorRunning = true
	go database.MonitorDB[database.Database](ctx, m)
}

// StopMonitor gracefully stops the database health monitoring goroutine.
// It cancels the context and closes channels to ensure clean shutdown.
func (m *MySQLDB[T]) StopMonitor() {
	m.monitorMu.Lock()
	defer m.monitorMu.Unlock()

	if m.monitorRunning {
		if m.monitorCancel != nil {
			m.monitorCancel()
		}
		if m.stopChan != nil {
			close(m.stopChan)
		}
		m.monitorRunning = false
	}
}

// GetLogger returns the configured logger instance for the MySQL adapter.
// It provides access to the logger for debugging and monitoring purposes.
func (m *MySQLDB[T]) GetLogger() *log.Log {
	return m.options.GetLogger()
}
