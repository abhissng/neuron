package constant

import (
	"time"

	"github.com/abhissng/neuron/utils/types"
)

// Database constants
const (
	PostgreSQL                 types.DBType  = "postgres"
	MySQL                      types.DBType  = "mysql"
	SQLCProvider               types.DBType  = "sqlc"
	MongoDB                    types.DBType  = "mongo"
	DatabaseCheckAliveInterval time.Duration = 60 * time.Second
)

// Database constant Queries
const (
	DatabaseExistQuery = "SELECT 1 FROM pg_database WHERE datname = $1"
)
