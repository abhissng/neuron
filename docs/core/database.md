# Database Usage

Packages covered:

- [`database`](../../database)
- [`adapters/postgres`](../../adapters/postgres)
- [`adapters/mysql`](../../adapters/mysql)
- [`adapters/mongo`](../../adapters/mongo)
- [`adapters/redis`](../../adapters/redis)
- [`adapters/opensearch`](../../adapters/opensearch)

Primary files:

- [`database/database.go`](../../database/database.go)
- [`database/postgres.go`](../../database/postgres.go)
- [`database/mysql.go`](../../database/mysql.go)
- [`adapters/postgres/postgres.go`](../../adapters/postgres/postgres.go)
- [`adapters/postgres/adaptor.go`](../../adapters/postgres/adaptor.go)
- [`adapters/mysql/mysql.go`](../../adapters/mysql/mysql.go)

## Purpose

The database layer defines common interfaces and adapter wrappers for SQL and non-SQL stores.

## SQL Adapters

### Postgres

Use Postgres manager/adaptor for:

- pooled connection lifecycle
- typed query execution patterns
- health-check and graceful stop channel hooks

### MySQL

MySQL adapter mirrors most Postgres helper/adaptor patterns for consistent SQL backend usage.

## Generic Database Interface

`database.Database` is the shared contract used by context-level dependency wiring:

- connect and close
- query/exec operations
- transaction-oriented methods (where applicable)
- health and lifecycle controls

## Typical Wiring Pattern

```go
pgOpts := database.NewDBOptions[*database.PostgresDBOptions](
	database.WithDSN("postgresql://user:pass@localhost:5432/app"),
	database.WithMaxConns(20),
)

db, err := database.NewDatabase(
	postgres.NewPostgresFactory[any],
	nil, // query factory if using sqlc-style wrapper
	pgOpts,
)
if err != nil {
	return err
}

appCtx := context.NewAppContext(
	context.WithPostgresDB(db),
)
```

Downstream services can fetch DB from `ServiceContext`/`AppContext` rather than constructing their own clients.

## Non-SQL Stores

- Mongo (`adapters/mongo`) for document storage
- Redis (`adapters/redis`) for cache/session/state
- OpenSearch (`adapters/opensearch`) for search indexing/query

These are manager-style adapters and may not fully conform to SQL-oriented database interfaces.

## Internal Helpers (implementation detail)

Files such as `adapters/postgres/helper.go` and `adapters/mysql/helper.go` include internal lifecycle/query helpers and compatibility wrappers.

## Caveats

- Treat SQL and non-SQL adapters as separate contracts even when both live in `AppContext`.
- Align timeout/retry settings with service SLOs.
- Keep migration/query-layer assumptions outside generic wrappers.
