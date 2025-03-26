package database

import (
	"fmt"
	"reflect"
	"regexp"
	"time"

	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/types"
)

// BuildDSN constructs a DSN (Data Source Name) string based on input parameters.
// host should be in the format "host:port" or the domain if it
func BuildDSN(
	dbType types.DBType,
	host string,
	dbName,
	user,
	password string,
	options map[string]string) string {

	// Based on DBType DSN will be build
	switch dbType {
	case constant.PostgreSQL:
		// For PostgreSQL: format is "postgresql://user:password@host:port/dbname?options"
		dsn := fmt.Sprintf("postgresql://%s:%s@%s/%s", user, password, host, dbName)
		// dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user, password, host, dbName)
		if len(options) > 0 {
			dsn += "?"
			for key, value := range options {
				dsn += fmt.Sprintf("%s=%s&", key, value)
			}
			dsn = dsn[:len(dsn)-1] // Remove the trailing '&'
		}
		return dsn

	case constant.MySQL:
		// For MySQL: format is "user:password@tcp(host:port)/dbname?options"
		// dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, dbName)
		dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, password, host, dbName)
		if len(options) > 0 {
			dsn += "?"
			for key, value := range options {
				dsn += fmt.Sprintf("%s=%s&", key, value)
			}
			dsn = dsn[:len(dsn)-1] // Remove the trailing '&'
		}
		return dsn

	default:
		// Unsupported database type
		return ""
	}
}

// ExtractDBNameFromDSN extracts the database name from a DSN string.
func ExtractDBNameFromDSN(dbType types.DBType, dsn string) (string, error) {
	regxStr := ""
	switch dbType {
	case constant.PostgreSQL:
		regxStr = constant.PostgresDSNRegex
	case constant.MySQL:
		regxStr = constant.MysqlDSNRegex
	default:
		return "", fmt.Errorf("unable to extract database name from DSN: %s", dsn)
	}
	dsnRegex := regexp.MustCompile(regxStr)
	if matches := dsnRegex.FindStringSubmatch(dsn); len(matches) > 1 {
		return matches[1], nil
	}
	// If no match is found, return an error
	return "", fmt.Errorf("unable to extract database name from DSN: %s", dsn)
}

// RowsToSlice converts database.Rows object into a 2D slice of strings.
//
// Each row in the database.Rows object is converted into a slice of strings,
// where each string represents a column value in that row.
//
// Returns a 2D slice of strings representing the data from the database.Rows object,
// and an error if any occurs during the conversion process.
func RowsToSlice(rows Rows) ([][]string, error) {
	// 1. Get column names from the result set
	columns, err := getColumns(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// 2. Initialize results slice with column headers as the first row
	results := [][]string{columns}

	// 3. Iterate over each row in the result set
	for rows.Next() {
		// 4. Scan the current row's values
		values, err := scanRow(rows, len(columns))
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		// 5. Convert scanned values to strings
		row := convertValuesToStrings(values)

		// 6. Append the converted row to the results slice
		results = append(results, row)
	}

	// 7. Check for errors during row iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	// 8. Return the slice containing all rows as strings
	return results, nil
}

// Helper function to check if the rows implement the ExtendedRows interface
// This interface is required to access column names from the result set
func getColumns(rows Rows) ([]string, error) {
	extRows, ok := rows.(ExtendedRows)
	if !ok {
		return nil, fmt.Errorf("rows does not implement ExtendedRows")
	}

	// Get the column names using the ExtendedRows interface
	columns, err := extRows.Columns()
	if err != nil {
		return nil, err
	}
	return columns, nil
}

// Helper function to scan the values of the current row
func scanRow(rows Rows, columnCount int) ([]any, error) {
	// Create a slice to hold the scanned values
	values := make([]any, columnCount)

	// Prepare a slice of pointers to the values slice
	// This is necessary because rows.Scan expects pointers as arguments
	for i := range values {
		var val any
		values[i] = &val
	}

	// Scan the row's values into the slice of pointers
	if err := rows.Scan(values...); err != nil {
		return nil, err
	}

	// Return the slice containing the scanned values
	return values, nil
}

// Helper function to convert scanned values (of any type) to strings
func convertValuesToStrings(values []any) []string {
	// Create a slice to hold the converted string values
	row := make([]string, len(values))

	// Iterate over each value and convert it to a string based on its type
	for i, value := range values {
		row[i] = convertToString(reflect.Indirect(reflect.ValueOf(value)).Interface())
	}

	// Return the slice containing the converted string values
	return row
}

// Helper function to handle type conversion based on the value's type
func convertToString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v // Already a string, return directly
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v) // Format integers as decimals
	case float32, float64:
		return fmt.Sprintf("%f", v) // Format floats
	case time.Time:
		return v.Format(time.RFC3339) // Format time as RFC3339
	case fmt.Stringer:
		return v.String() // Use the String() method if the value implements it
	default:
		return fmt.Sprintf("%v", v) // Default formatting for unknown types
	}
}
