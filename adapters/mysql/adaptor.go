package mysql

import (
	"database/sql"
)

// SqlRowsAdapter adapts sql.Rows to database.Rows interface.
type SqlRowsAdapter struct {
	rows *sql.Rows
}

// Close closes the underlying sql.Rows.
func (a *SqlRowsAdapter) Close() error {
	return a.rows.Close()
}

// Next calls Next on the underlying sql.Rows.
func (a *SqlRowsAdapter) Next() bool {
	return a.rows.Next()
}

// Scan calls Scan on the underlying sql.Rows.
func (a *SqlRowsAdapter) Scan(dest ...any) error {
	return a.rows.Scan(dest...)
}

// Err returns the error from the underlying sql.Rows.
func (a *SqlRowsAdapter) Err() error {
	return a.rows.Err()
}

// SqlRowAdapter adapts sql.Row to database.Row interface.
type SqlRowAdapter struct {
	row *sql.Row
}

// Scan calls Scan on the underlying sql.Row.
func (a *SqlRowAdapter) Scan(dest ...any) error {
	return a.row.Scan(dest...)
}

// SqlExecResult adapts sql.Result to database.ExecResult interface.
type SqlExecResult struct {
	result sql.Result
}

// RowsAffected returns the number of rows affected by the executed statement.
func (r *SqlExecResult) RowsAffected() int64 {
	rowsAffected, err := r.result.RowsAffected()
	if err != nil {
		return 0
	}
	return rowsAffected
}

// LastInsertId returns the ID of the last inserted row, if applicable.
func (r *SqlExecResult) LastInsertId() int64 {
	lastInsertId, err := r.result.LastInsertId()
	if err != nil {
		return 0
	}
	return lastInsertId
}

// Columns returns the column names from the result set.
func (mr *SqlRowsAdapter) Columns() ([]string, error) {
	return mr.rows.Columns()
}
