# HTTP and Gin Adapters Usage

Packages covered:

- [`adapters/gin/server`](../../adapters/gin/server)
- [`adapters/gin/middleware`](../../adapters/gin/middleware)
- [`adapters/gin/middleware/redis_rate_limiter`](../../adapters/gin/middleware/redis_rate_limiter)
- [`adapters/gin/handler`](../../adapters/gin/handler)
- [`adapters/gin/request`](../../adapters/gin/request)
- [`adapters/http`](../../adapters/http)
- [`adapters/grpcserver`](../../adapters/grpcserver)
- [`adapters/grpcserver/example`](../../adapters/grpcserver/example)

Primary files:

- [`adapters/gin/server/server.go`](../../adapters/gin/server/server.go)
- [`adapters/gin/middleware/middleware.go`](../../adapters/gin/middleware/middleware.go)
- [`adapters/gin/handler/handler.go`](../../adapters/gin/handler/handler.go)
- [`adapters/gin/request/request.go`](../../adapters/gin/request/request.go)
- [`adapters/http/wrapper.go`](../../adapters/http/wrapper.go)
- [`adapters/grpcserver/grpcserver.go`](../../adapters/grpcserver/grpcserver.go)

## Gin Server Lifecycle

`adapters/gin/server` provides server construction and startup options:

- host/port and timeout settings
- middleware registration
- route grouping and handler attachment

Typical flow:

1. Build options.
2. Attach middleware.
3. Register handlers.
4. Start listener.

```go
err := server.StartServer(
	server.WithPort("8080"),
	server.WithLogger(logger),
	server.WithGlobalMiddleware(middleware.RequestIDMiddleware()),
)
if err != nil {
	return err
}
```

## Middleware Stack

`adapters/gin/middleware` includes:

- request ID and correlation propagation
- service context binding
- auth/session gates
- CORS/CSRF/IP limiter support
- metrics/logging wrappers

`redis_rate_limiter` is an optional Redis-backed limiter integration package.

## Handler Wrapper Pattern

`adapters/gin/handler` provides standardized controller wrappers that:

- recover and normalize panics
- map `blame` errors to API responses
- keep response envelopes consistent

## Request Utilities

`adapters/gin/request` centralizes parsing and validation for:

- path/query/header values
- typed conversions and required checks
- consistent error construction

## HTTP Client Adapter

`adapters/http` wraps outbound HTTP behavior:

- client configuration (timeouts, retries, headers)
- request/response helper routines
- middleware-compatible instrumentation points

```go
cfg := http.NewHttpClientManager(
	"https://api.example.com/orders",
	http.WithMethod("GET"),
	http.WithLogger(logger),
)

res := http.DoRequest[OrderResponse](nil, cfg)
_ = res
```

## gRPC Server Adapter

`adapters/grpcserver` provides server setup and interceptors for auth/validation/logging concerns.

The `adapters/grpcserver/example` package contains example setup snippets; treat it as reference code.

## Internal Helpers (implementation detail)

- `helpers.go` and wrapper internals in these packages are transport plumbing and may change.
- request parser internals are not intended as stable cross-package APIs unless exported.

## Caveats

- Keep business logic outside middleware and wrappers.
- Enforce a single response envelope strategy through handler wrappers.
- Prefer typed request helpers over ad-hoc parsing in handlers.
