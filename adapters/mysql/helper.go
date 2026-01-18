package mysql

import (
	"context"
	"database/sql"
	"time"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/database"
)

// Close closes the connection to the MySQL database.
func (db *MySQLDB[T]) Close() error {
	err := db.conn.Close()
	if err != nil {
		return err
	}
	close(db.stopChan)
	return nil
}

// FetchStopChannel returns a channel that is closed when the database connection is closed.
func (db *MySQLDB[T]) FetchStopChannel() <-chan struct{} {
	return db.stopChan
}

// FetchCheckAliveInterval returns the interval for checking the health of the database connection.
func (db *MySQLDB[T]) FetchCheckAliveInterval() time.Duration {
	return db.checkAliveInterval
}

// StartMonitorEnabled returns whether health monitoring should start automatically.
func (db *MySQLDB[T]) StartMonitorEnabled() bool {
	return db.options.GetStartMonitor()
}

// Query executes a SQL query and returns a Rows object.
func (db *MySQLDB[T]) Query(ctx context.Context, query string, args ...any) (database.Rows, error) {
	db.IsDebugQuery(query, args...)
	rows, err := db.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &SqlRowsAdapter{rows: rows}, nil
}

// QueryRow executes a SQL query that is expected to return at most one row.
func (db *MySQLDB[T]) QueryRow(ctx context.Context, query string, args ...any) database.Row {
	db.IsDebugQuery(query, args...)
	row := db.conn.QueryRowContext(ctx, query, args...)
	return &SqlRowAdapter{row: row}
}

// Exec executes a SQL statement that does not return rows.
func (db *MySQLDB[T]) Exec(ctx context.Context, query string, args ...any) (database.ExecResult, error) {
	db.IsDebugQuery(query, args...)
	result, err := db.conn.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &SqlExecResult{result: result}, nil
}

// IsDebugQuery logs the query and arguments if debug mode is enabled.
func (db *MySQLDB[T]) IsDebugQuery(query string, args ...any) {
	if db.options.IsDebugMode() {
		logs := db.options.GetLogger()
		logs.Info("Executing Query", log.Any("Query", query))
		logs.Info("Query Args", log.Any("Args", args))
	}
}

// RowsToSlice converts rows to a [][]string representation.
func (db *MySQLDB[T]) RowsToSlice(rows database.Rows) ([][]string, error) {
	return database.RowsToSlice(rows)
}

// MySQLTransaction represents a database transaction within the MySQL context.
type MySQLTransaction struct {
	tx        *sql.Tx
	log       *log.Log
	debugMode bool
}

// BeginTransaction starts a new database transaction.
func (db *MySQLDB[T]) BeginTransaction(ctx context.Context) (database.Transaction, error) {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &MySQLTransaction{tx: tx, log: db.options.GetLogger(), debugMode: db.options.IsDebugMode()}, nil
}

// GetTx returns the transaction object.
func (db *MySQLTransaction) GetTx() any {
	return db.tx
}

// IsDebugQuery logs the query and arguments if debug mode is enabled.
func (db *MySQLTransaction) IsDebugQuery(query string, args ...any) {
	if db.debugMode {
		logs := db.log
		logs.Info("Executing Query", log.Any("Query", query))
		logs.Info("Query Args", log.Any("Args", args))
	}
}

// Query executes a SQL query within the current transaction.
func (t *MySQLTransaction) Query(ctx context.Context, query string, args ...any) (database.Rows, error) {
	t.IsDebugQuery(query, args...)
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &SqlRowsAdapter{rows: rows}, nil
}

// QueryRow executes a SQL query that is expected to return at most one row within the current transaction.
func (t *MySQLTransaction) QueryRow(ctx context.Context, query string, args ...any) database.Row {
	t.IsDebugQuery(query, args...)
	row := t.tx.QueryRowContext(ctx, query, args...)
	return &SqlRowAdapter{row: row}
}

// Exec executes a SQL statement that does not return rows within the current transaction.
func (t *MySQLTransaction) Exec(ctx context.Context, query string, args ...any) (database.ExecResult, error) {
	t.IsDebugQuery(query, args...)
	result, err := t.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &SqlExecResult{result: result}, nil
}

// Commit commits the current transaction.
func (t *MySQLTransaction) Commit(ctx context.Context) error {
	return t.tx.Commit()
}

// Rollback rolls back the current transaction.
func (t *MySQLTransaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback()
}

// RowsToSlice converts rows to a [][]string representation.
func (t *MySQLTransaction) RowsToSlice(rows database.Rows) ([][]string, error) {
	return database.RowsToSlice(rows)
}
