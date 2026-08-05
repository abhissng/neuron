# Blame and Result Usage

Packages covered:

- [`blame`](../../blame)
- [`result`](../../result)

Primary files:

- [`blame/manager.go`](../../blame/manager.go)
- [`blame/error.go`](../../blame/error.go)
- [`blame/general.go`](../../blame/general.go)
- [`blame/general_auth.go`](../../blame/general_auth.go)
- [`blame/general_http.go`](../../blame/general_http.go)
- [`result/result.go`](../../result/result.go)

## Purpose

- `blame` provides typed and categorized error values with metadata.
- `result` provides generic success/failure wrappers to standardize flow outcomes.

These two packages work together in transport layers (HTTP/NATS/gRPC) to map failures to stable response envelopes.

## Blame Pattern

Use `blame` factory functions to create context-rich failures:

```go
if err != nil {
	return blame.DatabaseOperationFailed(err)
}
```

Common error groups:

- request/validation errors
- auth/session/token errors
- transport and external dependency errors
- orchestration/state transition errors

## Result Pattern

`result.Result[T]` can represent:

- success with typed value
- failure with `blame.Blame`

Typical usage:

```go
func LoadOrder(id string) result.Result[Order] {
	order, err := repo.Fetch(id)
	if err != nil {
		return result.NewFailure[Order](blame.InternalServerError(err))
	}
	return result.NewSuccess(order)
}
```

## Integrating in Handlers

In wrapped controller handlers, convert errors to `blame` and map to HTTP status/response payload. This keeps responses consistent across services.

```go
func CreateOrder(sc *context.ServiceContext) result.Result[Order] {
	order, err := repo.Create(sc, payload)
	if err != nil {
		return result.NewFailure[Order](blame.DatabaseOperationFailed(err))
	}
	return result.NewSuccess(order)
}
```

## Internal Helpers (implementation detail)

- `blame` includes initialization and mapping helpers used by startup code and middleware.
- `result` contains helper methods meant for internal chaining patterns.

These helpers are not a stable contract unless exported and intentionally documented.

## Caveats

- Prefer package-provided error constructors over raw `fmt.Errorf` in service boundaries.
- Keep reason codes and response types aligned with `blame/error_definition.json`.
- Do not leak internal error stacks to external clients.
