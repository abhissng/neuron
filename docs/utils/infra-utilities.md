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

`utils/cache` provides cache abstractions and implementations (including basic/LRU patterns). Prefer wiring a `CacheManager` into `AppContext`, then creating named caches as needed. Each named cache has its own capacity (for LRU, `MaxSize` is total keys in that cache).

### Standalone CacheManager

```go
cm := cache.NewCacheManagerWithConfig(cache.DefaultLRUConfig())

orders := cm.GetOrCreateCache("orders")
orders.Set("ord_123", map[string]any{"status": "created"})
orders.SetWithExpiry("ord_tmp", map[string]any{"status": "pending"}, 2*time.Minute)

value, ok := orders.Get("ord_123")
_ = value
_ = ok

tokens := cm.GetOrCreateCache("tokens") // separate cache, own MaxSize limit
tokens.SetWithExpiry("t1", "abc", 5*time.Minute)

defer cm.StopAll()
```

### AppContext (recommended)

```go
appCtx := context.NewAppContext(
	context.WithCacheManager(cache.DefaultLRUConfig()),
)

orders := appCtx.GetNamedCache("orders")
orders.Set("ord_123", map[string]any{"status": "created"})

tokens := appCtx.GetNamedCache("tokens")
tokens.SetWithExpiry("t1", "abc", 5*time.Minute)
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

cipher, err := cm.Encrypt([]byte("secret"))
if err != nil {
	return err
}
plain, err := cm.Decrypt(cipher)
if err != nil {
	return err
}
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
