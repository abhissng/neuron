package postgres

import (
	"context"
	"time"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// // Acquire acquires a connection from the connection pool.
// func (db PostgresDB[T]) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
// 	db.monitorMu.Unlock()
// 	return db.pool.Acquire(ctx)
// }

// Stat returns statistics about the connection pool.
func (db *PostgresDB[T]) Stat() *pgxpool.Stat {
	return db.pool.Stat()
}

// Close closes the connection pool.
func (db *PostgresDB[T]) Close() error {
	if db.pool != nil {
		db.pool.Close()
		close(db.stopChan)
	}
	return nil
}

// FetchStopChannel returns a channel that is closed when the database connection is closed.
func (db *PostgresDB[T]) FetchStopChannel() <-chan struct{} {
	return db.stopChan
}

// FetchCheckAliveInterval returns the interval for checking the health of the database connection.
func (db *PostgresDB[T]) FetchCheckAliveInterval() time.Duration {
	return db.checkAliveInterval
}

// Query executes a SQL query and returns a Rows object.
func (db *PostgresDB[T]) Query(ctx context.Context, query string, args ...any) (database.Rows, error) {
	db.IsDebugQuery(query, args...)
	rows, err := db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &PgxRowsAdapter{rows: rows}, nil
}

// QueryRow executes a SQL query that is expected to return at most one row.
func (db *PostgresDB[T]) QueryRow(ctx context.Context, query string, args ...any) database.Row {
	db.IsDebugQuery(query, args...)
	row := db.pool.QueryRow(ctx, query, args...)
	return &PgxRowAdapter{row: row}
}

// Exec executes a SQL statement that does not return rows.
func (db *PostgresDB[T]) Exec(ctx context.Context, query string, args ...any) (database.ExecResult, error) {
	db.IsDebugQuery(query, args...)
	tag, err := db.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &PgxExecResult{commandTag: tag}, nil
}

// IsDebugQuery logs the query and arguments if debug mode is enabled.
func (db *PostgresDB[T]) IsDebugQuery(query string, args ...any) {
	if db.options.IsDebugMode() {
		logs := db.options.GetLogger()
		logs.Info("Executing Query", log.Any("Query", query))
		logs.Info("Query Args", log.Any("Args", args))
	}
}

// RowsToSlice converts rows to a [][]string representation.
func (db *PostgresDB[T]) RowsToSlice(rows database.Rows) ([][]string, error) {
	return database.RowsToSlice(rows)
}

// PostgresTransaction represents a database transaction within the PostgreSQL context.
type PostgresTransaction struct {
	tx        pgx.Tx
	log       *log.Log
	debugMode bool
}

// BeginTransaction starts a new database transaction.
func (db *PostgresDB[T]) BeginTransaction(ctx context.Context) (database.Transaction, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &PostgresTransaction{tx: tx, log: db.options.GetLogger(), debugMode: db.options.IsDebugMode()}, nil
}

// IsDebugQuery logs the query and arguments if debug mode is enabled.
func (db *PostgresTransaction) IsDebugQuery(query string, args ...any) {
	if db.debugMode {
		logs := db.log
		logs.Info("Executing Query", log.Any("Query", query))
		logs.Info("Query Args", log.Any("Args", args))
	}
}

// Query executes a SQL query within the current transaction.
func (t *PostgresTransaction) Query(ctx context.Context, query string, args ...any) (database.Rows, error) {
	t.IsDebugQuery(query, args...)
	rows, err := t.tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &PgxRowsAdapter{rows: rows}, nil
}

// QueryRow executes a SQL query that is expected to return at most one row within the current transaction.
func (t *PostgresTransaction) QueryRow(ctx context.Context, query string, args ...any) database.Row {
	t.IsDebugQuery(query, args...)
	row := t.tx.QueryRow(ctx, query, args...)
	return &PgxRowAdapter{row: row}
}

// Exec executes a SQL statement that does not return rows within the current transaction.
func (t *PostgresTransaction) Exec(ctx context.Context, query string, args ...any) (database.ExecResult, error) {
	t.IsDebugQuery(query, args...)
	tag, err := t.tx.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &PgxExecResult{commandTag: tag}, nil
}

// Commit commits the current transaction.
func (t *PostgresTransaction) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

// Rollback rolls back the current transaction.
func (t *PostgresTransaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

// RowsToSlice converts rows to a [][]string representation.
func (t *PostgresTransaction) RowsToSlice(rows database.Rows) ([][]string, error) {
	return database.RowsToSlice(rows)
}
