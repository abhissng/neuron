package postgres

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/database"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgreSQL-specific implementation
type PostgresDB[T any] struct {
	pool               *pgxpool.Pool
	options            *database.PostgresDBOptions
	db                 T
	factory            database.PostgresQueriesFactory[T]
	stopChan           chan struct{}
	checkAliveInterval time.Duration
	monitorCancel      context.CancelFunc
	monitorRunning     bool
	monitorMu          sync.Mutex
}

// NewPostgresFactory creates a new PostgresDB with a custom factory.
func NewPostgresFactory[T any](options *database.PostgresDBOptions, factory database.GenericQueriesFactory[T]) *PostgresDB[T] {
	return &PostgresDB[T]{
		options: options,
		factory: func(db database.PostgresDBTX) T {
			return factory(db)
		},
		stopChan: make(chan struct{}),
	}
}

// Connect establishes a connection and initializes the Queries struct using the factory.
func (p *PostgresDB[T]) Connect(ctx context.Context) error {
	cfg, err := pgxpool.ParseConfig(p.options.GetDSN())
	if err != nil {
		return fmt.Errorf("failed to parse DSN: %w", err)
	}

	cfg.MaxConns = int32(p.options.GetMaxConns()) // #nosec G115

	p.pool, err = pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	if err := p.Ping(); err != nil {
		return err
	}

	p.checkAliveInterval = p.options.GetCheckAliveInterval()

	// Use the factory to initialize the Queries struct.
	if !helpers.IsEmpty(p.options.GetQueryProvider()) {
		p.applyQueryProvider()
	}

	if p.options.IsDebugMode() {
		logs := p.options.GetLogger()
		if logs == nil {
			logs = log.NewBasicLogger(helpers.IsProdEnvironment(), true)
		}
		logs.Info("Connected to PostgreSQL")
		_ = logs.Sync()
	}
	return nil
}

// applyQueryProvider applies the query provider to the database.
func (p *PostgresDB[T]) applyQueryProvider() {
	switch p.options.GetQueryProvider() {
	case constant.SQLCProvider.String():
		{
			p.db = p.factory(p.pool)
		}
	}
}

// GetProviderDB returns the initialized query struct.
func (p *PostgresDB[T]) GetProviderDB() any {
	switch p.options.GetQueryProvider() {
	case constant.SQLCProvider.String():
		{
			return p.db
		}
	default:
		var defaultValue T // This will give you the zero value of T.
		return defaultValue
	}
}

// IsQueryProviderAvailable returns the if queryProvider is provided or not
func (p *PostgresDB[T]) IsQueryProviderAvailable() bool {
	return !helpers.IsEmpty(p.options.GetQueryProvider())
}

// Ping checks the connection to the database.
func (p *PostgresDB[T]) Ping() error {
	if err := p.pool.Ping(context.Background()); err != nil {
		return err
	}

	if err := p.checkDatabaseExists(context.Background()); err != nil {
		return err
	}
	return nil
}

// checkDatabaseExists verifies that the specified database exists.
func (p *PostgresDB[T]) checkDatabaseExists(ctx context.Context) error {
	// Check if the database exists
	dbName, err := database.ExtractDBNameFromDSN(constant.PostgreSQL, p.options.GetDSN()) // Extract the database name from the DSN
	if err != nil {
		return fmt.Errorf("database name cannot be blank: %w", err)
	}

	// Query the PostgreSQL system catalog to check for the database
	var exists int32
	err = p.pool.QueryRow(ctx, constant.DatabaseExistQuery, dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("database '%s' does not exist", dbName)
	}
	return nil
}

func (p *PostgresDB[T]) StartMonitor() {
	p.monitorMu.Lock()
	defer p.monitorMu.Unlock()

	// If already running, stop the previous one
	if p.monitorRunning {
		p.StopMonitor()
	}

	ctx, cancel := context.WithCancel(context.Background())
	p.monitorCancel = cancel
	p.stopChan = make(chan struct{})

	p.monitorRunning = true
	go database.MonitorDB[database.Database](ctx, p)
}

func (p *PostgresDB[T]) StopMonitor() {
	p.monitorMu.Lock()
	defer p.monitorMu.Unlock()

	if p.monitorRunning {
		if p.monitorCancel != nil {
			p.monitorCancel()
		}
		if p.stopChan != nil {
			close(p.stopChan)
		}
		p.monitorRunning = false
	}
}

func (p *PostgresDB[T]) GetLogger() *log.Log {
	return p.options.GetLogger()
}
