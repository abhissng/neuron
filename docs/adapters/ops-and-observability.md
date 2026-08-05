# Ops and Observability Adapters Usage

Packages covered:

- [`adapters/log`](../../adapters/log)
- [`adapters/prometheus`](../../adapters/prometheus)
- [`adapters/validator`](../../adapters/validator)
- [`adapters/viper`](../../adapters/viper)

## Logging

`adapters/log` provides structured logging utilities and wrapper methods used across HTTP, engine, and adapter layers.

Recommended pattern:

- initialize logger once at startup
- inject through `AppContext`
- use structured key/value fields in all boundary layers

```go
logger := log.NewBasicLogger(false, true)
logger.Info("service started", log.String("service", "orders"))
logger.Error("dependency failed", log.Any("error", err))
```

## Prometheus

`adapters/prometheus` exposes metrics wiring patterns and optionized setup for instrumentation in service runtime paths.

```go
mc := prometheus.NewMetricsCollector(
	prometheus.WithServiceName("orders-service"),
)
_ = mc
```

## Validator

`adapters/validator` provides validation helpers for request/data checks and complements request parsing wrappers.

## Viper

`adapters/viper` provides configuration load/default helper patterns and environment-aware settings usage.

```go
cfg := viper.NewViper("config", "yaml", ".")
_ = cfg
```

## Internal Helpers (implementation detail)

Formatting defaults, helper wrappers, and adapter-specific bootstrap routines may change without compatibility guarantees.

## Caveats

- Keep logs structured and redact sensitive fields.
- Ensure metrics labels remain low cardinality.
- Validate configuration fail-fast during startup.
