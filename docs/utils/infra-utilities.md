# Infrastructure Utilities Usage

Packages covered:

- [`utils/cache`](../../utils/cache)
- [`utils/circuitBreaker`](../../utils/circuitBreaker)
- [`utils/cryptography`](../../utils/cryptography)
- [`utils/graceful`](../../utils/graceful)
- [`utils/idempotency`](../../utils/idempotency)
- [`utils/schedule`](../../utils/schedule)
- [`utils/timeutil`](../../utils/timeutil)
- [`utils/workerpool`](../../utils/workerpool)
- [`utils/concurrent/concurrentMap`](../../utils/concurrent/concurrentMap)

## Cache

`utils/cache` provides cache abstractions and implementations (including basic/LRU patterns) used by app-level managers and request paths.

```go
cm := cache.NewCacheManager()
c := cm.GetOrCreateCache("orders")

c.Set("ord_123", map[string]any{"status": "created"})
value, ok := c.Get("ord_123")
_ = value
_ = ok
```

## Circuit Breaker

`utils/circuitBreaker` provides resilience primitives for external dependency calls.

```go
cb := circuitBreaker.NewCircuitBreaker(
	circuitBreaker.WithName("orders-http"),
)
_ = cb
```

## Cryptography

`utils/cryptography` provides key and crypto helpers:

- RSA/ed25519 related helpers
- manager-style utility wrappers
- encoding/signing helper paths

```go
cm, err := cryptography.NewCryptoManager()
if err != nil {
	return err
}

cipher, _ := cm.Encrypt([]byte("secret"))
plain, _ := cm.Decrypt(cipher)
_ = plain
```

## Graceful and Scheduling

- `utils/graceful` provides shutdown/lifecycle helpers.
- `utils/schedule` provides scheduling configuration/runner helpers.
- `utils/timeutil` provides time parsing/formatting/conversion helpers.

```go
tw := timeutil.Now().AddDate(0, 0, 1)
formatted := tw.Format(time.RFC3339)
_ = formatted

parsed, _ := timeutil.ParseTime("2026-08-05T12:00:00Z")
_ = parsed
```

## Worker Pool and Concurrency

- `utils/workerpool` provides worker pool primitives for bounded async work.
- `utils/concurrent/concurrentMap` provides concurrency-safe map helper.

```go
wp := workerpool.NewWorkerPool[string, string](processor,
	workerpool.WithNumWorkers[string, string](4),
)
wp.Submit(result.Task[string]{Input: "task-1"})
defer wp.Shutdown()

m := concurrentMap.NewConcurrentMap[string, int]()
m.Set("active", 1)
count, _ := m.Get("active")
_ = count
```

## Idempotency

`utils/idempotency` provides helper primitives for idempotency controls in request/event handling flows.

## Internal Helpers (implementation detail)

Many utility internals are optimized for current package consumers and may change without API guarantees unless exported.

## Caveats

- Treat infra utilities as building blocks; enforce domain invariants in calling services.
- Apply cancellation/timeouts explicitly in worker and scheduling workflows.
- Validate crypto key/config assumptions in startup checks.
