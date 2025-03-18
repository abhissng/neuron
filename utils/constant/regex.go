package constant

const (
	// Regular expression for PostgreSQL DSN: "postgresql://user:password@host:port/dbname?options"
	PostgresDSNRegex = `^postgresql://[^/]+/([^?]+)`

	// Regular expression for MySQL DSN: "user:password@tcp(host:port)/dbname?options"
	MysqlDSNRegex = `^[^/]+/\([^?]+`
)
