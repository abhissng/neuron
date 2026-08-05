# Core Context Usage

Packages covered:

- [`context`](../../context)

Primary files:

- [`context/app.go`](../../context/app.go)
- [`context/context.go`](../../context/context.go)
- [`context/service.go`](../../context/service.go)
- [`context/helper.go`](../../context/helper.go)

## Purpose

The context layer provides two levels:

- `AppContext`: long-lived dependency container for adapters/managers.
- `ServiceContext`: per-request/per-message context with request metadata.

## AppContext Bootstrap

Typical bootstrap pattern:

```go
appCtx := context.NewAppContext(
	context.WithServiceID(types.Service("orders")),
	context.WithPostgresDB(pg),
	context.WithNATSManager(natsMgr),
	context.WithLogger(logger),
	context.WithPaymentManager(paymentMgr),
)
```

Common option families in `app.go`:

- identity and service metadata
- error manager (`blame`) and logger
- event manager (`nats`) and HTTP client
- data adapters (`postgres`, `mysql`, `redis`, `mongo`, `opensearch`)
- auth/session/vault/cloud/payment adapters

## ServiceContext Lifecycle

`ServiceContext` is created during request handling and carries:

- correlation/request IDs
- actor and auth metadata
- request payload context
- cancellation/deadline semantics through embedded `context.Context`

In HTTP flows it is usually initialized in Gin middleware and consumed in wrapped handlers.

## Common Usage Patterns

### Accessing shared managers from handler logic

```go
func HandleCreate(sc *context.ServiceContext) error {
	log := sc.FetchLogger()
	db := sc.FetchDatabase()
	_ = log
	_ = db
	return nil
}
```

### Building child contexts

Use standard context derivation to preserve cancellation:

```go
import stdctx "context"

ctx, cancel := stdctx.WithTimeout(sc.Context, 5*time.Second)
defer cancel()
```

## Internal Helpers (implementation detail)

Files in `context/helper.go` and portions of `context/context.go` include internal setters/getters and conversion helpers used by middleware and wrappers. These are not stable external APIs.

## Caveats

- Keep `AppContext` creation centralized in startup code.
- Avoid mutating `ServiceContext` outside boundary middleware/transport wrappers.
- Internal fields and helper methods may change between releases.
