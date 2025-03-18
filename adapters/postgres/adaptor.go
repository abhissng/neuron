package postgres

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// PgxRowsAdapter adapts pgx.Rows to database.Rows interface.
type PgxRowsAdapter struct {
	rows pgx.Rows
}

// Close closes the underlying pgx.Rows.
func (a *PgxRowsAdapter) Close() error {
	a.rows.Close()
	return nil
}

// Next calls Next on the underlying pgx.Rows.
func (a *PgxRowsAdapter) Next() bool {
	return a.rows.Next()
}

// Scan calls Scan on the underlying pgx.Rows.
func (a *PgxRowsAdapter) Scan(dest ...any) error {
	return a.rows.Scan(dest...)
}

// Err returns the error from the underlying pgx.Rows.
func (a *PgxRowsAdapter) Err() error {
	return a.rows.Err()
}
func (a *PgxRowsAdapter) Columns() ([]string, error) {
	fields := a.rows.FieldDescriptions()
	columns := make([]string, len(fields))
	for i, field := range fields {
		columns[i] = string(field.Name)
	}
	return columns, nil
}

// PgxRowAdapter adapts pgx.Row to database.Row interface.
type PgxRowAdapter struct {
	row pgx.Row
}

// Scan calls Scan on the underlying pgx.Row.
func (a *PgxRowAdapter) Scan(dest ...any) error {
	return a.row.Scan(dest...)
}

// PgxExecResult adapts pgconn.CommandTag to database.ExecResult interface.
type PgxExecResult struct {
	commandTag pgconn.CommandTag
}

// RowsAffected returns the number of rows affected from the underlying pgconn.CommandTag.
func (r *PgxExecResult) RowsAffected() int64 {
	return r.commandTag.RowsAffected()
}

// LastInsertId returns 0 as PostgreSQL does not support retrieving the last inserted ID.
func (r *PgxExecResult) LastInsertId() int64 {
	// PostgreSQL does not support LastInsertId, so return 0
	return 0
}
