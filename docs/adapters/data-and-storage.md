# Data and Storage Adapters Usage

Packages covered:

- [`adapters/postgres`](../../adapters/postgres)
- [`adapters/mysql`](../../adapters/mysql)
- [`adapters/mongo`](../../adapters/mongo)
- [`adapters/redis`](../../adapters/redis)
- [`adapters/opensearch`](../../adapters/opensearch)
- [`adapters/cosmos`](../../adapters/cosmos)
- [`adapters/store`](../../adapters/store)
- [`adapters/store/regex`](../../adapters/store/regex)
- [`database`](../../database)

## SQL Backends

Postgres and MySQL adapters follow similar lifecycle/query patterns:

- initialize connection pool
- expose query/exec methods
- provide stop-channel and health helpers

```go
pgOpts := database.NewDBOptions[*database.PostgresDBOptions](
	database.WithDSN("postgresql://user:pass@localhost:5432/orders"),
	database.WithMaxConns(20),
)
pgDB, err := database.NewDatabase(
	postgres.NewPostgresFactory[any],
	nil,
	pgOpts,
)
_ = pgDB
_ = err
```

Use core database abstractions described in [`../core/database.md`](../core/database.md).

## Mongo Adapter

`adapters/mongo` provides document-store manager capabilities for services needing collection-oriented persistence.

## Redis Adapter

`adapters/redis` provides cache/session/state operations and is often used by middleware/session layers.

```go
redisMgr, err := redis.NewRedisManager(&redis.Config{
	Addr: "localhost:6379",
})
_ = redisMgr
_ = err
```

## OpenSearch Adapter

`adapters/opensearch` provides search client setup and helper methods for index/query usage.

## Cosmos Package

`adapters/cosmos` currently contains draft/disabled content and is not a full production-ready adapter.

## Store and Regex Helpers

- `adapters/store` provides higher-level store manager helpers.
- `adapters/store/regex` provides regex-based matching helpers tied to store patterns.

## Internal Helpers (implementation detail)

Adapter helper files include conversion, defaulting, and lifecycle glue methods that are not stable external contracts.

## Caveats

- Keep adapter init at startup, not per request.
- Separate SQL and non-SQL concerns in service logic.
- Validate indexing and query assumptions outside generic wrappers.
